package api

import (
	"encoding/json"
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/asset/backend"
)

type Handler struct{ Backend *backend.MemoryAssetBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		assets, _ := h.Backend.Export(r.Context(), nil)
		writeJSON(w, http.StatusOK, map[string]interface{}{"assets": assets})
	case http.MethodGet:
		assets, _ := h.Backend.Export(r.Context(), nil)
		writeJSON(w, http.StatusOK, map[string]interface{}{"assets": assets})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}
func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
