package model

type ErrorDetail struct {
	Domain       string `json:"domain"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	LocationType string `json:"locationType,omitempty"`
	Location     string `json:"location,omitempty"`
}

type GCPError struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors"`
}

type GCPErrorResponse struct {
	Error GCPError `json:"error"`
}

func NewGCPError(code int, message string, errors []ErrorDetail) *GCPErrorResponse {
	return &GCPErrorResponse{
		Error: GCPError{
			Code:    code,
			Message: message,
			Errors:  errors,
		},
	}
}

func NewSimpleGCPError(code int, message string) *GCPErrorResponse {
	return NewGCPError(code, message, []ErrorDetail{
		{
			Domain:  "global",
			Reason:  getReasonFromCode(code),
			Message: message,
		},
	})
}

func getReasonFromCode(code int) string {
	switch code {
	case 400:
		return "invalid"
	case 403:
		return "forbidden"
	case 404:
		return "notFound"
	case 409:
		return "duplicate"
	case 412:
		return "conditionNotMet"
	case 416:
		return "invalidRange"
	case 500:
		return "internalError"
	default:
		return "unknown"
	}
}
