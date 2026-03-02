package inventory

import (
	"testing"

	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/memory"
)

func TestSyncUnknownNE(t *testing.T) {
	service := NewService(memory.New())
	_, err := service.Sync("unknown")
	if err == nil {
		t.Fatal("expected error for unknown network element")
	}
}

func TestSyncSuccess(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	service := NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	snapshot, err := service.Sync(neItem.ID)
	if err != nil {
		t.Fatalf("sync inventory: %v", err)
	}

	if snapshot.NEID != neItem.ID {
		t.Fatalf("snapshot ne_id mismatch: %s", snapshot.NEID)
	}
	if len(snapshot.Objects) == 0 {
		t.Fatal("expected inventory objects")
	}

	latest, ok := service.GetLatest(neItem.ID)
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if latest.ID != snapshot.ID {
		t.Fatalf("latest snapshot mismatch: %s != %s", latest.ID, snapshot.ID)
	}
}
