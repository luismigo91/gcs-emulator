package backend

import (
	"context"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/logging/model"
)

type MemoryLoggingBackend struct {
	mu      sync.RWMutex
	entries []*model.LogEntry
}

func NewMemoryLoggingBackend() *MemoryLoggingBackend {
	return &MemoryLoggingBackend{}
}

func (m *MemoryLoggingBackend) Write(ctx context.Context, entries []*model.LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, e := range entries {
		if e.Timestamp.IsZero() {
			e.Timestamp = now
		}
		m.entries = append(m.entries, e)
	}
	return nil
}

func (m *MemoryLoggingBackend) List(ctx context.Context, logName string, pageSize int) ([]*model.LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.LogEntry
	count := 0
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := m.entries[i]
		if logName != "" && e.LogName != logName {
			continue
		}
		result = append(result, e)
		count++
		if pageSize > 0 && count >= pageSize {
			break
		}
	}
	return result, nil
}

func (m *MemoryLoggingBackend) Shutdown() error { return nil }
