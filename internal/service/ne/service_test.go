package ne

import (
	"testing"

	"nms_lte/internal/store/memory"
)

func TestNERegisterValidation(t *testing.T) {
	store := memory.New()
	neService := NewService(store)

	_, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = neService.Register("", "10.0.0.1", "vendor-a", nil)
	if err == nil || err.Error() != "name is required" {
		t.Fatalf("expected 'name is required', got %s", err)
	}

	_, err = neService.Register("enb-1", "", "vendor-a", nil)
	if err == nil || err.Error() != "address is required" {
		t.Fatalf("expected 'address is required', got %s", err)
	}
}

func TestTrimSpace(t *testing.T) {
	store := memory.New()
	neService := NewService(store)

	neItem, err := neService.Register("  enb-2  ", "  10.0.0.2  ", " vendor-b ", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	if neItem.Name != "enb-2" {
		t.Fatalf("expected trimmed name, got '%s'", neItem.Name)
	}
	if neItem.Address != "10.0.0.2" {
		t.Fatalf("expected trimmed address, got '%s'", neItem.Address)
	}
	if neItem.Vendor != "vendor-b" {
		t.Fatalf("expected trimmed vendor, got '%s'", neItem.Vendor)
	}

	if neItem.Status != "active" {
		t.Fatalf("expected status 'active', got %s", neItem.Status)
	}
}

func TestGetListNE(t *testing.T) {
	store := memory.New()
	neService := NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	saved, ok := neService.Get(neItem.ID)
	if !ok {
		t.Fatalf("expected NE to be saved")
	}

	if saved.ID != neItem.ID {
		t.Fatalf("saved NE mismatch")
	}

	list := neService.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 NE, got %d", len(list))
	}

	if list[0].ID != neItem.ID {
		t.Fatalf("unexpected NE in list")
	}
}