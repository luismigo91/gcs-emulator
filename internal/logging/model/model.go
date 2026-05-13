package model

import "time"

type LogEntry struct {
	LogName      string            `json:"logName"`
	Resource     *MonitoredResource `json:"resource"`
	Timestamp    time.Time         `json:"timestamp,omitempty"`
	Severity     string            `json:"severity,omitempty"`
	TextPayload  string            `json:"textPayload,omitempty"`
	JSONPayload  interface{}       `json:"jsonPayload,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	InsertID     string            `json:"insertId,omitempty"`
	HTTPRequest  *HTTPRequest      `json:"httpRequest,omitempty"`
	Trace        string            `json:"trace,omitempty"`
}

type MonitoredResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type HTTPRequest struct {
	RequestMethod string `json:"requestMethod,omitempty"`
	RequestURL    string `json:"requestUrl,omitempty"`
	Status        int    `json:"status,omitempty"`
}

type WriteRequest struct {
	Entries []*LogEntry `json:"entries"`
}
