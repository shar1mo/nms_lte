package pm

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
)

type Store interface {
	GetNE(id string) (model.NetworkElement, bool, error)
	AddPMSample(sample model.PMSample) error
	ListPMSamples(neID, metric string, from, to *time.Time, limit int) ([]model.PMSample, error)
}

type Service struct {
	store Store
	rnd   *rand.Rand
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) Collect(neID, metric string) (model.PMSample, error) {
	_, ok, err := s.store.GetNE(neID)
	if err != nil {
		return model.PMSample{}, err
	}
	if !ok {
		return model.PMSample{}, errors.New("network element not found")
	}
	if strings.TrimSpace(metric) == "" {
		metric = "availability"
	}
	metric = strings.ToLower(strings.TrimSpace(metric))

	sample := model.PMSample{
		ID:          id.New("pm"),
		NEID:        neID,
		Metric:      metric,
		Value:       s.generateValue(metric),
		CollectedAt: time.Now().UTC(),
	}

	err = s.store.AddPMSample(sample)
	if err != nil {
		return model.PMSample{}, err
	}
	return sample, nil
}

func (s *Service) List(neID, metric string, from, to *time.Time, limit int) ([]model.PMSample, error) {
	neID = strings.TrimSpace(neID)
	metric = strings.ToLower(strings.TrimSpace(metric))

	return s.store.ListPMSamples(
		neID,
		metric,
		from,
		to,
		limit,
	)
}

func (s *Service) generateValue(metric string) float64 {
	switch strings.ToLower(metric) {
	case "availability":
		return 95 + s.rnd.Float64()*5
	case "cpu_load":
		return 10 + s.rnd.Float64()*70
	case "users":
		return float64(s.rnd.Intn(200) + 20)
	default:
		return s.rnd.Float64() * 100
	}
}
