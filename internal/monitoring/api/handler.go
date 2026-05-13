package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/model"
)

type Handler struct {
	Backend backend.MonitoringBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, ":createTimeSeries") && r.Method == http.MethodPost:
		h.createTimeSeries(w, r)
	case strings.Contains(r.URL.Path, "/timeSeries") && r.Method == http.MethodGet:
		h.listTimeSeries(w, r)
	default:
		writeError(w, http.StatusBadRequest, "Invalid path or method")
	}
}

func (h *Handler) createTimeSeries(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTimeSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.Backend.CreateTimeSeries(r.Context(), req.TimeSeries); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (h *Handler) listTimeSeries(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	name := r.URL.Query().Get("name")
	project := ""
	if name != "" {
		project = strings.TrimPrefix(name, "projects/")
	}
	_ = project
	series, _ := h.Backend.ListTimeSeries(r.Context(), "", filter, 100)
	if series == nil {
		series = []*model.TimeSeries{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"timeSeries": series})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": status, "message": message},
	})
}
