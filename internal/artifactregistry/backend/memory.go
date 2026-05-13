package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/artifactregistry/model"
)

var ErrRepoNotFound = errors.New("repository not found")

type MemoryArtifactRegistryBackend struct {
	mu     sync.RWMutex
	repos  map[string]*model.Repository
	images map[string][]*model.DockerImage
}

func NewMemoryArtifactRegistryBackend() *MemoryArtifactRegistryBackend {
	return &MemoryArtifactRegistryBackend{
		repos:  make(map[string]*model.Repository),
		images: make(map[string][]*model.DockerImage),
	}
}

func (m *MemoryArtifactRegistryBackend) CreateRepository(ctx context.Context, repo *model.Repository) (*model.Repository, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.repos[repo.Name]; exists {
		return nil, errors.New("repository already exists")
	}
	m.repos[repo.Name] = repo
	return repo, nil
}

func (m *MemoryArtifactRegistryBackend) GetRepository(ctx context.Context, name string) (*model.Repository, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	r, exists := m.repos[name]
	if !exists { return nil, ErrRepoNotFound }
	return r, nil
}

func (m *MemoryArtifactRegistryBackend) ListRepositories(ctx context.Context, parent string) ([]*model.Repository, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Repository
	for n, r := range m.repos {
		if strings.HasPrefix(n, parent+"/repositories/") {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemoryArtifactRegistryBackend) DeleteRepository(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.repos[name]; !exists { return ErrRepoNotFound }
	delete(m.repos, name)
	return nil
}

func (m *MemoryArtifactRegistryBackend) ListDockerImages(ctx context.Context, parent string) ([]*model.DockerImage, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	return m.images[parent], nil
}

func (m *MemoryArtifactRegistryBackend) Shutdown() error { return nil }
