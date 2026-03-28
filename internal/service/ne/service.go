package ne

import (
	"errors"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
	"nms_lte/internal/store/memory"
	"nms_lte/internal/store/postgres"
)

var ErrNENotFound = errors.New("ne doesn't exist")

type Service struct {
	store *memory.Store
}

type ServicePG struct {
	store *postgres.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func NewServicePG(store *postgres.Store) *ServicePG {
	return &ServicePG{store: store}
}

//in-memory methods
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

	if !s.store.DeleteNE(id) {
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

//posgres methods
func (s *ServicePG) RegisterPG(name, address, vendor string, capabilities []string) (model.NetworkElement, error) {
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
	if err := s.store.SaveNE(ne); err != nil {
		return model.NetworkElement{}, err
	}

	return ne, nil
}

func (s *ServicePG) UnRegisterPG(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}

	ne, ok, err := s.store.GetNE(id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNENotFound
	}

	//if ne.Status == "active" {
	//	return errors.New("status is active, deactivate ne first")
	//}

	if err := s.store.DeleteNE(ne.ID); err != nil {
		return errors.New("failed to delete ne")
	}

	return nil
}

func (s *ServicePG) ListPG() ([]model.NetworkElement, error) {
	out, err := s.store.ListNE()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *ServicePG) GetPG(neID string) (model.NetworkElement, bool) {
	ne, ok, err := s.store.GetNE(neID)
	if err != nil {
		return model.NetworkElement{}, false
	}
	return ne, ok
}