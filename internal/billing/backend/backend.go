package backend

import (
	"context"
	"errors"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/billing/model"
)

var ErrBudgetNotFound = errors.New("budget not found")

type MemoryBillingBackend struct {
	mu      sync.RWMutex
	budgets map[string]*model.Budget
}

func NewMemoryBillingBackend() *MemoryBillingBackend {
	return &MemoryBillingBackend{budgets: make(map[string]*model.Budget)}
}

func (m *MemoryBillingBackend) CreateBudget(ctx context.Context, budget *model.Budget) (*model.Budget, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.budgets[budget.Name]; exists { return nil, errors.New("budget exists") }
	m.budgets[budget.Name] = budget
	return budget, nil
}
func (m *MemoryBillingBackend) GetBudget(ctx context.Context, name string) (*model.Budget, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	b, exists := m.budgets[name]
	if !exists { return nil, ErrBudgetNotFound }
	return b, nil
}
func (m *MemoryBillingBackend) ListBudgets(ctx context.Context) ([]*model.Budget, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	result := make([]*model.Budget, 0, len(m.budgets))
	for _, b := range m.budgets { result = append(result, b) }
	return result, nil
}
func (m *MemoryBillingBackend) DeleteBudget(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.budgets, name)
	return nil
}
func (m *MemoryBillingBackend) Shutdown() error { return nil }
