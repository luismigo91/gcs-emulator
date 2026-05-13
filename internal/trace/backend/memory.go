package backend

import (
	"context"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/trace/model"
)

type MemoryTraceBackend struct {
	mu    sync.RWMutex
	spans []*model.Span
}

func NewMemoryTraceBackend() *MemoryTraceBackend { return &MemoryTraceBackend{} }

func (m *MemoryTraceBackend) BatchWrite(ctx context.Context, spans []*model.Span) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spans = append(m.spans, spans...)
	return nil
}

func (m *MemoryTraceBackend) List(ctx context.Context, pageSize int) ([]*model.Span, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*model.Span, 0)
	count := 0
	for i := len(m.spans) - 1; i >= 0; i-- {
		result = append(result, m.spans[i])
		count++
		if pageSize > 0 && count >= pageSize {
			break
		}
	}
	return result, nil
}

func (m *MemoryTraceBackend) GetTrace(ctx context.Context, traceID string) ([]*model.Span, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.Span
	for _, s := range m.spans {
		if s.TraceID == traceID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MemoryTraceBackend) Shutdown() error { return nil }
