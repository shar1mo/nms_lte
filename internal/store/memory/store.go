package memory

import (
	"sort"
	"sync"

	"nms_lte/internal/model"
)

type Store struct {
	mu                 sync.RWMutex
	nes                map[string]model.NetworkElement
	inventorySnapshots map[string]model.InventorySnapshot
	cmRequests         map[string]model.CMRequest
	faultEvents        []model.FaultEvent
	heartbeats         map[string]model.HeartbeatStatus
	pmSamples          []model.PMSample
}

func New() *Store {
	return &Store{
		nes:                make(map[string]model.NetworkElement),
		inventorySnapshots: make(map[string]model.InventorySnapshot),
		cmRequests:         make(map[string]model.CMRequest),
		heartbeats:         make(map[string]model.HeartbeatStatus),
	}
}

func (s *Store) SaveNE(ne model.NetworkElement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nes[ne.ID] = ne
}

func (s *Store) GetNE(id string) (model.NetworkElement, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ne, ok := s.nes[id]
	//ok check exist in map
	return ne, ok
}

func (s *Store) ListNE() []model.NetworkElement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.NetworkElement, 0, len(s.nes))
	for _, ne := range s.nes {
		out = append(out, ne)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *Store) DeleteNE(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, ok := s.nes[id]; ok {
		delete(s.nes, id)
		return true
	}
	return false
}

func (s *Store) SaveInventorySnapshot(snapshot model.InventorySnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inventorySnapshots[snapshot.NEID] = snapshot
}

func (s *Store) GetLatestInventorySnapshot(neID string) (model.InventorySnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.inventorySnapshots[neID]
	return snapshot, ok
}

func (s *Store) SaveCMRequest(req model.CMRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cmRequests[req.ID] = req
}

func (s *Store) GetCMRequest(id string) (model.CMRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.cmRequests[id]
	return req, ok
}

func (s *Store) ListCMRequests() []model.CMRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.CMRequest, 0, len(s.cmRequests))
	for _, req := range s.cmRequests {
		out = append(out, req)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *Store) AddFaultEvent(event model.FaultEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faultEvents = append(s.faultEvents, event)
}

func (s *Store) ListFaultEvents(neID string) []model.FaultEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.FaultEvent, 0, len(s.faultEvents))
	for _, event := range s.faultEvents {
		if neID == "" || event.NEID == neID {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *Store) SaveHeartbeat(hb model.HeartbeatStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heartbeats[hb.NEID] = hb
}

func (s *Store) GetHeartbeat(neID string) (model.HeartbeatStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hb, ok := s.heartbeats[neID]
	return hb, ok
}

func (s *Store) AddPMSample(sample model.PMSample) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pmSamples = append(s.pmSamples, sample)
}

func (s *Store) ListPMSamples(neID, metric string, limit int) []model.PMSample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.PMSample, 0, len(s.pmSamples))
	for _, sample := range s.pmSamples {
		if neID != "" && sample.NEID != neID {
			continue
		}
		if metric != "" && sample.Metric != metric {
			continue
		}
		out = append(out, sample)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CollectedAt.After(out[j].CollectedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
