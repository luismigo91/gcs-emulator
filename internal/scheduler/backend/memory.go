package backend

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/scheduler/model"
)

var ErrJobNotFound = errors.New("job not found")

type MemorySchedulerBackend struct {
	mu   sync.RWMutex
	jobs map[string]*model.Job
}

func NewMemorySchedulerBackend() *MemorySchedulerBackend {
	return &MemorySchedulerBackend{jobs: make(map[string]*model.Job)}
}

func (m *MemorySchedulerBackend) CreateJob(ctx context.Context, job *model.Job) (*model.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[job.Name]; exists {
		return nil, errors.New("job already exists")
	}
	job.State = "ENABLED"
	m.jobs[job.Name] = job
	return job, nil
}

func (m *MemorySchedulerBackend) GetJob(ctx context.Context, name string) (*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, exists := m.jobs[name]
	if !exists {
		return nil, ErrJobNotFound
	}
	return j, nil
}

func (m *MemorySchedulerBackend) ListJobs(ctx context.Context, project, location string) ([]*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefix := "projects/" + project + "/locations/" + location + "/"
	var result []*model.Job
	for n, j := range m.jobs {
		if strings.HasPrefix(n, prefix) {
			result = append(result, j)
		}
	}
	return result, nil
}

func (m *MemorySchedulerBackend) UpdateJob(ctx context.Context, name string, job *model.Job) (*model.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, exists := m.jobs[name]
	if !exists {
		return nil, ErrJobNotFound
	}
	if job.Schedule != "" {
		existing.Schedule = job.Schedule
	}
	if job.TimeZone != "" {
		existing.TimeZone = job.TimeZone
	}
	if job.HTTPTarget != nil {
		existing.HTTPTarget = job.HTTPTarget
	}
	if job.PubSubTarget != nil {
		existing.PubSubTarget = job.PubSubTarget
	}
	return existing, nil
}

func (m *MemorySchedulerBackend) DeleteJob(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[name]; !exists {
		return ErrJobNotFound
	}
	delete(m.jobs, name)
	return nil
}

func (m *MemorySchedulerBackend) PauseJob(ctx context.Context, name string) (*model.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, exists := m.jobs[name]
	if !exists {
		return nil, ErrJobNotFound
	}
	j.State = "PAUSED"
	return j, nil
}

func (m *MemorySchedulerBackend) ResumeJob(ctx context.Context, name string) (*model.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, exists := m.jobs[name]
	if !exists {
		return nil, ErrJobNotFound
	}
	j.State = "ENABLED"
	return j, nil
}

func (m *MemorySchedulerBackend) RunJob(ctx context.Context, name string) error {
	m.mu.RLock()
	j, exists := m.jobs[name]
	m.mu.RUnlock()
	if !exists {
		return ErrJobNotFound
	}
	if j.HTTPTarget != nil {
		req, _ := http.NewRequest(j.HTTPTarget.HTTPMethod, j.HTTPTarget.URI, bytes.NewReader(j.HTTPTarget.Body))
		for k, v := range j.HTTPTarget.Headers {
			req.Header.Set(k, v)
		}
		http.DefaultClient.Do(req)
	}
	return nil
}

func (m *MemorySchedulerBackend) Shutdown() error { return nil }
