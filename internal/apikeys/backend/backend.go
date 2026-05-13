package backend

import (
	"context"
	"errors"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/apikeys/model"
)

var ErrNotFound = errors.New("key not found")

type MemoryAPIKeysBackend struct {
	mu   sync.RWMutex
	keys map[string]*model.Key
}

func NewMemoryAPIKeysBackend() *MemoryAPIKeysBackend {
	return &MemoryAPIKeysBackend{keys: make(map[string]*model.Key)}
}
func (m *MemoryAPIKeysBackend) Create(ctx context.Context, k *model.Key) (*model.Key, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	k.KeyString = "AIza" + k.DisplayName + "-emulator-key"
	m.keys[k.Name] = k
	return k, nil
}
func (m *MemoryAPIKeysBackend) Get(ctx context.Context, name string) (*model.Key, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	k, exists := m.keys[name]
	if !exists { return nil, ErrNotFound }
	return k, nil
}
func (m *MemoryAPIKeysBackend) List(ctx context.Context) ([]*model.Key, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	result := make([]*model.Key, 0, len(m.keys))
	for _, k := range m.keys { result = append(result, k) }
	return result, nil
}
func (m *MemoryAPIKeysBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.keys, name)
	return nil
}
func (m *MemoryAPIKeysBackend) Shutdown() error { return nil }
