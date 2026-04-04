package app

import (
	"net/http"
	"time"
	"log"
	
	"nms_lte/internal/httpapi"
	// "nms_lte/internal/service/cm"
	// "nms_lte/internal/service/fault"
	"nms_lte/internal/service/inventory"
	"nms_lte/internal/service/ne"
	// "nms_lte/internal/service/pm"
	// "nms_lte/internal/store/memory"
	"nms_lte/internal/store/postgres"
)

// func NewHTTPServer(port string) *http.Server {
// 	// store := memory.New()

// 	//postgres sql
// 	connStr := "postgres://nms_user:nms_password@localhost:5432/nms_lte?sslmode=disable"
// 	store, err := postgres.New(connStr)
// 	if err != nil {
// 		log.Fatalf("failed to connect to postgres: %v", err)
// 	}

// 	neService := ne.NewServicePG(store)
// 	// inventoryService := inventory.NewService(store)
// 	// cmService := cm.NewService(store)
// 	// faultService := fault.NewService(store)
// 	// pmService := pm.NewService(store)

// 	handler := httpapi.NewHandler(neService, inventoryService, cmService, faultService, pmService)

// 	return &http.Server{
// 		Addr:              ":" + port,
// 		Handler:           handler,
// 		ReadHeaderTimeout: 5 * time.Second,
// 	}
// }

func NewHTTPServer(port string) *http.Server {
	connStr := "postgres://nms_user:nms_password@localhost:5432/nms_lte?sslmode=disable"
	store, err := postgres.New(connStr)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	neService := ne.NewServicePG(store)
	inventoryService := inventory.NewServicePG(store)

	handler := httpapi.NewHandlerPG(neService, inventoryService) // <- используем отдельный HandlerNE для Postgres

	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}