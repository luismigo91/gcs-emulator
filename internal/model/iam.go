package model

type PolicyBinding struct {
	Role      string   `json:"role"`
	Members   []string `json:"members"`
	Condition *PolicyCondition `json:"condition,omitempty"`
}

type PolicyCondition struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Expression  string `json:"expression"`
}

type Policy struct {
	Kind     string         `json:"kind"`
	ResourceID string       `json:"resourceId,omitempty"`
	Version  int            `json:"version"`
	Etag     string         `json:"etag"`
	Bindings []PolicyBinding `json:"bindings"`
}

type HMACKey struct {
	AccessID string `json:"accessId"`
	Secret   string `json:"secret"`
	ProjectID string `json:"projectId"`
	ServiceAccountEmail string `json:"serviceAccountEmail"`
	State    string `json:"state"`
	ETag     string `json:"etag"`
}
