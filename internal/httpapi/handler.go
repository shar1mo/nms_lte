package httpapi

import (
	"encoding/json"
	"errors"
	"io/fs"
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

type HandlerPG struct {
	neService        *ne.Service
	inventoryService *inventory.ServicePG
}

func NewHandler(
	neService *ne.Service,
	inventoryService *inventory.Service,
	cmService *cm.Service,
	faultService *fault.Service,
	pmService *pm.Service,
	frontendFS fs.FS,
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
	registerSwaggerRoutes(mux)
	mux.HandleFunc("/api/v1/ne", h.handleNECollection)
	mux.HandleFunc("/api/v1/ne/", h.handleNEDetails)
	mux.HandleFunc("/api/v1/cm/requests", h.handleCMRequests)
	mux.HandleFunc("/api/v1/fault/events", h.handleFaultEvents)
	mux.HandleFunc("/api/v1/pm/samples", h.handlePMSamples)
	if frontendFS != nil {
		mux.Handle("/", newFrontendHandler(frontendFS))
	}

	return mux
}

func NewHandlerPG(neService *ne.Service, inventoryService *inventory.ServicePG) http.Handler {
	h := &HandlerPG{
		neService:        neService,
		inventoryService: inventoryService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	registerSwaggerRoutes(mux)
	mux.HandleFunc("/api/v1/ne", h.handleNECollection)
	mux.HandleFunc("/api/v1/ne/", h.handleNEDetails)

	return mux
}

func (h *HandlerPG) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HandlerPG) handleNECollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := h.neService.ListPG()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, list)
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
		// POST /api/v1/ne
		neItem, err := h.neService.RegisterPG(req.Name, req.Address, req.Vendor, req.Capabilities)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, neItem)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *HandlerPG) handleNEDetails(w http.ResponseWriter, r *http.Request) {
	segments := splitNESubPath(r.URL.Path)
	if len(segments) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	neID := segments[0]

	// GET /api/v1/ne/{id}
	if len(segments) == 1 && r.Method == http.MethodGet {
		item, ok := h.neService.GetPG(neID)
		if !ok {
			writeError(w, http.StatusNotFound, "network element not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}

	// DELETE /api/v1/ne/{id}
	if len(segments) == 1 && r.Method == http.MethodDelete {
		if err := h.neService.UnRegisterPG(neID); err != nil {
			if errors.Is(err, ne.ErrNENotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//inventory
	//POST /api/v1/ne/{id}/inventory/sync
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

	//GET /api/v1/ne/{id}/inventory/latest
	if len(segments) == 3 && segments[1] == "inventory" && segments[2] == "latest" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		snapshot, err := h.inventoryService.GetLatestPG(neID)
		if err != nil {
			writeError(w, http.StatusNotFound, "inventory snapshot not found")
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

// handleHealth godoc
// @Summary Health check
// @Description Returns current API health status.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleNECollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleNEList(w, r)
	case http.MethodPost:
		h.handleNECreate(w, r)
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

	// DELETE /api/v1/ne/{id}
	if r.Method == http.MethodDelete && len(segments) == 1 {
		h.handleNEDelete(w, r)
		return
	}

	// GET /api/v1/ne/{id}
	if len(segments) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleNEGet(w, r)
		return
	}

	if len(segments) == 3 && segments[1] == "inventory" && segments[2] == "sync" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleInventorySync(w, r)
		return
	}

	if len(segments) == 3 && segments[1] == "inventory" && segments[2] == "latest" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleInventoryLatest(w, r)
		return
	}

	if len(segments) == 3 && segments[1] == "heartbeat" && segments[2] == "check" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleHeartbeatCheck(w, r)
		return
	}

	if len(segments) == 3 && segments[1] == "heartbeat" && segments[2] == "latest" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleHeartbeatLatest(w, r)
		return
	}

	if len(segments) == 3 && segments[1] == "pm" && segments[2] == "collect" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handlePMCollect(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleCMRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleCMRequestList(w, r)
	case http.MethodPost:
		h.handleCMRequestCreate(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleFaultEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleFaultEventList(w, r)
	case http.MethodPost:
		h.handleFaultEventCreate(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleNEList godoc
// @Summary List network elements
// @Description Returns all registered network elements.
// @Tags network-elements
// @Produce json
// @Success 200 {array} NetworkElement
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ne [get]
func (h *Handler) handleNEList(w http.ResponseWriter, _ *http.Request) {
	nelist, err := h.neService.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nelist)
}

// handleNECreate godoc
// @Summary Register network element
// @Description Creates a new managed network element.
// @Tags network-elements
// @Accept json
// @Produce json
// @Param request body RegisterNERequest true "Network element payload"
// @Success 201 {object} NetworkElement
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/ne [post]
func (h *Handler) handleNECreate(w http.ResponseWriter, r *http.Request) {
	var req RegisterNERequest
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
}

// handleNEGet godoc
// @Summary Get network element
// @Description Returns a single network element by ID.
// @Tags network-elements
// @Produce json
// @Param id path string true "Network element ID"
// @Success 200 {object} NetworkElement
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id} [get]
func (h *Handler) handleNEGet(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	item, ok, err := h.neService.Get(neID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "network element not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// handleNEDelete godoc
// @Summary Delete network element
// @Description Removes a network element by ID.
// @Tags network-elements
// @Produce json
// @Param id path string true "Network element ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id} [delete]
func (h *Handler) handleNEDelete(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	if err := h.neService.UnRegister(neID); err != nil {
		if errors.Is(err, ne.ErrNENotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleInventorySync godoc
// @Summary Sync inventory
// @Description Triggers inventory synchronization for a network element.
// @Tags inventory
// @Produce json
// @Param id path string true "Network element ID"
// @Success 200 {object} InventorySnapshot
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id}/inventory/sync [post]
func (h *Handler) handleInventorySync(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	snapshot, err := h.inventoryService.Sync(neID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleInventoryLatest godoc
// @Summary Get latest inventory snapshot
// @Description Returns the latest inventory snapshot for a network element.
// @Tags inventory
// @Produce json
// @Param id path string true "Network element ID"
// @Success 200 {object} InventorySnapshot
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id}/inventory/latest [get]
func (h *Handler) handleInventoryLatest(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	snapshot, err := h.inventoryService.GetLatest(neID)
	if err != nil {
		writeError(w, http.StatusNotFound, "inventory snapshot not found")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleHeartbeatCheck godoc
// @Summary Check heartbeat
// @Description Records a heartbeat result for a network element.
// @Tags heartbeat
// @Accept json
// @Produce json
// @Param id path string true "Network element ID"
// @Param request body CheckHeartbeatRequest false "Heartbeat payload"
// @Success 200 {object} HeartbeatStatus
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id}/heartbeat/check [post]
func (h *Handler) handleHeartbeatCheck(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	var req CheckHeartbeatRequest
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
}

// handleHeartbeatLatest godoc
// @Summary Get latest heartbeat
// @Description Returns the latest heartbeat status for a network element.
// @Tags heartbeat
// @Produce json
// @Param id path string true "Network element ID"
// @Success 200 {object} HeartbeatStatus
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id}/heartbeat/latest [get]
func (h *Handler) handleHeartbeatLatest(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	hb, ok := h.faultService.GetHeartbeat(neID)
	if !ok {
		writeError(w, http.StatusNotFound, "heartbeat not found")
		return
	}
	writeJSON(w, http.StatusOK, hb)
}

// handlePMCollect godoc
// @Summary Collect performance sample
// @Description Collects a performance metric for a network element.
// @Tags performance
// @Accept json
// @Produce json
// @Param id path string true "Network element ID"
// @Param request body CollectPMSampleRequest false "Performance collection payload"
// @Success 200 {object} PMSample
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ne/{id}/pm/collect [post]
func (h *Handler) handlePMCollect(w http.ResponseWriter, r *http.Request) {
	neID := neIDFromRequest(r)
	var req CollectPMSampleRequest
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
}

// handleCMRequestList godoc
// @Summary List configuration requests
// @Description Returns configuration change requests history.
// @Tags configuration-management
// @Produce json
// @Success 200 {array} CMRequest
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/cm/requests [get]
func (h *Handler) handleCMRequestList(w http.ResponseWriter, _ *http.Request) {
	req, err := h.cmService.ListRequests()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, req)
}

// handleCMRequestCreate godoc
// @Summary Apply configuration change
// @Description Creates a configuration management request for a network element.
// @Tags configuration-management
// @Accept json
// @Produce json
// @Param request body ApplyChangeRequest true "Configuration change payload"
// @Success 201 {object} CMRequest
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} CMRequest
// @Router /api/v1/cm/requests [post]
func (h *Handler) handleCMRequestCreate(w http.ResponseWriter, r *http.Request) {
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
}

// handleFaultEventList godoc
// @Summary List fault events
// @Description Returns fault events, optionally filtered by network element ID.
// @Tags fault-management
// @Produce json
// @Param ne_id query string false "Network element ID"
// @Success 200 {array} FaultEvent
// @Router /api/v1/fault/events [get]
func (h *Handler) handleFaultEventList(w http.ResponseWriter, r *http.Request) {
	neID := r.URL.Query().Get("ne_id")
	writeJSON(w, http.StatusOK, h.faultService.ListEvents(neID))
}

// handleFaultEventCreate godoc
// @Summary Report fault event
// @Description Creates a new fault event for a network element.
// @Tags fault-management
// @Accept json
// @Produce json
// @Param request body CreateFaultEventRequest true "Fault event payload"
// @Success 201 {object} FaultEvent
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/fault/events [post]
func (h *Handler) handleFaultEventCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateFaultEventRequest
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
}

// handlePMSamples godoc
// @Summary List performance samples
// @Description Returns collected performance samples with optional filters.
// @Tags performance
// @Produce json
// @Param ne_id query string false "Network element ID"
// @Param metric query string false "Metric name"
// @Param limit query int false "Maximum number of samples" default(100)
// @Success 200 {array} PMSample
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/pm/samples [get]
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

func neIDFromRequest(r *http.Request) string {
	segments := splitNESubPath(r.URL.Path)
	if len(segments) == 0 {
		return ""
	}
	return segments[0]
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

func newFrontendHandler(frontendFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(frontendFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		requestedPath := strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if requestedPath == "" {
			requestedPath = "index.html"
		}

		if frontendFileExists(frontendFS, requestedPath) {
			fileServer.ServeHTTP(w, r)
			return
		}

		indexRequest := r.Clone(r.Context())
		indexURL := *r.URL
		indexURL.Path = "/index.html"
		indexRequest.URL = &indexURL
		fileServer.ServeHTTP(w, indexRequest)
	})
}

func frontendFileExists(frontendFS fs.FS, name string) bool {
	file, err := frontendFS.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return !info.IsDir()
}
