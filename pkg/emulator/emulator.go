package emulator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/config"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	emulatorrouter "github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

type Server struct {
	config        *config.Config
	backend       backend.Backend
	pubsub        pubsub.PubSubBackend
	secretmanager secretmanager.SecretManagerBackend
	server        *http.Server
	testSrv       *httptest.Server
	mu            sync.Mutex
	running       bool
}

type Config struct {
	Port           int
	DefaultProject string
	StorageMode    string
	StoragePath    string
	SeedPath       string
	PubSubMode     string
}

func NewServer(cfg Config) (*Server, error) {
	c := &config.Config{
		Port:           cfg.Port,
		DefaultProject: cfg.DefaultProject,
		StorageMode:    backend.BackendMode(cfg.StorageMode),
		StoragePath:    cfg.StoragePath,
	}

	if c.Port == 0 {
		c.Port = 9090
	}
	if c.DefaultProject == "" {
		c.DefaultProject = "test-project"
	}
	if c.StorageMode == "" {
		c.StorageMode = backend.ModeMemory
	}

	b, err := backend.NewBackend(backend.BackendConfig{
		Mode:        c.StorageMode,
		StoragePath: c.StoragePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	pubsubMode := cfg.PubSubMode
	if pubsubMode == "" {
		pubsubMode = "memory"
	}
	pubsubPath := cfg.StoragePath + "/pubsub"
	psb, err := pubsub.NewPubSubBackend(pubsubMode, pubsubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub backend: %w", err)
	}

	return &Server{
		config:        c,
		backend:       b,
		pubsub:        psb,
		secretmanager: secretmanager.NewMemorySecretManagerBackend(),
	}, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	mux := emulatorrouter.New(s.backend, s.pubsub, s.secretmanager, nil, nil, s.config.DefaultProject)

	addr := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.running = true

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	if err := s.backend.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown backend: %w", err)
	}

	if err := s.pubsub.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown pubsub backend: %w", err)
	}

	s.running = false
	return nil
}

func (s *Server) URL() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.testSrv != nil {
		return s.testSrv.URL
	}

	if s.server != nil {
		return fmt.Sprintf("http://localhost:%d", s.config.Port)
	}

	return ""
}

func (s *Server) StartTestServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	mux := emulatorrouter.New(s.backend, s.pubsub, s.secretmanager, nil, nil, s.config.DefaultProject)

	s.testSrv = httptest.NewServer(mux)
	s.running = true

	return nil
}

func (s *Server) StopTestServer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.testSrv != nil {
		s.testSrv.Close()
		s.testSrv = nil
	}
	s.running = false
}
