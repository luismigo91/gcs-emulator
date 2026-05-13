package model

import "time"

type Span struct {
	Name         string            `json:"name"`
	SpanID       string            `json:"spanId"`
	ParentSpanID string            `json:"parentSpanId,omitempty"`
	TraceID      string            `json:"traceId"`
	StartTime    time.Time         `json:"startTime"`
	EndTime      time.Time         `json:"endTime"`
	Kind         string            `json:"kind,omitempty"`
	Status       *Status           `json:"status,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type BatchWriteRequest struct {
	Spans []*Span `json:"spans"`
}
