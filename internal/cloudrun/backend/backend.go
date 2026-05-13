package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrun/model"
)

var ErrNotFound = errors.New("not found")

type MemoryCloudRunBackend struct {
	mu       sync.RWMutex
	services map[string]*model.Service
}

func NewMemoryCloudRunBackend() *MemoryCloudRunBackend {
	return &MemoryCloudRunBackend{services: make(map[string]*model.Service)}
}
func (m *MemoryCloudRunBackend) Create(ctx context.Context, s *model.Service) (*model.Service, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.services[s.Name]; exists { return nil, errors.New("exists") }
	s.Status = "RUNNING"
	m.services[s.Name] = s
	return s, nil
}
func (m *MemoryCloudRunBackend) Get(ctx context.Context, name string) (*model.Service, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	s, exists := m.services[name]
	if !exists { return nil, ErrNotFound }
	return s, nil
}
func (m *MemoryCloudRunBackend) List(ctx context.Context, parent string) ([]*model.Service, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Service
	for n, s := range m.services {
		if strings.HasPrefix(n, parent+"/") { result = append(result, s) }
	}
	return result, nil
}
func (m *MemoryCloudRunBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.services, name)
	return nil
}
func (m *MemoryCloudRunBackend) Shutdown() error { return nil }
