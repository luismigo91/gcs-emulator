package model

import "time"

type Queue struct {
	Name        string            `json:"name"`
	RateLimits  *RateLimits       `json:"rateLimits,omitempty"`
	RetryConfig *RetryConfig      `json:"retryConfig,omitempty"`
	State       string            `json:"state,omitempty"`
}

type RateLimits struct {
	MaxDispatchesPerSecond  float64 `json:"maxDispatchesPerSecond"`
	MaxBurstSize            int     `json:"maxBurstSize"`
	MaxConcurrentDispatches int     `json:"maxConcurrentDispatches"`
}

type RetryConfig struct {
	MaxAttempts int    `json:"maxAttempts"`
	MaxBackoff  string `json:"maxBackoff,omitempty"`
	MinBackoff  string `json:"minBackoff,omitempty"`
}

type Task struct {
	Name          string          `json:"name"`
	ScheduleTime  time.Time       `json:"scheduleTime,omitempty"`
	CreateTime    time.Time       `json:"createTime,omitempty"`
	DispatchDeadline *string      `json:"dispatchDeadline,omitempty"`
	View          string          `json:"view,omitempty"`
	HTTPRequest   *HTTPRequest    `json:"httpRequest,omitempty"`
	AppEngineHTTPRequest *AppEngineHTTPRequest `json:"appEngineHttpRequest,omitempty"`
}

type HTTPRequest struct {
	URL     string            `json:"url"`
	HTTPMethod string         `json:"httpMethod"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
}

type AppEngineHTTPRequest struct {
	HTTPMethod string `json:"httpMethod"`
	RelativeURI string `json:"relativeUri"`
}

type CreateTaskRequest struct {
	Task *Task `json:"task"`
}

type RunTaskRequest struct {
	ResponseView string `json:"responseView,omitempty"`
}
