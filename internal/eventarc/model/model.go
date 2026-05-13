package model

type Trigger struct {
	Name            string            `json:"name"`
	EventType       string            `json:"eventType"`
	Transport       *Transport        `json:"transport,omitempty"`
	Destination     string            `json:"destination,omitempty"`
	ServiceAccount  string            `json:"serviceAccount,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

type Transport struct {
	PubSub *PubSubTarget `json:"pubsub,omitempty"`
}

type PubSubTarget struct {
	Topic string `json:"topic"`
}
