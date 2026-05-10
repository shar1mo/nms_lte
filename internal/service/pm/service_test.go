package pm

import (
	"testing"
	"time"

	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/postgres"
)

func TestCollectPostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-1",
		"10.10.0.1",
		"vendor-a",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	sample, err := pmService.Collect(neItem.ID, "cpu_load")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if sample.ID == "" {
		t.Fatalf("expected sample id")
	}

	if sample.NEID != neItem.ID {
		t.Fatalf("expected ne id %s, got %s", neItem.ID, sample.NEID)
	}

	if sample.Metric != "cpu_load" {
		t.Fatalf("expected metric cpu_load, got %s", sample.Metric)
	}

	if sample.CollectedAt.IsZero() {
		t.Fatalf("expected collected_at")
	}
}

func TestCollectUnknownNEPostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	pmService := NewService(store)

	_, err = pmService.Collect("unknown-ne", "availability")
	if err == nil {
		t.Fatalf("expected error for unknown ne")
	}
}

func TestCollectDefaultMetricPostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-2",
		"10.10.0.2",
		"vendor-b",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	sample, err := pmService.Collect(neItem.ID, "")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if sample.Metric != "availability" {
		t.Fatalf(
			"expected default metric availability, got %s",
			sample.Metric,
		)
	}
}

func TestCollectAvailabilityRangePostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-3",
		"10.10.0.3",
		"vendor-c",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	sample, err := pmService.Collect(neItem.ID, "availability")
	if err != nil {
		t.Fatalf("collect pm: %v", err)
	}

	if sample.Value < 95 || sample.Value > 100 {
		t.Fatalf("availability out of range: %f", sample.Value)
	}
}

func TestListSamplesPostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-4",
		"10.10.0.4",
		"vendor-d",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = pmService.Collect(neItem.ID, "users")
	if err != nil {
		t.Fatalf("collect pm #1: %v", err)
	}

	_, err = pmService.Collect(neItem.ID, "users")
	if err != nil {
		t.Fatalf("collect pm #2: %v", err)
	}

	samples, err := pmService.List(
		neItem.ID,
		"users",
		&time.Time{},
		&time.Time{},
		10,
	)
	if err != nil {
		t.Fatalf("list samples: %v", err)
	}

	if len(samples) < 2 {
		t.Fatalf("expected at least 2 samples, got %d", len(samples))
	}

	for _, sample := range samples {
		if sample.NEID != neItem.ID {
			t.Fatalf(
				"expected ne id %s, got %s",
				neItem.ID,
				sample.NEID,
			)
		}

		if sample.Metric != "users" {
			t.Fatalf(
				"expected metric users, got %s",
				sample.Metric,
			)
		}
	}
}

func TestListSamplesLimitPostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-5",
		"10.10.0.5",
		"vendor-e",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := pmService.Collect(neItem.ID, "cpu_load")
		if err != nil {
			t.Fatalf("collect pm: %v", err)
		}
	}

	samples, err := pmService.List(
		neItem.ID,
		"cpu_load",
		&time.Time{},
		&time.Time{},
		2,
	)
	if err != nil {
		t.Fatalf("list samples: %v", err)
	}

	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}
}

func TestListSamplesTimeRangePostgres(t *testing.T) {
	store, err := postgres.New(postgres.ConnString)
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}

	neService := ne.NewService(store)
	pmService := NewService(store)

	neItem, err := neService.Register(
		"enb-pg-6",
		"10.10.0.6",
		"vendor-f",
		nil,
	)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = pmService.Collect(neItem.ID, "cpu_load")
	if err != nil {
		t.Fatalf("collect old sample: %v", err)
	}

	time.Sleep(2 * time.Second)

	from := time.Now().UTC()

	time.Sleep(1 * time.Second)

	_, err = pmService.Collect(neItem.ID, "cpu_load")
	if err != nil {
		t.Fatalf("collect new sample: %v", err)
	}

	to := time.Now().UTC().Add(1 * time.Second)

	samples, err := pmService.List(
		neItem.ID,
		"cpu_load",
		&from,
		&to,
		10,
	)
	if err != nil {
		t.Fatalf("list samples: %v", err)
	}

	if len(samples) == 0 {
		t.Fatalf("expected samples in time range")
	}

	for _, sample := range samples {
		if sample.CollectedAt.Before(from) {
			t.Fatalf(
				"sample before from range: %v < %v",
				sample.CollectedAt,
				from,
			)
		}

		if sample.CollectedAt.After(to) {
			t.Fatalf(
				"sample after to range: %v > %v",
				sample.CollectedAt,
				to,
			)
		}
	}
}