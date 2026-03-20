package pm

import (
	"testing"

	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/memory"
)

func TestCollect(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	pmCollect, err := pmService.Collect(neItem.ID, "dl_latency")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if pmCollect.NEID != neItem.ID {
		t.Fatalf("expected NEID %s, got %s", neItem.ID, pmCollect.ID)
	}

	if pmCollect.Metric != "dl_latency" {
		t.Fatalf("expected metric 'dl_latency', got %s", pmCollect.Metric)
	}
}

func TestUnknownNE(t *testing.T) {
	store := memory.New()
	pmService := NewService(store)

	_, err := pmService.Collect("unknown-ne", "cpu_load")
	if err == nil {
		t.Fatalf("expected error for unknown network element")
	}
}

func TestCollectDefaultMetric(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, _ := neService.Register("enb-2", "10.0.0.2", "vendor-b", nil)

	pmCollect, err := pmService.Collect(neItem.ID, "")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if pmCollect.Metric != "availability" {
		t.Fatalf("expected default metric 'availability', got %s", pmCollect.Metric)
	}
}

func TestCollectValueRange(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, _ := neService.Register("enb-3", "10.0.0.3", "vendor-c", nil)

	pmCollect, err := pmService.Collect(neItem.ID, "availability")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if pmCollect.Value < 95 || pmCollect.Value > 100 {
		t.Fatalf("availability out of range: %f", pmCollect.Value)
	}
}

func TestListSamples(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, _ := neService.Register("enb-4", "10.0.0.4", "vendor-d", nil)

	_, _ = pmService.Collect(neItem.ID, "cpu_load")
	_, _ = pmService.Collect(neItem.ID, "cpu_load")

	samples := pmService.List(neItem.ID, "cpu_load", 10)

	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}

	for _, s := range samples {
		if s.Metric != "cpu_load" {
			t.Fatalf("unexpected metric %s", s.Metric)
		}
	}
}