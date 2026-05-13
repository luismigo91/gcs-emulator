package backend

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/certificatemanager/model"
)

var ErrNotFound = errors.New("not found")

type MemoryCertManagerBackend struct {
	mu    sync.RWMutex
	certs map[string]*model.Certificate
}

func NewMemoryCertManagerBackend() *MemoryCertManagerBackend {
	return &MemoryCertManagerBackend{certs: make(map[string]*model.Certificate)}
}

func (m *MemoryCertManagerBackend) Create(ctx context.Context, cert *model.Certificate) (*model.Certificate, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.certs[cert.Name]; exists { return nil, errors.New("exists") }
	cert.CreateTime = time.Now()
	cert.ExpireTime = time.Now().Add(90 * 24 * time.Hour)
	m.certs[cert.Name] = cert
	return cert, nil
}

func (m *MemoryCertManagerBackend) Get(ctx context.Context, name string) (*model.Certificate, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	c, exists := m.certs[name]
	if !exists { return nil, ErrNotFound }
	return c, nil
}

func (m *MemoryCertManagerBackend) List(ctx context.Context, parent string) ([]*model.Certificate, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Certificate
	for n, c := range m.certs {
		if strings.HasPrefix(n, parent+"/") { result = append(result, c) }
	}
	return result, nil
}

func (m *MemoryCertManagerBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.certs, name)
	return nil
}

func (m *MemoryCertManagerBackend) Shutdown() error { return nil }
