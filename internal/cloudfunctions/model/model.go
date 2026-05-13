package model

type Function struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Runtime     string            `json:"runtime,omitempty"`
	EntryPoint  string            `json:"entryPoint,omitempty"`
	HTTPTrigger *HTTPTrigger      `json:"httpsTrigger,omitempty"`
	EventTrigger *EventTrigger    `json:"eventTrigger,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Status      string            `json:"status,omitempty"`
}

type HTTPTrigger struct {
	URL string `json:"url,omitempty"`
}

type EventTrigger struct {
	EventType string `json:"eventType"`
	Resource  string `json:"resource"`
}

type CallRequest struct {
	Data string `json:"data,omitempty"`
}

type CallResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}
