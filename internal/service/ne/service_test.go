package ne

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
	"nms_lte/internal/store/memory"
)

type fakeConnector struct {
	mu      sync.Mutex
	results []fakeConnectResult
	calls   int
}

type fakeConnectResult struct {
	session *fakeSession
	err     error
	wait    <-chan struct{}
}

type fakeEditCall struct {
	datastore string
	payload   string
}

type fakeSession struct {
	mu           sync.Mutex
	capabilities []string
	alive        bool
	operations   []string
	edits        []fakeEditCall
	lockErr      error
	editErr      error
	validateErr  error
	commitErr    error
	discardErr   error
	unlockErr    error
}

func (c *fakeConnector) Connect(string) (Session, error) {
	c.mu.Lock()
	c.calls++
	if len(c.results) == 0 {
		c.mu.Unlock()
		return nil, errors.New("unexpected connect call")
	}

	result := c.results[0]
	c.results = c.results[1:]
	c.mu.Unlock()

	if result.wait != nil {
		<-result.wait
	}

	if result.err != nil {
		return nil, result.err
	}

	return result.session, nil
}

func (c *fakeConnector) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (s *fakeSession) Capabilities() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.capabilities...), nil
}

func (s *fakeSession) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alive
}

func (s *fakeSession) Close() {
	s.mu.Lock()
	s.alive = false
	s.operations = append(s.operations, "close")
	s.mu.Unlock()
}

func (s *fakeSession) setAlive(value bool) {
	s.mu.Lock()
	s.alive = value
	s.mu.Unlock()
}

func (s *fakeSession) Get(string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) GetConfig(string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Edit(datastore, editContent string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "edit")
	s.edits = append(s.edits, fakeEditCall{datastore: datastore, payload: editContent})
	if s.editErr != nil {
		return nil, s.editErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) Copy(string, string, string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Delete(string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Lock(datastore string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "lock")
	if datastore != netconf.DatastoreCandidate {
		return nil, fmt.Errorf("unexpected datastore %q", datastore)
	}
	if s.lockErr != nil {
		return nil, s.lockErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) Unlock(datastore string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "unlock")
	if datastore != netconf.DatastoreCandidate {
		return nil, fmt.Errorf("unexpected datastore %q", datastore)
	}
	if s.unlockErr != nil {
		return nil, s.unlockErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) Commit() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "commit")
	if s.commitErr != nil {
		return nil, s.commitErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) Discard() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "discard")
	if s.discardErr != nil {
		return nil, s.discardErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) Cancel(string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Validate(source, urlOrConfig string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operations = append(s.operations, "validate")
	if source != netconf.DatastoreCandidate {
		return nil, fmt.Errorf("unexpected validate source %q", source)
	}
	if urlOrConfig != "" {
		return nil, fmt.Errorf("unexpected validate payload %q", urlOrConfig)
	}
	if s.validateErr != nil {
		return nil, s.validateErr
	}
	return []byte("<ok/>"), nil
}

func (s *fakeSession) GetSchema(string, string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Subscribe(string, string, string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) GetData(string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) EditData(string, string) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) Kill(uint32) ([]byte, error) {
	return nil, nil
}

func (s *fakeSession) snapshotOperations() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.operations...)
}

func (s *fakeSession) snapshotEdits() []fakeEditCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]fakeEditCall(nil), s.edits...)
}

func TestRegisterKeepsDisconnectedNodeAndReconnects(t *testing.T) {
	store := memory.New()
	connector := &fakeConnector{
		results: []fakeConnectResult{
			{err: errors.New("dial failed")},
			{session: &fakeSession{capabilities: []string{"cap-1"}, alive: true}},
		},
	}

	service := NewService(
		store,
		WithConnector(connector),
		WithReconnectInterval(10*time.Millisecond),
	)
	defer service.Close()

	neItem, err := service.Register("enb-1", "127.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	waitForConnectCalls(t, connector, 2)
	waitForNEStatus(t, service, neItem.ID, "connected")

	updated, ok := service.Get(neItem.ID)
	if !ok {
		t.Fatal("expected network element")
	}
	if len(updated.Capabilities) != 1 || updated.Capabilities[0] != "cap-1" {
		t.Fatalf("unexpected capabilities: %#v", updated.Capabilities)
	}
}

func TestRegisterReconnectsAfterSessionDrop(t *testing.T) {
	store := memory.New()
	reconnectGate := make(chan struct{})
	firstSession := &fakeSession{capabilities: []string{"cap-1"}, alive: true}
	secondSession := &fakeSession{capabilities: []string{"cap-2"}, alive: true}
	connector := &fakeConnector{
		results: []fakeConnectResult{
			{session: firstSession},
			{session: secondSession, wait: reconnectGate},
		},
	}

	service := NewService(
		store,
		WithConnector(connector),
		WithReconnectInterval(10*time.Millisecond),
	)
	defer service.Close()

	neItem, err := service.Register("enb-1", "127.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	waitForNEStatus(t, service, neItem.ID, "connected")

	firstSession.setAlive(false)

	waitForConnectCalls(t, connector, 2)
	waitForNEStatus(t, service, neItem.ID, "reconnecting")

	close(reconnectGate)

	waitForNEStatus(t, service, neItem.ID, "connected")
	waitForCapability(t, service, neItem.ID, "cap-2")
}

func TestApplyTransactionSuccess(t *testing.T) {
	store := memory.New()
	service := NewService(store)

	neItem, err := service.Register("enb-1", "127.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	session := &fakeSession{alive: true}
	service.setManagedConn(neItem.ID, &managedConn{client: session})

	err = service.ApplyTransaction(neItem.ID, []model.InventoryObject{
		{
			DN:    "SubNetwork=LTE,ManagedElement=enb-1,ENBFunction=1,EUtranCellFDD=cell-1",
			Class: "EUtranCellFDD",
			Attributes: map[string]string{
				"pci": "101",
			},
		},
	})
	if err != nil {
		t.Fatalf("apply transaction: %v", err)
	}

	if ops := session.snapshotOperations(); !equalStrings(ops, []string{"lock", "edit", "validate", "commit", "unlock"}) {
		t.Fatalf("unexpected operation order: %#v", ops)
	}

	edits := session.snapshotEdits()
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].datastore != netconf.DatastoreCandidate {
		t.Fatalf("unexpected edit datastore: %q", edits[0].datastore)
	}
	if edits[0].payload != "<config><EUtranCellFDD><pci>101</pci></EUtranCellFDD></config>" {
		t.Fatalf("unexpected edit payload: %q", edits[0].payload)
	}
}

func TestApplyTransactionValidateFailureDiscardsAndUnlocks(t *testing.T) {
	store := memory.New()
	service := NewService(store)

	neItem, err := service.Register("enb-1", "127.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	session := &fakeSession{
		alive:       true,
		validateErr: errors.New("validate failed"),
	}
	service.setManagedConn(neItem.ID, &managedConn{client: session})

	err = service.ApplyTransaction(neItem.ID, []model.InventoryObject{
		{
			Class: "EUtranCellFDD",
			Attributes: map[string]string{
				"pci": "101",
			},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if ops := session.snapshotOperations(); !equalStrings(ops, []string{"lock", "edit", "validate", "discard", "unlock"}) {
		t.Fatalf("unexpected operation order: %#v", ops)
	}
}

func TestApplyTransactionEditFailureDiscardsAndUnlocks(t *testing.T) {
	store := memory.New()
	service := NewService(store)

	neItem, err := service.Register("enb-1", "127.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	session := &fakeSession{
		alive:   true,
		editErr: errors.New("edit failed"),
	}
	service.setManagedConn(neItem.ID, &managedConn{client: session})

	err = service.ApplyTransaction(neItem.ID, []model.InventoryObject{
		{
			Class: "EUtranCellFDD",
			Attributes: map[string]string{
				"pci": "101",
			},
		},
	})
	if err == nil {
		t.Fatal("expected edit error")
	}

	if ops := session.snapshotOperations(); !equalStrings(ops, []string{"lock", "edit", "discard", "unlock"}) {
		t.Fatalf("unexpected operation order: %#v", ops)
	}
}

func waitForNEStatus(t *testing.T, service *Service, neID, status string) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		item, ok := service.Get(neID)
		if ok && item.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	item, _ := service.Get(neID)
	t.Fatalf("timed out waiting for status %q, current status %q", status, item.Status)
}

func waitForCapability(t *testing.T, service *Service, neID, capability string) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		item, ok := service.Get(neID)
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		for _, current := range item.Capabilities {
			if current == capability {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	item, _ := service.Get(neID)
	t.Fatalf("timed out waiting for capability %q, current capabilities %#v", capability, item.Capabilities)
}

func waitForConnectCalls(t *testing.T, connector *fakeConnector, expected int) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if connector.callCount() >= expected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %d connect calls, got %d", expected, connector.callCount())
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
