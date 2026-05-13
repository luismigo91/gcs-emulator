package backend

import (
	"context"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/model"
)

type MemoryMonitoringBackend struct {
	mu     sync.RWMutex
	series []*model.TimeSeries
}

func NewMemoryMonitoringBackend() *MemoryMonitoringBackend {
	return &MemoryMonitoringBackend{}
}

func (m *MemoryMonitoringBackend) CreateTimeSeries(ctx context.Context, series []*model.TimeSeries) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.series = append(m.series, series...)
	return nil
}

func (m *MemoryMonitoringBackend) ListTimeSeries(ctx context.Context, project, filter string, pageSize int) ([]*model.TimeSeries, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.TimeSeries
	count := 0
	for i := len(m.series) - 1; i >= 0; i-- {
		s := m.series[i]
		if filter != "" && !matchFilter(s, filter) {
			continue
		}
		result = append(result, s)
		count++
		if pageSize > 0 && count >= pageSize {
			break
		}
	}
	return result, nil
}

func matchFilter(s *model.TimeSeries, filter string) bool {
	if strings.Contains(filter, "metric.type") {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) == 2 {
			want := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			if s.Metric != nil && s.Metric.Type == want {
				return true
			}
		}
	}
	return true
}

func (m *MemoryMonitoringBackend) Shutdown() error { return nil }
