package api

import (
	"encoding/json"
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/errorreporting/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/errorreporting/model"
)

type Handler struct {
	Backend backend.ErrorReportingBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1beta1/projects/test-project/events:report":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.report(w, r)
	case "/-/errors":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.list(w, r)
	default:
		writeError(w, http.StatusNotFound, "Not found")
	}
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	var req model.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	event := &model.ReportedErrorEvent{
		Message:        req.Message,
		ServiceContext: req.ServiceContext,
		Context:        req.Context,
	}
	h.Backend.Report(r.Context(), event)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	events, _ := h.Backend.List(r.Context(), 50)
	if events == nil {
		events = []*model.ReportedErrorEvent{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"errorEvents": events})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": status, "message": message},
	})
}
