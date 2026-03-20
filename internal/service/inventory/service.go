package inventory

import (
	"errors"
	"fmt"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
	"nms_lte/internal/store/memory"
)

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Sync(neID string) (model.InventorySnapshot, error) {
	ne, ok := s.store.GetNE(neID)
	if !ok {
		return model.InventorySnapshot{}, errors.New("network element not found")
	}

	now := time.Now().UTC()
	objects := []model.InventoryObject{
		{
			DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s", ne.ID),
			Class: "ManagedElement",
			Attributes: map[string]string{
				"name":   ne.Name,
				"vendor": ne.Vendor,
			},
		},
		{
			DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1", ne.ID),
			Class: "ENBFunction",
			Attributes: map[string]string{
				"status": ne.Status,
			},
		},
		{
			DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1,EUtranCellFDD=cell-1", ne.ID),
			Class: "EUtranCellFDD",
			Attributes: map[string]string{
				"pci":      "100",
				"earfcnDL": "300",
			},
		},
		{
			DN:    fmt.Sprintf("SubNetwork=LTE,ManagedElement=%s,ENBFunction=1,EUtranFrequency=1", ne.ID),
			Class: "EUtranFrequency",
			Attributes: map[string]string{
				"earfcn": "300",
			},
		},
	}

	snapshot := model.InventorySnapshot{
		ID:       id.New("inv"),
		NEID:     ne.ID,
		SyncedAt: now,
		Objects:  objects,
	}

	s.store.SaveInventorySnapshot(snapshot)
	return snapshot, nil
}

func (s *Service) GetLatest(neID string) (model.InventorySnapshot, bool) {
	return s.store.GetLatestInventorySnapshot(neID)
}