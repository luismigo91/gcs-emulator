package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/serviceusage/backend"
)

type Handler struct{ Backend *backend.MemoryServiceUsageBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == "projects" && parts[2] == "services" {
		svc := ""
		if len(parts) >= 4 { svc = parts[3] }
		if strings.HasSuffix(svc, ":enable") {
			w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{}) ; return
		}
		if strings.HasSuffix(svc, ":disable") {
			w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{}) ; return
		}
		if svc != "" {
			s, _ := h.Backend.Get(r.Context(), svc)
			w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(s); return
		}
		services, _ := h.Backend.List(r.Context())
		w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"services": services}); return
	}
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
}
