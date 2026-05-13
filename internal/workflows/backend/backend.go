package backend

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/workflows/model"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("workflow not found")

type MemoryWorkflowsBackend struct {
	mu        sync.RWMutex
	workflows map[string]*model.Workflow
}

func NewMemoryWorkflowsBackend() *MemoryWorkflowsBackend {
	return &MemoryWorkflowsBackend{workflows: make(map[string]*model.Workflow)}
}

func (m *MemoryWorkflowsBackend) Create(ctx context.Context, wf *model.Workflow) (*model.Workflow, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.workflows[wf.Name]; exists { return nil, errors.New("exists") }
	wf.State = "ACTIVE"
	m.workflows[wf.Name] = wf
	return wf, nil
}
func (m *MemoryWorkflowsBackend) Get(ctx context.Context, name string) (*model.Workflow, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	wf, exists := m.workflows[name]
	if !exists { return nil, ErrNotFound }
	return wf, nil
}
func (m *MemoryWorkflowsBackend) List(ctx context.Context, parent string) ([]*model.Workflow, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Workflow
	for n, wf := range m.workflows {
		if strings.HasPrefix(n, parent+"/") { result = append(result, wf) }
	}
	return result, nil
}
func (m *MemoryWorkflowsBackend) Delete(ctx context.Context, name string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.workflows, name)
	return nil
}
func (m *MemoryWorkflowsBackend) Execute(ctx context.Context, name string) (*model.Execution, error) {
	m.mu.RLock()
	wf, exists := m.workflows[name]
	m.mu.RUnlock()
	if !exists { return nil, ErrNotFound }

	var doc struct {
		Main struct {
			Steps []map[string]interface{} `yaml:"steps"`
		} `yaml:"main"`
	}
	if err := yaml.Unmarshal([]byte(wf.SourceContents), &doc); err != nil {
		return &model.Execution{State: "FAILED", Error: err.Error()}, nil
	}

	for _, step := range doc.Main.Steps {
		if call, ok := step["callHttp"]; ok {
			cm, ok := call.(map[string]interface{})
			if !ok { continue }
			url := ""
			method := "GET"
			if cm["call"] == "http.post" { method = "POST" }
			if args, ok := cm["args"].(map[string]interface{}); ok {
				if u, ok := args["url"].(string); ok { url = u }
			}
			if url != "" {
				req, _ := http.NewRequest(method, url, nil)
				resp, err := http.DefaultClient.Do(req)
				if err != nil { return &model.Execution{State: "FAILED", Error: err.Error()}, nil }
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				_ = body
			}
		}
	}

	return &model.Execution{State: "SUCCEEDED"}, nil
}
func (m *MemoryWorkflowsBackend) Shutdown() error { return nil }

var _ = strings.TrimPrefix
