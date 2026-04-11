package cm

import (
	"testing"

	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/postgres"
)

func TestApplyChangeSuccess(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}
	neService := ne.NewService(store)
	cmService := NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	req, err := cmService.ApplyChange(ApplyChangeInput{
		NEID:      neItem.ID,
		Parameter: "cell.pci",
		Value:     "101",
	})
	if err != nil {
		t.Fatalf("apply change: %v", err)
	}
	if req.Status != "success" {
		t.Fatalf("expected success status, got %s", req.Status)
	}
	if len(req.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(req.Steps))
	}
}

func TestApplyChangeValidationFail(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}
	neService := ne.NewService(store)
	cmService := NewService(store)

	neItem, err := neService.Register("enb-2", "10.0.0.2", "vendor-b", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	req, err := cmService.ApplyChange(ApplyChangeInput{
		NEID:      neItem.ID,
		Parameter: "forbidden.parameter",
		Value:     "1",
	})
	if err != nil {
		t.Fatalf("apply change: %v", err)
	}
	if req.Status != "failed" {
		t.Fatalf("expected failed status, got %s", req.Status)
	}
	if len(req.Steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(req.Steps))
	}
}
