package inventory

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/infra/netconf"
	"nms_lte/internal/model"
)

type Store interface {
	GetNE(id string) (model.NetworkElement, bool, error)
	SaveInventorySnapshot(snapshot model.InventorySnapshot) error
	GetLatestInventorySnapshot(neID string) (model.InventorySnapshot, error)
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
	if strings.TrimSpace(neID) == "" {
		return model.InventorySnapshot{}, errors.New("neID is required")
	}

	neItem, ok, err := s.store.GetNE(neID)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	if !ok {
		return model.InventorySnapshot{}, errors.New("network element not found")
	}

	var objects []model.InventoryObject

	if s.rpcProvider == nil {
		objects = []model.InventoryObject{
			{
				DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s", neItem.ID),
				Class: "ManagedElement",
				Attributes: map[string]string{
					"name":   neItem.Name,
					"vendor": neItem.Vendor,
				},
			},
			{
				DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1", neItem.ID),
				Class: "ENBFunction",
				Attributes: map[string]string{
					"status": neItem.Status,
				},
			},
			{
				DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1,EUtranCellFDD=cell-1", neItem.ID),
				Class: "EUtranCellFDD",
				Attributes: map[string]string{
					"pci":      "100",
					"earfcnDL": "300",
				},
			},
			{
				DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1,EUtranFrequency=1", neItem.ID),
				Class: "EUtranFrequency",
				Attributes: map[string]string{
					"earfcn": "300",
				},
			},
		}
	} else {
		configReply, stateReply, err := s.readInventoryData(neID)
		if err != nil {
			return model.InventorySnapshot{}, err
		}

		objects, err = buildInventoryObjects(configReply, stateReply)
		if err != nil {
			return model.InventorySnapshot{}, err
		}
		if err := validateInventoryObjects(objects); err != nil {
			return model.InventorySnapshot{}, err
		}
	}

	snapshot := model.InventorySnapshot{
		ID:       id.New("inv"),
		NEID:     neItem.ID,
		SyncedAt: time.Now().UTC(),
		Objects:  objects,
	}

	if err := s.store.SaveInventorySnapshot(snapshot); err != nil {
		return model.InventorySnapshot{}, err
	}

	return snapshot, nil
}

func (s *Service) GetLatest(neID string) (model.InventorySnapshot, error) {
	if strings.TrimSpace(neID) == "" {
		return model.InventorySnapshot{}, errors.New("neID is required")
	}

	snapshot, err := s.store.GetLatestInventorySnapshot(neID)
	if err != nil {
		return model.InventorySnapshot{}, err
	}
	if snapshot.ID == "" {
		return model.InventorySnapshot{}, errors.New("snapshot not found")
	}

	return snapshot, nil
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