package backend

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudfunctions/model"
)

var ErrNotFound = errors.New("function not found")

type MemoryCloudFunctionsBackend struct {
	mu        sync.RWMutex
	functions map[string]*model.Function
}

func NewMemoryCloudFunctionsBackend() *MemoryCloudFunctionsBackend {
	return &MemoryCloudFunctionsBackend{functions: make(map[string]*model.Function)}
}

func (m *MemoryCloudFunctionsBackend) Create(ctx context.Context, f *model.Function) (*model.Function, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.functions[f.Name]; exists { return nil, errors.New("exists") }
	f.Status = "ACTIVE"
	m.functions[f.Name] = f
	return f, nil
}
func (m *MemoryCloudFunctionsBackend) Get(ctx context.Context, name string) (*model.Function, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	f, exists := m.functions[name]
	if !exists { return nil, ErrNotFound }
	return f, nil
}
func (m *MemoryCloudFunctionsBackend) List(ctx context.Context, parent string) ([]*model.Function, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Function
	for n, f := range m.functions {
		if strings.HasPrefix(n, parent+"/") { result = append(result, f) }
	}
	return result, nil
}
func (m *MemoryCloudFunctionsBackend) Update(ctx context.Context, name string, f *model.Function) (*model.Function, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	existing, exists := m.functions[name]
	if !exists { return nil, ErrNotFound }
	if f.Runtime != "" { existing.Runtime = f.Runtime }
	if f.EntryPoint != "" { existing.EntryPoint = f.EntryPoint }
	return existing, nil
}
func (m *MemoryCloudFunctionsBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.functions, name)
	return nil
}
func (m *MemoryCloudFunctionsBackend) Call(ctx context.Context, name string, data string) (*model.CallResponse, error) {
	m.mu.RLock()
	f, exists := m.functions[name]
	m.mu.RUnlock()
	if !exists { return nil, ErrNotFound }

	if f.HTTPTrigger != nil && f.HTTPTrigger.URL != "" {
		body := strings.NewReader(data)
		resp, err := http.Post(f.HTTPTrigger.URL, "application/json", body)
		if err != nil { return &model.CallResponse{Error: err.Error()}, nil }
		defer resp.Body.Close()
		result, _ := io.ReadAll(resp.Body)
		return &model.CallResponse{Result: string(result)}, nil
	}
	return &model.CallResponse{Result: "function called"}, nil
}
func (m *MemoryCloudFunctionsBackend) Shutdown() error { return nil }
