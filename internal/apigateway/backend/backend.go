package backend

import (
	"context"
	"errors"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/apigateway/model"
)

var ErrNotFound = errors.New("not found")

type MemoryAPIGatewayBackend struct {
	mu       sync.RWMutex
	gateways map[string]*model.Gateway
	configs  map[string]*model.APIConfig
	apis     map[string]*model.API
}

func NewMemoryAPIGatewayBackend() *MemoryAPIGatewayBackend {
	return &MemoryAPIGatewayBackend{
		gateways: make(map[string]*model.Gateway),
		configs:  make(map[string]*model.APIConfig),
		apis:     make(map[string]*model.API),
	}
}

func (m *MemoryAPIGatewayBackend) CreateGateway(ctx context.Context, g *model.Gateway) (*model.Gateway, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.gateways[g.Name]; exists { return nil, errors.New("exists") }
	m.gateways[g.Name] = g
	return g, nil
}
func (m *MemoryAPIGatewayBackend) GetGateway(ctx context.Context, name string) (*model.Gateway, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	g, exists := m.gateways[name]
	if !exists { return nil, ErrNotFound }
	return g, nil
}
func (m *MemoryAPIGatewayBackend) ListGateways(ctx context.Context) ([]*model.Gateway, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	result := make([]*model.Gateway, 0, len(m.gateways))
	for _, g := range m.gateways { result = append(result, g) }
	return result, nil
}
func (m *MemoryAPIGatewayBackend) DeleteGateway(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.gateways, name)
	return nil
}
func (m *MemoryAPIGatewayBackend) Shutdown() error { return nil }
