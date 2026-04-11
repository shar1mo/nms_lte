package app

import (
	"io/fs"
	"net/http"
	"os"
	"time"
	"errors"

	"nms_lte/internal/httpapi"
	"nms_lte/internal/service/cm"
	"nms_lte/internal/service/fault"
	"nms_lte/internal/service/inventory"
	"nms_lte/internal/service/ne"
	"nms_lte/internal/service/pm"
	"nms_lte/internal/store/postgres"
)

func NewHTTPServer(port string, frontendFS fs.FS) (*http.Server, error) {
	// store := memory.New()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	store, err := postgres.New(os.Getenv("DATABASE_URL"))
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

	handler := httpapi.NewHandler(neService, inventoryService, cmService, faultService, pmService, frontendFS)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}
