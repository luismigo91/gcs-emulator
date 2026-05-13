package backend

import (
	"context"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/errorreporting/model"
)

type MemoryErrorReportingBackend struct {
	mu     sync.RWMutex
	events []*model.ReportedErrorEvent
}

func NewMemoryErrorReportingBackend() *MemoryErrorReportingBackend {
	return &MemoryErrorReportingBackend{}
}

func (m *MemoryErrorReportingBackend) Report(ctx context.Context, event *model.ReportedErrorEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}
	m.events = append(m.events, event)
	return nil
}

func (m *MemoryErrorReportingBackend) List(ctx context.Context, pageSize int) ([]*model.ReportedErrorEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*model.ReportedErrorEvent, 0)
	count := 0
	for i := len(m.events) - 1; i >= 0; i-- {
		result = append(result, m.events[i])
		count++
		if pageSize > 0 && count >= pageSize {
			break
		}
	}
	return result, nil
}

func (m *MemoryErrorReportingBackend) Shutdown() error { return nil }
