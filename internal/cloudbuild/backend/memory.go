package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudbuild/model"
)

var ErrTriggerNotFound = errors.New("trigger not found")

type MemoryCloudBuildBackend struct {
	mu       sync.RWMutex
	triggers map[string]*model.BuildTrigger
	builds   []*model.Build
	buildID  int64
}

func NewMemoryCloudBuildBackend() *MemoryCloudBuildBackend {
	return &MemoryCloudBuildBackend{triggers: make(map[string]*model.BuildTrigger)}
}

func (m *MemoryCloudBuildBackend) CreateTrigger(ctx context.Context, trigger *model.BuildTrigger) (*model.BuildTrigger, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.triggers[trigger.Name]; exists {
		return nil, errors.New("trigger already exists")
	}
	m.triggers[trigger.Name] = trigger
	return trigger, nil
}
func (m *MemoryCloudBuildBackend) GetTrigger(ctx context.Context, name string) (*model.BuildTrigger, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	t, exists := m.triggers[name]
	if !exists { return nil, ErrTriggerNotFound }
	return t, nil
}
func (m *MemoryCloudBuildBackend) ListTriggers(ctx context.Context, project string) ([]*model.BuildTrigger, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.BuildTrigger
	prefix := "projects/" + project + "/"
	for n, t := range m.triggers {
		if strings.HasPrefix(n, prefix) { result = append(result, t) }
	}
	return result, nil
}
func (m *MemoryCloudBuildBackend) DeleteTrigger(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.triggers[name]; !exists { return ErrTriggerNotFound }
	delete(m.triggers, name)
	return nil
}
func (m *MemoryCloudBuildBackend) RunTrigger(ctx context.Context, name string) (*model.Build, error) {
	m.mu.Lock()
	m.buildID++
	b := &model.Build{ID: "build-" + name, Status: "SUCCESS", TriggerID: name}
	m.builds = append(m.builds, b)
	m.mu.Unlock()
	return b, nil
}
func (m *MemoryCloudBuildBackend) Shutdown() error { return nil }
