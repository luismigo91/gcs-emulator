package model

type Service struct {
	Name   string `json:"name"`
	URL    string `json:"url,omitempty"`
	Status string `json:"status,omitempty"`
}
