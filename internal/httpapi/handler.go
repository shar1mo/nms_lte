package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"nms_lte/internal/service/cm"
	"nms_lte/internal/service/fault"
	"nms_lte/internal/service/inventory"
	"nms_lte/internal/service/ne"
	"nms_lte/internal/service/pm"
)

type Handler struct {
	neService        *ne.Service
	inventoryService *inventory.Service
	cmService        *cm.Service
	faultService     *fault.Service
	pmService        *pm.Service
}

func NewHandler(
	neService *ne.Service,
	inventoryService *inventory.Service,
	cmService *cm.Service,
	faultService *fault.Service,
	pmService *pm.Service,
) http.Handler {
	h := &Handler{
		neService:        neService,
		inventoryService: inventoryService,
		cmService:        cmService,
		faultService:     faultService,
		pmService:        pmService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/api/v1/ne", h.handleNECollection)
	mux.HandleFunc("/api/v1/ne/", h.handleNEDetails)
	mux.HandleFunc("/api/v1/cm/requests", h.handleCMRequests)
	mux.HandleFunc("/api/v1/fault/events", h.handleFaultEvents)
	mux.HandleFunc("/api/v1/pm/samples", h.handlePMSamples)

	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleNECollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.neService.List())
	case http.MethodPost:
		var req struct {
			Name         string   `json:"name"`
			Address      string   `json:"address"`
			Vendor       string   `json:"vendor"`
			Capabilities []string `json:"capabilities"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		neItem, err := h.neService.Register(req.Name, req.Address, req.Vendor, req.Capabilities)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, neItem)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleNEDetails(w http.ResponseWriter, r *http.Request) {
	segments := splitNESubPath(r.URL.Path)
	if len(segments) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	neID := segments[0]
	if len(segments) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		item, ok := h.neService.Get(neID)
		if !ok {
			writeError(w, http.StatusNotFound, "network element not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}

	if len(segments) == 3 && segments[1] == "inventory" && segments[2] == "sync" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		snapshot, err := h.inventoryService.Sync(neID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
		return
	}

	if len(segments) == 3 && segments[1] == "inventory" && segments[2] == "latest" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		snapshot, ok := h.inventoryService.GetLatest(neID)
		if !ok {
			writeError(w, http.StatusNotFound, "inventory snapshot not found")
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
		return
	}

	if len(segments) == 3 && segments[1] == "heartbeat" && segments[2] == "check" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req struct {
			Healthy *bool `json:"healthy"`
		}
		if r.ContentLength > 0 {
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		healthy := true
		if req.Healthy != nil {
			healthy = *req.Healthy
		}
		hb, err := h.faultService.CheckHeartbeat(neID, healthy)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, hb)
		return
	}

	if len(segments) == 3 && segments[1] == "heartbeat" && segments[2] == "latest" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		hb, ok := h.faultService.GetHeartbeat(neID)
		if !ok {
			writeError(w, http.StatusNotFound, "heartbeat not found")
			return
		}
		writeJSON(w, http.StatusOK, hb)
		return
	}

	if len(segments) == 3 && segments[1] == "pm" && segments[2] == "collect" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req struct {
			Metric string `json:"metric"`
		}
		if r.ContentLength > 0 {
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		sample, err := h.pmService.Collect(neID, req.Metric)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sample)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleCMRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.cmService.ListRequests())
	case http.MethodPost:
		var req cm.ApplyChangeInput
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		item, err := h.cmService.ApplyChange(req)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if item.Status == "failed" {
			writeJSON(w, http.StatusConflict, item)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleFaultEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		neID := r.URL.Query().Get("ne_id")
		writeJSON(w, http.StatusOK, h.faultService.ListEvents(neID))
	case http.MethodPost:
		var req struct {
			NEID     string `json:"ne_id"`
			Severity string `json:"severity"`
			Message  string `json:"message"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		event, err := h.faultService.ReportEvent(req.NEID, req.Severity, req.Message)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, event)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handlePMSamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	neID := r.URL.Query().Get("ne_id")
	metric := r.URL.Query().Get("metric")
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be positive integer")
			return
		}
		limit = parsed
	}
	writeJSON(w, http.StatusOK, h.pmService.List(neID, metric, limit))
}

func splitNESubPath(path string) []string {
	base := strings.TrimPrefix(path, "/api/v1/ne/")
	base = strings.Trim(base, "/")
	if base == "" {
		return nil
	}
	return strings.Split(base, "/")
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
