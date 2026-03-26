package ne

import (
	"errors"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
	"nms_lte/internal/store/memory"
)

var ErrNENotFound = errors.New("ne doesn't exist")

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(name, address, vendor string, capabilities []string) (model.NetworkElement, error) {
	if strings.TrimSpace(name) == "" {
		return model.NetworkElement{}, errors.New("name is required")
	}
	if strings.TrimSpace(address) == "" {
		return model.NetworkElement{}, errors.New("address is required")
	}
	if len(capabilities) == 0 {
		capabilities = []string{
			"urn:ietf:params:netconf:base:1.0",
			"urn:ietf:params:netconf:base:1.1",
		}
	}
	now := time.Now().UTC()
	ne := model.NetworkElement{
		ID:           id.New("ne"),
		Name:         strings.TrimSpace(name),
		Address:      strings.TrimSpace(address),
		Vendor:       strings.TrimSpace(vendor),
		Status:       "active",
		Capabilities: capabilities,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.store.SaveNE(ne)
	return ne, nil
}

func (s *Service) UnRegister(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}

	ne, ok := s.store.GetNE(id)
	if !ok {
		return ErrNENotFound
	}

	if ne.Status == "active" {
		return errors.New("status is active, deactivate ne first")
	}

	if !s.store.DeleteNE(ne.ID) {
		return errors.New("failed to delete ne")
	}

	return nil
}

func (s *Service) List() []model.NetworkElement {
	return s.store.ListNE()
}

func (s *Service) Get(neID string) (model.NetworkElement, bool) {
	return s.store.GetNE(neID)
}
