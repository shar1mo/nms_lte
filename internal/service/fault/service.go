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
	SaveInventorySnapshot(snapshot model.InventorySnapshot) error
	GetLatestInventorySnapshot(neID string) (model.InventorySnapshot, error)
	SaveHeartbeat(hb model.HeartbeatStatus)
	GetHeartbeat(neID string) (model.HeartbeatStatus, bool)
	AddFaultEvent(event model.FaultEvent)
	ListFaultEvents(neID string) []model.FaultEvent
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
	s.store.AddFaultEvent(event)
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
