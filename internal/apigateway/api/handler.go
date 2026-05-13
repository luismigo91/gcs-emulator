package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/apigateway/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/apigateway/model"
)

type Handler struct{ Backend *backend.MemoryAPIGatewayBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	isGateway := len(parts) >= 4 && parts[2] == "locations" && len(parts) >= 6 && parts[5] == "gateways"
	_ = isGateway

	switch r.Method {
	case http.MethodPost:
		var g model.Gateway
		json.NewDecoder(r.Body).Decode(&g)
		if g.Name == "" { g.Name = r.URL.Query().Get("gatewayId") }
		result, _ := h.Backend.CreateGateway(r.Context(), &g)
		writeJSON(w, http.StatusOK, result)
	case http.MethodGet:
		gs, _ := h.Backend.ListGateways(r.Context())
		writeJSON(w, http.StatusOK, map[string]interface{}{"gateways": gs})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}
func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
