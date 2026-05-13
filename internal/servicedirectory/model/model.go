package model

type Namespace struct {
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type Service struct {
	Name      string            `json:"name"`
	Endpoints []*Endpoint       `json:"endpoints,omitempty"`
}

type Endpoint struct {
	Name       string `json:"name,omitempty"`
	Address    string `json:"address,omitempty"`
	Port       int    `json:"port"`
	Network    string `json:"network,omitempty"`
}
