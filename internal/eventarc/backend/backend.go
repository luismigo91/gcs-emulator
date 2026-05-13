package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/eventarc/model"
)

var ErrNotFound = errors.New("trigger not found")

type MemoryEventarcBackend struct {
	mu       sync.RWMutex
	triggers map[string]*model.Trigger
}

func NewMemoryEventarcBackend() *MemoryEventarcBackend {
	return &MemoryEventarcBackend{triggers: make(map[string]*model.Trigger)}
}

func (m *MemoryEventarcBackend) Create(ctx context.Context, t *model.Trigger) (*model.Trigger, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.triggers[t.Name]; exists { return nil, errors.New("exists") }
	m.triggers[t.Name] = t
	return t, nil
}
func (m *MemoryEventarcBackend) Get(ctx context.Context, name string) (*model.Trigger, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	t, exists := m.triggers[name]
	if !exists { return nil, ErrNotFound }
	return t, nil
}
func (m *MemoryEventarcBackend) List(ctx context.Context, parent string) ([]*model.Trigger, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Trigger
	for n, t := range m.triggers {
		if strings.HasPrefix(n, parent+"/") { result = append(result, t) }
	}
	return result, nil
}
func (m *MemoryEventarcBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.triggers, name)
	return nil
}
func (m *MemoryEventarcBackend) Shutdown() error { return nil }
