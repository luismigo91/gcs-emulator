package model

type Job struct {
	Name      string     `json:"name"`
	Schedule  string     `json:"schedule"`
	TimeZone  string     `json:"timeZone,omitempty"`
	State     string     `json:"state,omitempty"`
	HTTPTarget *HTTPTarget `json:"httpTarget,omitempty"`
	PubSubTarget *PubSubTarget `json:"pubsubTarget,omitempty"`
	RetryConfig *RetryConfig  `json:"retryConfig,omitempty"`
}

type HTTPTarget struct {
	URI     string            `json:"uri"`
	HTTPMethod string         `json:"httpMethod"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
}

type PubSubTarget struct {
	TopicName  string            `json:"topicName"`
	Data       []byte            `json:"data,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type RetryConfig struct {
	RetryCount   int    `json:"retryCount,omitempty"`
	MaxBackoff   string `json:"maxBackoffDuration,omitempty"`
	MinBackoff   string `json:"minBackoffDuration,omitempty"`
}
