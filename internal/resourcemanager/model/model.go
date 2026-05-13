package model

type Project struct {
	Name       string            `json:"name"`
	ProjectID  string            `json:"projectId"`
	Labels     map[string]string `json:"labels,omitempty"`
	LifecycleState string        `json:"lifecycleState,omitempty"`
}
