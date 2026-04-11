package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nms_lte/internal/app"
)

//go:embed frontend/dist
var Dist embed.FS

// @title NMS LTE API
// @version 1.0
// @description REST API for managing network elements, inventory, configuration, faults, and performance metrics.
// @BasePath /
// @schemes http

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontendFS, err := fs.Sub(Dist, "frontend/dist")
	if err != nil {
		log.Fatalf("frontend init error: %v", err)
	}

	srv, err := app.NewHTTPServer(port, frontendFS)
	if err != nil {
		log.Fatalf("startup error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("nms-lte started on :%s", port)
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
