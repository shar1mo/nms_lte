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
		t.Fatalf("expected severity normalized to 'critical', got %s", event.Severity)
	}

	if event.Message != "link down" {
		t.Fatalf("unexpected message: %s", event.Message)
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
	neItem, _ := neService.Register("enb-2", "10.0.0.2", "vendor-b", nil)

	_, err = faultService.ReportEvent(neItem.ID, "major", "   ")
	if err == nil {
		t.Fatalf("expected error for empty message")
	}
}

func TestCheckHeartbeatHealthy(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, _ := neService.Register("enb-3", "10.0.0.3", "vendor-c", nil)

	hb, err := faultService.CheckHeartbeat(neItem.ID, true)
	if err != nil {
		t.Fatalf("check heartbeat: %v", err)
	}

	if !hb.Healthy {
		t.Fatalf("expected healthy=true")
	}

	events := faultService.ListEvents(neItem.ID)
	if len(events) != 0 {
		t.Fatalf("expected no fault events, got %d", len(events))
	}
}

func TestCheckHeartbeatFailureCreatesEvent(t *testing.T) {
	store := memory.New()
	neService := ne.NewService(store)
	faultService := NewService(store)

	neItem, _ := neService.Register("enb-4", "10.0.0.4", "vendor-d", nil)

	_, err := faultService.CheckHeartbeat(neItem.ID, false)
	if err != nil {
		t.Fatalf("check heartbeat: %v", err)
	}

	events := faultService.ListEvents(neItem.ID)
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

	neItem, _ := neService.Register("enb-5", "10.0.0.5", "vendor-e", nil)

	_, _ = faultService.CheckHeartbeat(neItem.ID, true)

	hb, ok := faultService.GetHeartbeat(neItem.ID)
	if !ok {
		t.Fatalf("expected heartbeat to exist")
	}
	if hb.NEID != neItem.ID {
		t.Fatalf("expected NEID %s, got %s", neItem.ID, hb.NEID)
	}
}