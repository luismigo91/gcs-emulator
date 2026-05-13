package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/servicedirectory/model"
)

var ErrNotFound = errors.New("not found")

type MemoryServiceDirectoryBackend struct {
	mu         sync.RWMutex
	namespaces map[string]*model.Namespace
	services   map[string][]*model.Service
}

func NewMemoryServiceDirectoryBackend() *MemoryServiceDirectoryBackend {
	return &MemoryServiceDirectoryBackend{
		namespaces: make(map[string]*model.Namespace),
		services:   make(map[string][]*model.Service),
	}
}

func nsKey(project, location, ns string) string { return project + "/" + location + "/" + ns }
func svcKey(project, location, ns string) string { return project + "/" + location + "/" + ns + "/services" }

func (m *MemoryServiceDirectoryBackend) CreateNamespace(ctx context.Context, project, location string, ns *model.Namespace) (*model.Namespace, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	k := nsKey(project, location, ns.Name)
	if _, exists := m.namespaces[k]; exists { return nil, errors.New("exists") }
	m.namespaces[k] = ns
	return ns, nil
}
func (m *MemoryServiceDirectoryBackend) GetNamespace(ctx context.Context, k string) (*model.Namespace, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	ns, exists := m.namespaces[k]
	if !exists { return nil, ErrNotFound }
	return ns, nil
}
func (m *MemoryServiceDirectoryBackend) ListNamespaces(ctx context.Context, project, location string) ([]*model.Namespace, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	prefix := project + "/" + location + "/"
	var result []*model.Namespace
	for n, ns := range m.namespaces {
		if strings.HasPrefix(n, prefix) { result = append(result, ns) }
	}
	return result, nil
}
func (m *MemoryServiceDirectoryBackend) DeleteNamespace(ctx context.Context, k string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.namespaces, k)
	delete(m.services, k+"/services")
	return nil
}
func (m *MemoryServiceDirectoryBackend) CreateService(ctx context.Context, project, location, ns string, svc *model.Service) (*model.Service, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	k := svcKey(project, location, ns)
	if _, exists := m.namespaces[nsKey(project, location, ns)]; !exists { return nil, ErrNotFound }
	m.services[k] = append(m.services[k], svc)
	return svc, nil
}
func (m *MemoryServiceDirectoryBackend) ListServices(ctx context.Context, project, location, ns string) ([]*model.Service, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	return m.services[svcKey(project, location, ns)], nil
}
func (m *MemoryServiceDirectoryBackend) Shutdown() error { return nil }
