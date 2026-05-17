package fault

import (
	"errors"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
)

type Store interface {
	GetNE(id string) (model.NetworkElement, bool, error)
	SaveHeartbeat(hb model.HeartbeatStatus) error
	GetHeartbeat(neID string) (model.HeartbeatStatus, bool, error)
	AddFaultEvent(event model.FaultEvent) error
	ListFaultEvents(neID string, limit int) ([]model.FaultEvent, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ReportEvent(neID, severity, message string) (model.FaultEvent, error) {
	_, ok, err := s.store.GetNE(neID)
	if err != nil {
		return model.FaultEvent{}, err
	}
	if !ok {
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

	if err := s.store.AddFaultEvent(event); err != nil {
		return model.FaultEvent{}, err
	}

	return event, nil
}

func (s *Service) CheckHeartbeat(neID string, healthy bool) (model.HeartbeatStatus, error) {
	_, ok, err := s.store.GetNE(neID)
	if err != nil {
		return model.HeartbeatStatus{}, err
	}
	if !ok {
		return model.HeartbeatStatus{}, errors.New("network element not found")
	}

	hb := model.HeartbeatStatus{
		NEID:      strings.TrimSpace(neID),
		Healthy:   healthy,
		CheckedAt: time.Now().UTC(),
	}

	if err := s.store.SaveHeartbeat(hb); err != nil {
		return model.HeartbeatStatus{}, err
	}

	if !healthy {
		event := model.FaultEvent{
			ID:        id.New("fault"),
			NEID:      strings.TrimSpace(neID),
			Severity:  "major",
			Message:   "heartbeat check failed",
			CreatedAt: time.Now().UTC(),
		}

		if err := s.store.AddFaultEvent(event); err != nil {
			return model.HeartbeatStatus{}, err
		}
	}

	return hb, nil
}

func (s *Service) GetHeartbeat(neID string) (model.HeartbeatStatus, bool, error) {
	return s.store.GetHeartbeat(strings.TrimSpace(neID))
}

func (s *Service) ListEvents(neID string, limit int) ([]model.FaultEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	return s.store.ListFaultEvents(strings.TrimSpace(neID), limit)
}
