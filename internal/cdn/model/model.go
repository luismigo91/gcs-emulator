package model

type BackendService struct {
	Name              string   `json:"name"`
	Backend           string   `json:"backend"`
	Protocol          string   `json:"protocol,omitempty"`
	TimeoutSec        int      `json:"timeoutSec,omitempty"`
	EnableCDN         bool     `json:"enableCDN,omitempty"`
	HealthChecks      []string `json:"healthChecks,omitempty"`
}

type URLMap struct {
	Name            string    `json:"name"`
	DefaultService  string    `json:"defaultService"`
	HostRules       []*HostRule `json:"hostRules,omitempty"`
}

type HostRule struct {
	Hosts       []string `json:"hosts"`
	PathMatcher string   `json:"pathMatcher"`
}
