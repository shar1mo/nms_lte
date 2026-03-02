package fault

import (
	"errors"
	"strings"
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

func (s *Service) ReportEvent(neID, severity, message string) (model.FaultEvent, error) {
	if _, ok := s.store.GetNE(neID); !ok {
		return model.FaultEvent{}, errors.New("network element not found")
	}
	if strings.TrimSpace(message) == "" {
		return model.FaultEvent{}, errors.New("message is required")
	}
	if strings.TrimSpace(severity) == "" {
		severity = "warning"
	}
	event := model.FaultEvent{
		ID:        id.New("fault"),
		NEID:      strings.TrimSpace(neID),
		Severity:  strings.ToLower(strings.TrimSpace(severity)),
		Message:   strings.TrimSpace(message),
		CreatedAt: time.Now().UTC(),
	}
	s.store.AddFaultEvent(event)
	return event, nil
}

func (s *Service) CheckHeartbeat(neID string, healthy bool) (model.HeartbeatStatus, error) {
	if _, ok := s.store.GetNE(neID); !ok {
		return model.HeartbeatStatus{}, errors.New("network element not found")
	}
	hb := model.HeartbeatStatus{
		NEID:      neID,
		Healthy:   healthy,
		CheckedAt: time.Now().UTC(),
	}
	s.store.SaveHeartbeat(hb)

	if !healthy {
		s.store.AddFaultEvent(model.FaultEvent{
			ID:        id.New("fault"),
			NEID:      neID,
			Severity:  "major",
			Message:   "heartbeat check failed",
			CreatedAt: time.Now().UTC(),
		})
	}

	return hb, nil
}

func (s *Service) GetHeartbeat(neID string) (model.HeartbeatStatus, bool) {
	return s.store.GetHeartbeat(neID)
}

func (s *Service) ListEvents(neID string) []model.FaultEvent {
	return s.store.ListFaultEvents(neID)
}
