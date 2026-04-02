package inventory

import (
	"errors"
	"fmt"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
)

type Store interface {
	GetNE(id string) (model.NetworkElement, bool)
	SaveInventorySnapshot(snapshot model.InventorySnapshot)
	GetLatestInventorySnapshot(neID string) (model.InventorySnapshot, bool)
}

type RPCProvider interface {
	WithRPCClient(neID string, fn func(netconf.RPCClient) error) error
}

type Service struct {
	store       Store
	rpcProvider RPCProvider
}

func NewService(store Store, rpcProvider RPCProvider) *Service {
	return &Service{
		store:       store,
		rpcProvider: rpcProvider,
	}
}

func (s *Service) Sync(neID string) (model.InventorySnapshot, error) {
	neItem, ok := s.store.GetNE(neID)
	if !ok {
		return model.InventorySnapshot{}, errors.New("network element not found")
	}
	if s.rpcProvider == nil {
		return model.InventorySnapshot{}, errors.New("inventory reader is not configured")
	}

	configReply, stateReply, err := s.readInventoryData(neID)
	if err != nil {
		return model.InventorySnapshot{}, err
	}

	objects, err := buildInventoryObjects(configReply, stateReply)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	if err := validateInventoryObjects(objects); err != nil {
		return model.InventorySnapshot{}, err
	}

	snapshot := model.InventorySnapshot{
		ID:       id.New("inv"),
		NEID:     neItem.ID,
		SyncedAt: time.Now().UTC(),
		Objects:  objects,
	}

	s.store.SaveInventorySnapshot(snapshot)
	return snapshot, nil
}

func (s *Service) GetLatest(neID string) (model.InventorySnapshot, bool) {
	return s.store.GetLatestInventorySnapshot(neID)
}

func (s *Service) readInventoryData(neID string) ([]byte, []byte, error) {
	var configReply []byte
	var stateReply []byte

	err := s.rpcProvider.WithRPCClient(neID, func(client netconf.RPCClient) error {
		reply, err := client.GetConfig(netconf.DatastoreRunning, "")
		if err != nil {
			return fmt.Errorf("read running config: %w", err)
		}
		configReply = append([]byte(nil), reply...)

		reply, err = client.Get("")
		if err != nil {
			return fmt.Errorf("read operational data: %w", err)
		}
		stateReply = append([]byte(nil), reply...)

		return nil
	})
	if err == nil {
		return configReply, stateReply, nil
	}

	switch {
	case netconf.IsTimeout(err):
		return nil, nil, fmt.Errorf("inventory read timed out: %w", err)
	case netconf.IsReadFailed(err):
		return nil, nil, fmt.Errorf("inventory read failed: %w", err)
	default:
		return nil, nil, err
	}
}
