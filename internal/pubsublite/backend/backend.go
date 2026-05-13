package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsublite/model"
)

var ErrNotFound = errors.New("not found")

type MemoryPubSubLiteBackend struct {
	mu     sync.RWMutex
	topics map[string]*model.LiteTopic
}

func NewMemoryPubSubLiteBackend() *MemoryPubSubLiteBackend {
	return &MemoryPubSubLiteBackend{topics: make(map[string]*model.LiteTopic)}
}
func (m *MemoryPubSubLiteBackend) Create(ctx context.Context, t *model.LiteTopic) (*model.LiteTopic, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.topics[t.Name]; exists { return nil, errors.New("exists") }
	m.topics[t.Name] = t
	return t, nil
}
func (m *MemoryPubSubLiteBackend) Get(ctx context.Context, name string) (*model.LiteTopic, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	t, exists := m.topics[name]
	if !exists { return nil, ErrNotFound }
	return t, nil
}
func (m *MemoryPubSubLiteBackend) List(ctx context.Context, parent string) ([]*model.LiteTopic, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.LiteTopic
	for n, t := range m.topics {
		if strings.HasPrefix(n, parent+"/") { result = append(result, t) }
	}
	return result, nil
}
func (m *MemoryPubSubLiteBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.topics, name)
	return nil
}
func (m *MemoryPubSubLiteBackend) Shutdown() error { return nil }
