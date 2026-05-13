package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/config"
	kms "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	tasks "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	emulatorrouter "github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func main() {
	cfg := config.Load()

	b, err := backend.NewBackend(backend.BackendConfig{
		Mode:        cfg.StorageMode,
		StoragePath: cfg.StoragePath,
	})
	if err != nil {
		log.Fatalf("Failed to create backend: %v", err)
	}

	seedPath := os.Getenv("GCP_EMULATOR_SEED_PATH")
	if seedPath == "" {
		seedPath = "/data"
	}
	if err := backend.SeedFromDirectory(b, seedPath); err != nil {
		log.Printf("Warning: failed to seed from %s: %v", seedPath, err)
	}

	pubsubMode := os.Getenv("GCP_EMULATOR_PUBSUB_MODE")
	if pubsubMode == "" {
		pubsubMode = "memory"
	}
	pubsubPath := os.Getenv("GCP_EMULATOR_PUBSUB_PATH")
	if pubsubPath == "" {
		pubsubPath = cfg.StoragePath + "/pubsub"
	}
	psb, err := pubsub.NewPubSubBackend(pubsubMode, pubsubPath)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub backend: %v", err)
	}

	smb := secretmanager.NewMemorySecretManagerBackend()

	ctb := tasks.NewMemoryCloudTasksBackend()

	kmb := kms.NewMemoryKMSBackend()

	mux := emulatorrouter.New(b, psb, smb, ctb, kmb, cfg.DefaultProject)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting GCS Emulator on port %d", cfg.Port)
		log.Printf("Storage mode: %s", cfg.StorageMode)
		log.Printf("Default project: %s", cfg.DefaultProject)
		log.Printf("Storage path: %s", cfg.StoragePath)
		log.Printf("Seed path: %s", seedPath)
		log.Printf("Pub/Sub mode: %s", pubsubMode)
		log.Printf("Services: gcs, pubsub, secretmanager, cloudtasks, kms")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := b.Shutdown(); err != nil {
		log.Printf("Backend shutdown error: %v", err)
	}

	if err := psb.Shutdown(); err != nil {
		log.Printf("Pub/Sub backend shutdown error: %v", err)
	}

	log.Println("Server exited")
}
