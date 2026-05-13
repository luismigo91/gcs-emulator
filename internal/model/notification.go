package model

type Notification struct {
	Kind         string            `json:"kind"`
	ID           string            `json:"id"`
	SelfLink     string            `json:"selfLink,omitempty"`
	Topic        string            `json:"topic"`
	EventType    []string          `json:"event_types,omitempty"`
	ObjectNamePrefix string        `json:"object_name_prefix,omitempty"`
	CustomAttributes map[string]string `json:"custom_attributes,omitempty"`
	PayloadFormat string           `json:"payload_format,omitempty"`
}
