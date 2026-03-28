package ne

import (
	"testing"

	"nms_lte/internal/store/memory"
	"nms_lte/internal/store/postgres"
)

func TestNERegisterValidation(t *testing.T) {
	//in-memory
	memStore := memory.New()
	memService := NewService(memStore)

	_, err := memService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("memory register ne: %v", err)
	}

	_, err = memService.Register("", "10.0.0.1", "vendor-a", nil)
	if err == nil || err.Error() != "name is required" {
		t.Fatalf("expected 'name is required', got %s", err)
	}

	_, err = memService.Register("enb-1", "", "vendor-a", nil)
	if err == nil || err.Error() != "address is required" {
		t.Fatalf("expected 'address is required', got %s", err)
	}

	//postgres
	pgStore, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	pgService := NewServicePG(pgStore)

	_, err = pgService.RegisterPG("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("postgres register ne: %v", err)
	}

	_, err = pgService.RegisterPG("", "10.0.0.1", "vendor-a", nil)
	if err == nil || err.Error() != "name is required" {
		t.Fatalf("expected 'name is required' (PG), got %s", err)
	}

	_, err = pgService.RegisterPG("enb-1", "", "vendor-a", nil)
	if err == nil || err.Error() != "address is required" {
		t.Fatalf("expected 'address is required' (PG), got %s", err)
	}
}

func TestTrimSpace(t *testing.T) {
	//in-memory
	memStore := memory.New()
	memService := NewService(memStore)

	neItem, err := memService.Register("  enb-2  ", "  10.0.0.2  ", " vendor-b ", nil)
	if err != nil {
		t.Fatalf("memory register ne: %v", err)
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

	//postgres
	pgStore, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	pgService := NewServicePG(pgStore)

	neItem, err = pgService.RegisterPG("  enb-2  ", "  10.0.0.2  ", " vendor-b ", nil)
	if err != nil {
		t.Fatalf("postgres register ne: %v", err)
	}

	if neItem.Name != "enb-2" {
		t.Fatalf("expected trimmed name (PG), got '%s'", neItem.Name)
	}
	if neItem.Address != "10.0.0.2" {
		t.Fatalf("expected trimmed address (PG), got '%s'", neItem.Address)
	}
	if neItem.Vendor != "vendor-b" {
		t.Fatalf("expected trimmed vendor (PG), got '%s'", neItem.Vendor)
	}
	if neItem.Status != "active" {
		t.Fatalf("expected status 'active' (PG), got %s", neItem.Status)
	}
}

func TestGetListNE(t *testing.T) {
	//in-memory
	memStore := memory.New()
	memService := NewService(memStore)

	neItem, err := memService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("memory register ne: %v", err)
	}

	saved, ok := memService.Get(neItem.ID)
	if !ok {
		t.Fatalf("expected NE to be saved (memory)")
	}
	if saved.ID != neItem.ID {
		t.Fatalf("saved NE mismatch (memory)")
	}

	list := memService.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 NE (memory), got %d", len(list))
	}
	if list[0].ID != neItem.ID {
		t.Fatalf("unexpected NE in list (memory)")
	}

	//postgres
	pgStore, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	pgService := NewServicePG(pgStore)

	neItem, err = pgService.RegisterPG("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("postgres register ne: %v", err)
	}

	savedPG, ok := pgService.GetPG(neItem.ID)
	if !ok {
		t.Fatalf("expected NE to be saved (PG)")
	}
	if savedPG.ID != neItem.ID {
		t.Fatalf("saved NE mismatch (PG)")
	}

	listPG, err := pgService.ListPG()
	if err != nil {
		t.Fatalf("list ne (PG): %v", err)
	}
	if len(listPG) != 1 {
		t.Fatalf("expected 1 NE (PG), got %d", len(listPG))
	}
	if listPG[0].ID != neItem.ID {
		t.Fatalf("unexpected NE in list (PG)")
	}
}