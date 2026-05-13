package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cdn/model"
)

var ErrNotFound = errors.New("not found")

type MemoryCDNBackend struct {
	mu       sync.RWMutex
	backends map[string]*model.BackendService
	urlMaps  map[string]*model.URLMap
}

func NewMemoryCDNBackend() *MemoryCDNBackend {
	return &MemoryCDNBackend{
		backends: make(map[string]*model.BackendService),
		urlMaps:  make(map[string]*model.URLMap),
	}
}

func (m *MemoryCDNBackend) CreateBackend(ctx context.Context, b *model.BackendService) (*model.BackendService, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.backends[b.Name]; exists { return nil, errors.New("exists") }
	m.backends[b.Name] = b
	return b, nil
}
func (m *MemoryCDNBackend) GetBackend(ctx context.Context, name string) (*model.BackendService, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	b, exists := m.backends[name]
	if !exists { return nil, ErrNotFound }
	return b, nil
}
func (m *MemoryCDNBackend) ListBackends(ctx context.Context, project string) ([]*model.BackendService, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	prefix := "projects/" + project + "/"
	var result []*model.BackendService
	for n, b := range m.backends {
		if strings.HasPrefix(n, prefix) { result = append(result, b) }
	}
	return result, nil
}
func (m *MemoryCDNBackend) DeleteBackend(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.backends, name)
	return nil
}
func (m *MemoryCDNBackend) CreateURLMap(ctx context.Context, um *model.URLMap) (*model.URLMap, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.urlMaps[um.Name]; exists { return nil, errors.New("exists") }
	m.urlMaps[um.Name] = um
	return um, nil
}
func (m *MemoryCDNBackend) GetURLMap(ctx context.Context, name string) (*model.URLMap, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	um, exists := m.urlMaps[name]
	if !exists { return nil, ErrNotFound }
	return um, nil
}
func (m *MemoryCDNBackend) ListURLMaps(ctx context.Context, project string) ([]*model.URLMap, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	prefix := "projects/" + project + "/"
	var result []*model.URLMap
	for n, um := range m.urlMaps {
		if strings.HasPrefix(n, prefix) { result = append(result, um) }
	}
	return result, nil
}
func (m *MemoryCDNBackend) DeleteURLMap(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.urlMaps, name)
	return nil
}
func (m *MemoryCDNBackend) Shutdown() error { return nil }
