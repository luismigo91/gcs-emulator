package backend

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/model"
)

var (
	ErrSecretNotFound      = errors.New("secret not found")
	ErrVersionNotFound     = errors.New("version not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
)

func secretKey(project, name string) string {
	return "projects/" + project + "/secrets/" + name
}

type versionData struct {
	version *model.SecretVersion
	data    []byte
}

type MemorySecretManagerBackend struct {
	mu       sync.RWMutex
	secrets  map[string]*model.Secret
	versions map[string][]*versionData
}

func NewMemorySecretManagerBackend() *MemorySecretManagerBackend {
	return &MemorySecretManagerBackend{
		secrets:  make(map[string]*model.Secret),
		versions: make(map[string][]*versionData),
	}
}

func (m *MemorySecretManagerBackend) CreateSecret(ctx context.Context, secret *model.Secret) (*model.Secret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.secrets[secret.Name]; exists {
		return nil, ErrSecretAlreadyExists
	}
	if secret.Replication == nil {
		secret.Replication = &model.Replication{Automatic: &model.Automatic{}}
	}
	secret.CreateTime = time.Now()
	m.secrets[secret.Name] = secret
	return secret, nil
}

func (m *MemorySecretManagerBackend) GetSecret(ctx context.Context, project, name string) (*model.Secret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := secretKey(project, name)
	s, exists := m.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}
	return s, nil
}

func (m *MemorySecretManagerBackend) ListSecrets(ctx context.Context, project string) ([]*model.Secret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := "projects/" + project + "/"
	var result []*model.Secret
	for name, s := range m.secrets {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MemorySecretManagerBackend) UpdateSecret(ctx context.Context, project, name string, secret *model.Secret, paths []string) (*model.Secret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := secretKey(project, name)
	existing, exists := m.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}
	for _, path := range paths {
		switch path {
		case "labels":
			existing.Labels = secret.Labels
		}
	}
	return existing, nil
}

func (m *MemorySecretManagerBackend) DeleteSecret(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := secretKey(project, name)
	if _, exists := m.secrets[key]; !exists {
		return ErrSecretNotFound
	}
	delete(m.secrets, key)
	delete(m.versions, key)
	return nil
}

func (m *MemorySecretManagerBackend) AddVersion(ctx context.Context, project, name string, payload []byte) (*model.SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := secretKey(project, name)
	if _, exists := m.secrets[key]; !exists {
		return nil, ErrSecretNotFound
	}

	existing := m.versions[key]
	vNum := fmt.Sprintf("%d", len(existing)+1)
	version := &model.SecretVersion{
		Name:       key + "/versions/" + vNum,
		CreateTime: time.Now(),
		State:      "ENABLED",
	}
	m.versions[key] = append(existing, &versionData{version: version, data: payload})
	return version, nil
}

func (m *MemorySecretManagerBackend) ListVersions(ctx context.Context, project, name string) ([]*model.SecretVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := secretKey(project, name)
	if _, exists := m.secrets[key]; !exists {
		return nil, ErrSecretNotFound
	}

	result := make([]*model.SecretVersion, 0, len(m.versions[key]))
	for _, vd := range m.versions[key] {
		if vd.version.State != "DESTROYED" {
			result = append(result, vd.version)
		}
	}
	return result, nil
}

func (m *MemorySecretManagerBackend) GetVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vd, err := m.getVersionLocked(project, name, version)
	if err != nil {
		return nil, err
	}
	return vd.version, nil
}

func (m *MemorySecretManagerBackend) AccessVersion(ctx context.Context, project, name, version string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vd, err := m.getVersionLocked(project, name, version)
	if err != nil {
		return nil, err
	}
	if vd.version.State != "ENABLED" {
		return nil, errors.New("version is not enabled")
	}
	return vd.data, nil
}

func (m *MemorySecretManagerBackend) EnableVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vd, err := m.getVersionLocked(project, name, version)
	if err != nil {
		return nil, err
	}
	vd.version.State = "ENABLED"
	return vd.version, nil
}

func (m *MemorySecretManagerBackend) DisableVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vd, err := m.getVersionLocked(project, name, version)
	if err != nil {
		return nil, err
	}
	vd.version.State = "DISABLED"
	return vd.version, nil
}

func (m *MemorySecretManagerBackend) DestroyVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vd, err := m.getVersionLocked(project, name, version)
	if err != nil {
		return nil, err
	}
	vd.version.State = "DESTROYED"
	vd.version.DestroyTime = time.Now()
	return vd.version, nil
}

func (m *MemorySecretManagerBackend) getVersionLocked(project, name, version string) (*versionData, error) {
	key := secretKey(project, name)
	versions := m.versions[key]
	for _, vd := range versions {
		if vd.version.Name == key+"/versions/"+version || vd.version.Name == version {
			return vd, nil
		}
	}
	return nil, ErrVersionNotFound
}

func (m *MemorySecretManagerBackend) Shutdown() error {
	return nil
}

var _ = base64.StdEncoding
