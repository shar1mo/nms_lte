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
	"nms_lte/internal/store/postgres"
)

func NewHTTPServer(port string) (*http.Server, error) {
	// store := memory.New()

	store, err := postgres.New("")
	if err != nil {
		return nil, err
	}

	neService, err := ne.NewManagedService(store)
	if err != nil {
		return nil, err
	}

	inventoryService := inventory.NewService(store, neService)
	cmService := cm.NewService(store)
	faultService := fault.NewService(store)
	pmService := pm.NewService(store)

	handler := httpapi.NewHandler(neService, inventoryService, cmService, faultService, pmService)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}
