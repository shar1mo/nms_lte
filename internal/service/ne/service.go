package ne

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
)

var ErrNENotFound = errors.New("ne doesn't exist")

type Store interface {
	SaveNE(ne model.NetworkElement) error
	GetNE(id string) (model.NetworkElement, bool, error)
	DeleteNE(id string) error
	ListNE() ([]model.NetworkElement, error)
}

type managedConn struct {
	client netconf.RPCClient
	mu     sync.Mutex
}

type Service struct {
	store             Store
	connector         Connector
	reconnectInterval time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	mu                sync.Mutex
	watchers          map[string]struct{}
	destroyRuntime    bool
	conns             map[string]*managedConn
}

func NewService(store Store, opts ...Option) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{
		store:             store,
		reconnectInterval: defaultNetconfReconnectInterval,
		ctx:               ctx,
		cancel:            cancel,
		watchers:          make(map[string]struct{}),
		conns:             make(map[string]*managedConn),
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (s *Service) Register(name, address, vendor string, capabilities []string) (model.NetworkElement, error) {
	if strings.TrimSpace(name) == "" {
		return model.NetworkElement{}, errors.New("name is required")
	}
	if strings.TrimSpace(address) == "" {
		return model.NetworkElement{}, errors.New("address is required")
	}
	if len(capabilities) == 0 && s.connector == nil {
		capabilities = []string{
			"urn:ietf:params:netconf:base:1.0",
			"urn:ietf:params:netconf:base:1.1",
		}
	}

	now := time.Now().UTC()
	status := "active"
	if s.connector != nil {
		status = "connecting"
	}

	ne := model.NetworkElement{
		ID:           id.New("ne"),
		Name:         strings.TrimSpace(name),
		Address:      strings.TrimSpace(address),
		Vendor:       strings.TrimSpace(vendor),
		Status:       status,
		Capabilities: append([]string(nil), capabilities...),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.store.SaveNE(ne); err != nil {
		return model.NetworkElement{}, err
	}

	if s.connector != nil {
		s.startWatcher(ne.ID)
	}

	return ne, nil
}

func (s *Service) UnRegister(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}

	ne, ok, err := s.store.GetNE(id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNENotFound
	}

	if ne.Status == "active" {
		return errors.New("status is active, deactivate ne first")
	}

	if err := s.store.DeleteNE(id); err != nil {
		return errors.New("failed to delete ne")
	}

	return nil
}

func (s *Service) List() ([]model.NetworkElement, error) {
	return s.store.ListNE()
}

func (s *Service) Get(neID string) (model.NetworkElement, bool, error) {
	return s.store.GetNE(neID)
}

func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.closeManagedConns()
	if s.destroyRuntime {
		netconf.Destroy()
	}
}

func (s *Service) startWatcher(neID string) {
	s.mu.Lock()
	if _, exists := s.watchers[neID]; exists {
		s.mu.Unlock()
		return
	}
	s.watchers[neID] = struct{}{}
	s.wg.Add(1)
	s.mu.Unlock()

	go func() {
		defer s.wg.Done()
		defer s.removeWatcher(neID)
		s.connectionLoop(neID)
	}()
}

func (s *Service) removeWatcher(neID string) {
	s.mu.Lock()
	delete(s.watchers, neID)
	s.mu.Unlock()
}

func (s *Service) connectionLoop(neID string) {
	for {
		if err := s.ctx.Err(); err != nil {
			return
		}

		ne, ok, err := s.store.GetNE(neID)
		if err != nil || !ok {
			return
		}

		session, err := s.connector.Connect(ne.Address)
		if err != nil {
			s.clearManagedConn(neID, nil)
			if err := s.updateConnectionState(neID, "disconnected", nil); err != nil {
				return
			}

			if !s.waitReconnect() {
				return
			}
			continue
		}

		capabilities, err := session.Capabilities()
		if err != nil {
			session.Close()
			s.clearManagedConn(neID, nil)
			if err := s.updateConnectionState(neID, "disconnected", nil); err != nil {
				return
			}

			if !s.waitReconnect() {
				return
			}
			continue
		}

		conn := &managedConn{client: session}
		s.setManagedConn(neID, conn)
		if err := s.updateConnectionState(neID, "connected", capabilities); err != nil {
			s.releaseManagedConn(neID, conn)
			return
		}

		if !s.monitorSession(neID, conn) {
			return
		}
		if err := s.updateConnectionState(neID, "reconnecting", nil); err != nil {
			return
		}
	}
}

func (s *Service) monitorSession(neID string, conn *managedConn) bool {
	ticker := time.NewTicker(s.reconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.releaseManagedConn(neID, conn)
			return false
		case <-ticker.C:
			conn.mu.Lock()
			client := conn.client
			alive := client != nil && client.IsAlive()
			conn.mu.Unlock()
			if alive {
				continue
			}
			s.releaseManagedConn(neID, conn)
			return true
		}
	}
}

func (s *Service) waitReconnect() bool {
	timer := time.NewTimer(s.reconnectInterval)
	defer timer.Stop()

	select {
	case <-s.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (s *Service) updateConnectionState(neID, status string, capabilities []string) error {
	ne, ok, err := s.store.GetNE(neID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNENotFound
	}

	ne.Status = status
	if capabilities != nil {
		ne.Capabilities = append([]string(nil), capabilities...)
	}
	ne.UpdatedAt = time.Now().UTC()

	if err := s.store.SaveNE(ne); err != nil {
		return err
	}

	return nil
}

func (s *Service) setManagedConn(neID string, conn *managedConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[neID] = conn
}

func (s *Service) clearManagedConn(neID string, expected *managedConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.conns[neID]
	if !ok {
		return
	}
	if expected != nil && current != expected {
		return
	}
	delete(s.conns, neID)
}

func (s *Service) releaseManagedConn(neID string, conn *managedConn) {
	if conn == nil {
		s.clearManagedConn(neID, nil)
		return
	}

	conn.mu.Lock()
	client := conn.client
	conn.client = nil
	conn.mu.Unlock()

	s.clearManagedConn(neID, conn)

	if client != nil {
		client.Close()
	}
}

func (s *Service) closeManagedConns() {
	s.mu.Lock()
	if len(s.conns) == 0 {
		s.mu.Unlock()
		return
	}

	pending := make([]*managedConn, 0, len(s.conns))
	for neID, conn := range s.conns {
		delete(s.conns, neID)
		pending = append(pending, conn)
	}
	s.mu.Unlock()

	for _, conn := range pending {
		if conn == nil {
			continue
		}
		conn.mu.Lock()
		client := conn.client
		conn.client = nil
		conn.mu.Unlock()
		if client != nil {
			client.Close()
		}
	}
}