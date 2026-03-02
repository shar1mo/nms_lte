package pm

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
	"nms_lte/internal/store/memory"
)

type Service struct {
	store *memory.Store
	rnd   *rand.Rand
}

func NewService(store *memory.Store) *Service {
	return &Service{
		store: store,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) Collect(neID, metric string) (model.PMSample, error) {
	if _, ok := s.store.GetNE(neID); !ok {
		return model.PMSample{}, errors.New("network element not found")
	}
	if strings.TrimSpace(metric) == "" {
		metric = "availability"
	}
	metric = strings.TrimSpace(metric)

	sample := model.PMSample{
		ID:          id.New("pm"),
		NEID:        neID,
		Metric:      metric,
		Value:       s.generateValue(metric),
		CollectedAt: time.Now().UTC(),
	}

	s.store.AddPMSample(sample)
	return sample, nil
}

func (s *Service) List(neID, metric string, limit int) []model.PMSample {
	return s.store.ListPMSamples(neID, metric, limit)
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
