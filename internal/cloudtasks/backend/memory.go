package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/model"
)

var (
	ErrQueueNotFound        = errors.New("queue not found")
	ErrQueueAlreadyExists   = errors.New("queue already exists")
	ErrTaskNotFound         = errors.New("task not found")
)

type MemoryCloudTasksBackend struct {
	mu      sync.RWMutex
	queues  map[string]*model.Queue
	tasks   map[string][]*model.Task
}

func NewMemoryCloudTasksBackend() *MemoryCloudTasksBackend {
	return &MemoryCloudTasksBackend{
		queues: make(map[string]*model.Queue),
		tasks:  make(map[string][]*model.Task),
	}
}

func (m *MemoryCloudTasksBackend) CreateQueue(ctx context.Context, queue *model.Queue) (*model.Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.queues[queue.Name]; exists {
		return nil, ErrQueueAlreadyExists
	}
	queue.State = "RUNNING"
	m.queues[queue.Name] = queue
	m.tasks[queue.Name] = make([]*model.Task, 0)
	return queue, nil
}

func (m *MemoryCloudTasksBackend) GetQueue(ctx context.Context, name string) (*model.Queue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	q, exists := m.queues[name]
	if !exists {
		return nil, ErrQueueNotFound
	}
	return q, nil
}

func (m *MemoryCloudTasksBackend) ListQueues(ctx context.Context, project, location string) ([]*model.Queue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefix := "projects/" + project + "/locations/" + location + "/queues/"
	var result []*model.Queue
	for name, q := range m.queues {
		if strings.HasPrefix(name, prefix) {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *MemoryCloudTasksBackend) UpdateQueue(ctx context.Context, name string, queue *model.Queue) (*model.Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, exists := m.queues[name]
	if !exists {
		return nil, ErrQueueNotFound
	}
	if queue.RateLimits != nil {
		existing.RateLimits = queue.RateLimits
	}
	if queue.RetryConfig != nil {
		existing.RetryConfig = queue.RetryConfig
	}
	return existing, nil
}

func (m *MemoryCloudTasksBackend) DeleteQueue(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.queues[name]; !exists {
		return ErrQueueNotFound
	}
	delete(m.queues, name)
	delete(m.tasks, name)
	return nil
}

func (m *MemoryCloudTasksBackend) PauseQueue(ctx context.Context, name string) (*model.Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, exists := m.queues[name]
	if !exists {
		return nil, ErrQueueNotFound
	}
	q.State = "PAUSED"
	return q, nil
}

func (m *MemoryCloudTasksBackend) ResumeQueue(ctx context.Context, name string) (*model.Queue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, exists := m.queues[name]
	if !exists {
		return nil, ErrQueueNotFound
	}
	q.State = "RUNNING"
	return q, nil
}

func (m *MemoryCloudTasksBackend) PurgeQueue(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.queues[name]; !exists {
		return ErrQueueNotFound
	}
	m.tasks[name] = make([]*model.Task, 0)
	return nil
}

func (m *MemoryCloudTasksBackend) CreateTask(ctx context.Context, parent string, task *model.Task) (*model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[parent]; !exists {
		return nil, ErrQueueNotFound
	}

	task.CreateTime = time.Now()
	task.Name = parent + "/tasks/" + fmt.Sprintf("%d", len(m.tasks[parent])+1)
	m.tasks[parent] = append(m.tasks[parent], task)

	q := m.queues[parent]
	if q.State == "RUNNING" && task.HTTPRequest != nil && task.HTTPRequest.URL != "" {
		go func() {
			req, _ := http.NewRequest(task.HTTPRequest.HTTPMethod, task.HTTPRequest.URL, strings.NewReader(string(task.HTTPRequest.Body)))
			for k, v := range task.HTTPRequest.Headers {
				req.Header.Set(k, v)
			}
			http.DefaultClient.Do(req)
		}()
	}

	return task, nil
}

func (m *MemoryCloudTasksBackend) GetTask(ctx context.Context, name string) (*model.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.findTaskLocked(name)
}

func (m *MemoryCloudTasksBackend) ListTasks(ctx context.Context, parent string) ([]*model.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.queues[parent]; !exists {
		return nil, ErrQueueNotFound
	}
	return m.tasks[parent], nil
}

func (m *MemoryCloudTasksBackend) DeleteTask(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	parts := strings.Split(name, "/tasks/")
	if len(parts) != 2 {
		return ErrTaskNotFound
	}
	queue := parts[0]
	taskID := parts[1]
	tasks := m.tasks[queue]
	for i, t := range tasks {
		if t.Name == name || strings.HasSuffix(t.Name, "/"+taskID) || t.Name == taskID {
			m.tasks[queue] = append(tasks[:i], tasks[i+1:]...)
			return nil
		}
	}
	return ErrTaskNotFound
}

func (m *MemoryCloudTasksBackend) RunTask(ctx context.Context, name string) (*model.Task, error) {
	m.mu.RLock()
	task, err := m.findTaskLocked(name)
	m.mu.RUnlock()
	if err != nil {
		return nil, err
	}
	if task.HTTPRequest != nil && task.HTTPRequest.URL != "" {
		req, _ := http.NewRequest(task.HTTPRequest.HTTPMethod, task.HTTPRequest.URL, strings.NewReader(string(task.HTTPRequest.Body)))
		for k, v := range task.HTTPRequest.Headers {
			req.Header.Set(k, v)
		}
		http.DefaultClient.Do(req)
	}
	return task, nil
}

func (m *MemoryCloudTasksBackend) findTaskLocked(name string) (*model.Task, error) {
	parts := strings.Split(name, "/tasks/")
	if len(parts) != 2 {
		for _, tasks := range m.tasks {
			for _, t := range tasks {
				if t.Name == name {
					return t, nil
				}
			}
		}
		return nil, ErrTaskNotFound
	}
	queue := parts[0]
	tasks := m.tasks[queue]
	taskID := parts[1]
	for _, t := range tasks {
		if t.Name == name || strings.HasSuffix(t.Name, "/"+taskID) || t.Name == taskID {
			return t, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (m *MemoryCloudTasksBackend) Shutdown() error { return nil }
