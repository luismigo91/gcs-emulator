package backend

import (
	"context"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/serviceusage/model"
)

type MemoryServiceUsageBackend struct {
	mu       sync.RWMutex
	services map[string]*model.Service
}

func NewMemoryServiceUsageBackend() *MemoryServiceUsageBackend {
	m := &MemoryServiceUsageBackend{services: make(map[string]*model.Service)}
	for _, name := range []string{
		"storage.googleapis.com", "pubsub.googleapis.com", "secretmanager.googleapis.com",
		"cloudtasks.googleapis.com", "cloudkms.googleapis.com", "logging.googleapis.com",
		"monitoring.googleapis.com", "iam.googleapis.com", "bigquery.googleapis.com",
		"dns.googleapis.com", "cloudfunctions.googleapis.com", "cloudscheduler.googleapis.com",
		"cloudtrace.googleapis.com", "artifactregistry.googleapis.com", "cloudbuild.googleapis.com",
		"billingbudgets.googleapis.com", "servicedirectory.googleapis.com",
		"apigateway.googleapis.com", "certificatemanager.googleapis.com",
		"clouderrorreporting.googleapis.com", "cloudasset.googleapis.com",
		"cloudresourcemanager.googleapis.com", "apikeys.googleapis.com", "serviceusage.googleapis.com",
	} {
		m.services[name] = &model.Service{Name: name, State: "ENABLED"}
	}
	return m
}
func (m *MemoryServiceUsageBackend) Get(ctx context.Context, name string) (*model.Service, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	if s, exists := m.services[name]; exists { return s, nil }
	return &model.Service{Name: name, State: "ENABLED"}, nil
}
func (m *MemoryServiceUsageBackend) List(ctx context.Context) ([]*model.Service, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	result := make([]*model.Service, 0, len(m.services))
	for _, s := range m.services { result = append(result, s) }
	return result, nil
}
func (m *MemoryServiceUsageBackend) Shutdown() error { return nil }
