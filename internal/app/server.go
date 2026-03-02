package app

import (
	"net/http"
	"time"

	"nms_lte/internal/httpapi"
	"nms_lte/internal/service/cm"
	"nms_lte/internal/service/fault"
	"nms_lte/internal/service/inventory"
	"nms_lte/internal/service/ne"
	"nms_lte/internal/service/pm"
	"nms_lte/internal/store/memory"
)

func NewHTTPServer(port string) *http.Server {
	store := memory.New()

	neService := ne.NewService(store)
	inventoryService := inventory.NewService(store)
	cmService := cm.NewService(store)
	faultService := fault.NewService(store)
	pmService := pm.NewService(store)

	handler := httpapi.NewHandler(neService, inventoryService, cmService, faultService, pmService)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
