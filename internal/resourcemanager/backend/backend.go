package backend

import (
	"context"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/resourcemanager/model"
)

type MemoryResourceManagerBackend struct {
	mu       sync.RWMutex
	projects map[string]*model.Project
}

func NewMemoryResourceManagerBackend() *MemoryResourceManagerBackend {
	return &MemoryResourceManagerBackend{projects: make(map[string]*model.Project)}
}

func (m *MemoryResourceManagerBackend) Create(ctx context.Context, p *model.Project) (*model.Project, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	p.LifecycleState = "ACTIVE"
	m.projects[p.ProjectID] = p
	return p, nil
}
func (m *MemoryResourceManagerBackend) Get(ctx context.Context, id string) (*model.Project, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	p, exists := m.projects[id]
	if !exists { return &model.Project{ProjectID: id, LifecycleState: "ACTIVE"}, nil }
	return p, nil
}
func (m *MemoryResourceManagerBackend) List(ctx context.Context) ([]*model.Project, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	result := make([]*model.Project, 0, len(m.projects))
	for _, p := range m.projects { result = append(result, p) }
	return result, nil
}
func (m *MemoryResourceManagerBackend) Delete(ctx context.Context, id string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.projects, id)
	return nil
}
func (m *MemoryResourceManagerBackend) Shutdown() error { return nil }

var _ = strings.TrimPrefix
