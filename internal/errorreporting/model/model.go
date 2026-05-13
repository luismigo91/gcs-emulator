package model

import "time"

type ReportedErrorEvent struct {
	EventTime time.Time        `json:"eventTime,omitempty"`
	ServiceContext *ServiceContext `json:"serviceContext,omitempty"`
	Message    string           `json:"message,omitempty"`
	Context    *ErrorContext    `json:"context,omitempty"`
}

type ServiceContext struct {
	Service string `json:"service"`
	Version string `json:"version,omitempty"`
}

type ErrorContext struct {
	HTTPRequest *HTTPRequestContext `json:"httpRequest,omitempty"`
	User        string              `json:"user,omitempty"`
	ReportLocation *SourceLocation  `json:"reportLocation,omitempty"`
}

type HTTPRequestContext struct {
	Method   string `json:"method,omitempty"`
	URL      string `json:"url,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}

type SourceLocation struct {
	FilePath     string `json:"filePath,omitempty"`
	LineNumber   int    `json:"lineNumber,omitempty"`
	FunctionName string `json:"functionName,omitempty"`
}

type ReportRequest struct {
	EventTime string               `json:"eventTime,omitempty"`
	Message   string               `json:"message,omitempty"`
	ServiceContext *ServiceContext  `json:"serviceContext,omitempty"`
	Context   *ErrorContext         `json:"context,omitempty"`
}
