package cm

import (
	"errors"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
)

type ApplyChangeInput struct {
	NEID      string `json:"ne_id"`
	Parameter string `json:"parameter"`
	Value     string `json:"value"`
}

type Store interface {
	GetNE(id string) (model.NetworkElement, bool, error)
	SaveCMRequest(req model.CMRequest) error
	ListCMRequests() ([]model.CMRequest, error)
	GetCMRequest(id string) (model.CMRequest, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ApplyChange(input ApplyChangeInput) (model.CMRequest, error) {
	_, ok, err := s.store.GetNE(input.NEID)
	if err != nil {
		return model.CMRequest{}, err
	}
	if !ok {
		return model.CMRequest{}, errors.New("network element not found")
	}

	now := time.Now().UTC()
	req := model.CMRequest{
		ID:        id.New("cm"),
		NEID:      strings.TrimSpace(input.NEID),
		Parameter: strings.TrimSpace(input.Parameter),
		Value:     strings.TrimSpace(input.Value),
		Status:    "running",
		CreatedAt: now,
		UpdatedAt: now,
	}

	req.Steps = append(req.Steps, newStep("lock", "success", "candidate config locked"))
	req.Steps = append(req.Steps, newStep("edit-config", "success", "config payload applied"))

	if req.Parameter == "" || req.Value == "" || strings.Contains(strings.ToLower(req.Parameter), "forbidden") {
		req.Steps = append(req.Steps, newStep("validate", "failed", "invalid parameter or value"))
		req.Steps = append(req.Steps, newStep("unlock", "success", "candidate config unlocked"))
		req.Status = "failed"
		req.UpdatedAt = time.Now().UTC()
		err := s.store.SaveCMRequest(req)
		if err != nil {
			return req, err
		}
		return req, nil
	}

	req.Steps = append(req.Steps, newStep("validate", "success", "configuration is valid"))
	req.Steps = append(req.Steps, newStep("commit", "success", "configuration committed"))
	req.Steps = append(req.Steps, newStep("unlock", "success", "candidate config unlocked"))
	req.Status = "success"
	req.UpdatedAt = time.Now().UTC()

	err = s.store.SaveCMRequest(req)
	if err != nil {
		return req, err
	}
	return req, nil
}

func (s *Service) ListRequests() ([]model.CMRequest, error) {
	return s.store.ListCMRequests()
}

func (s *Service) GetCMRequest(id string) (model.CMRequest, error) {
	return s.store.GetCMRequest(id)
}

func newStep(name, status, message string) model.CMStep {
	return model.CMStep{
		Name:      name,
		Status:    status,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
}
