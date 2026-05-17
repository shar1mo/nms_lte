package fault

import (
	"testing"

	"nms_lte/internal/service/ne"
	"nms_lte/internal/store/memory"
)

func TestReportEventSuccess(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, err := neService.Register("enb-1", "10.0.0.1", "vendor-a", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	event, err := faultService.ReportEvent(neItem.ID, "CRITICAL", "link down")
	if err != nil {
		t.Fatalf("report event: %v", err)
	}

	if event.NEID != neItem.ID {
		t.Fatalf("expected NEID %s, got %s", neItem.ID, event.NEID)
	}

	if event.Severity != "critical" {
		t.Fatalf("expected severity 'critical', got %s", event.Severity)
	}

	if event.Message != "link down" {
		t.Fatalf("unexpected message: %s", event.Message)
	}

	if event.ID == "" {
		t.Fatalf("expected event ID")
	}

	if event.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt")
	}
}

func TestReportEventValidationFail(t *testing.T) {
	store := memory.New()
	faultService := NewService(store)

	_, err := faultService.ReportEvent("unknown-ne", "major", "test")
	if err == nil {
		t.Fatalf("expected error for unknown NE")
	}

	neService := ne.NewService(store)

	neItem, err := neService.Register("enb-2", "10.0.0.2", "vendor-b", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = faultService.ReportEvent(neItem.ID, "major", "   ")
	if err == nil {
		t.Fatalf("expected error for empty message")
	}
}

func TestCheckHeartbeatHealthy(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, err := neService.Register("enb-3", "10.0.0.3", "vendor-c", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	hb, err := faultService.CheckHeartbeat(neItem.ID, true)
	if err != nil {
		t.Fatalf("check heartbeat: %v", err)
	}

	if !hb.Healthy {
		t.Fatalf("expected healthy=true")
	}

	if hb.NEID != neItem.ID {
		t.Fatalf("expected NEID %s, got %s", neItem.ID, hb.NEID)
	}

	if hb.CheckedAt.IsZero() {
		t.Fatalf("expected CheckedAt")
	}

	events, err := faultService.ListEvents(neItem.ID, 100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("expected no fault events, got %d", len(events))
	}
}

func TestCheckHeartbeatFailureCreatesEvent(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, err := neService.Register("enb-4", "10.0.0.4", "vendor-d", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = faultService.CheckHeartbeat(neItem.ID, false)
	if err != nil {
		t.Fatalf("check heartbeat: %v", err)
	}

	events, err := faultService.ListEvents(neItem.ID, 100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 fault event, got %d", len(events))
	}

	if events[0].Severity != "major" {
		t.Fatalf("expected severity 'major', got %s", events[0].Severity)
	}

	if events[0].Message != "heartbeat check failed" {
		t.Fatalf("unexpected message: %s", events[0].Message)
	}
}

func TestGetHeartbeat(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, err := neService.Register("enb-5", "10.0.0.5", "vendor-e", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, err = faultService.CheckHeartbeat(neItem.ID, true)
	if err != nil {
		t.Fatalf("check heartbeat: %v", err)
	}

	hb, ok, err := faultService.GetHeartbeat(neItem.ID)
	if err != nil {
		t.Fatalf("get heartbeat: %v", err)
	}

	if !ok {
		t.Fatalf("expected heartbeat to exist")
	}

	if hb.NEID != neItem.ID {
		t.Fatalf("expected NEID %s, got %s", neItem.ID, hb.NEID)
	}

	if !hb.Healthy {
		t.Fatalf("expected healthy=true")
	}
}

func TestListEventsLimit(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, err := neService.Register("enb-6", "10.0.0.6", "vendor-f", nil)
	if err != nil {
		t.Fatalf("register ne: %v", err)
	}

	_, _ = faultService.ReportEvent(neItem.ID, "minor", "event 1")
	_, _ = faultService.ReportEvent(neItem.ID, "major", "event 2")
	_, _ = faultService.ReportEvent(neItem.ID, "critical", "event 3")

	events, err := faultService.ListEvents(neItem.ID, 2)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}