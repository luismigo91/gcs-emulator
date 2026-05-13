package api

import (
	"encoding/json"
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/trace/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/trace/model"
)

type Handler struct {
	Backend backend.TraceBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v2/traces:batchWrite":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.batchWrite(w, r)
	case "/-/traces":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		h.listTraces(w, r)
	default:
		writeError(w, http.StatusNotFound, "Not found")
	}
}

func (h *Handler) batchWrite(w http.ResponseWriter, r *http.Request) {
	var req model.BatchWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	h.Backend.BatchWrite(r.Context(), req.Spans)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (h *Handler) listTraces(w http.ResponseWriter, r *http.Request) {
	traceID := r.URL.Query().Get("traceId")
	var spans []*model.Span
	if traceID != "" {
		spans, _ = h.Backend.GetTrace(r.Context(), traceID)
	} else {
		spans, _ = h.Backend.List(r.Context(), 100)
	}
	if spans == nil {
		spans = []*model.Span{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"spans": spans})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": status, "message": message},
	})
}
