package backend

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/iam/model"
)

var ErrServiceAccountNotFound = errors.New("service account not found")

type MemoryIAMBackend struct {
	mu                sync.RWMutex
	serviceAccounts   map[string]*model.ServiceAccount
}

func NewMemoryIAMBackend() *MemoryIAMBackend {
	return &MemoryIAMBackend{
		serviceAccounts: make(map[string]*model.ServiceAccount),
	}
}

func (m *MemoryIAMBackend) CreateServiceAccount(ctx context.Context, project string, sa *model.ServiceAccount) (*model.ServiceAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sa.Name == "" {
		sa.Name = fmt.Sprintf("projects/%s/serviceAccounts/%s", project, sa.Email)
	}
	if _, exists := m.serviceAccounts[sa.Name]; exists {
		return nil, errors.New("service account already exists")
	}
	m.serviceAccounts[sa.Name] = sa
	return sa, nil
}

func (m *MemoryIAMBackend) GetServiceAccount(ctx context.Context, name string) (*model.ServiceAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sa, exists := m.serviceAccounts[name]
	if !exists {
		return nil, ErrServiceAccountNotFound
	}
	return sa, nil
}

func (m *MemoryIAMBackend) ListServiceAccounts(ctx context.Context, project string) ([]*model.ServiceAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefix := "projects/" + project + "/"
	var result []*model.ServiceAccount
	for n, sa := range m.serviceAccounts {
		if len(n) > len(prefix) && n[:len(prefix)] == prefix {
			result = append(result, sa)
		}
	}
	return result, nil
}

func (m *MemoryIAMBackend) DeleteServiceAccount(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.serviceAccounts[name]; !exists {
		return ErrServiceAccountNotFound
	}
	delete(m.serviceAccounts, name)
	return nil
}

func (m *MemoryIAMBackend) SignJwt(ctx context.Context, name string, payload string) (string, error) {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9." + payload + ".emulator_signature", nil
}

func (m *MemoryIAMBackend) GenerateAccessToken(ctx context.Context, name string) (*model.GenerateAccessTokenResponse, error) {
	return &model.GenerateAccessTokenResponse{
		AccessToken: "ya29.emulator.sa.access.token." + name,
		ExpireTime:  time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *MemoryIAMBackend) Shutdown() error { return nil }
