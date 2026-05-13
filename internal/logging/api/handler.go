package api

import (
	"encoding/json"
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/logging/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/logging/model"
)

type Handler struct {
	Backend backend.LoggingBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v2/entries:write":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.writeEntries(w, r)
	case "/-/logs":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.listEntries(w, r)
	default:
		writeError(w, http.StatusNotFound, "Not found")
	}
}

func (h *Handler) writeEntries(w http.ResponseWriter, r *http.Request) {
	var req model.WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.Backend.Write(r.Context(), req.Entries); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (h *Handler) listEntries(w http.ResponseWriter, r *http.Request) {
	logName := r.URL.Query().Get("logName")
	entries, _ := h.Backend.List(r.Context(), logName, 100)
	if entries == nil {
		entries = []*model.LogEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": entries})
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
