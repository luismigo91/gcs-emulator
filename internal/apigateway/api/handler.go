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
	if len(parts) < 6 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "gateways" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	parent := strings.Join(parts[:4], "/")
	gatewayName := ""
	if len(parts) >= 6 { gatewayName = parts[5] }
	fullName := parent + "/gateways/" + gatewayName

	switch {
	case gatewayName == "" && r.Method == http.MethodPost:
		h.createGateway(w, r, parent)
	case gatewayName == "" && r.Method == http.MethodGet:
		h.listGateways(w, r)
	case gatewayName != "" && r.Method == http.MethodGet:
		h.getGateway(w, r, fullName)
	case gatewayName != "" && r.Method == http.MethodDelete:
		h.deleteGateway(w, r, fullName)
	case gatewayName != "" && r.Method == http.MethodPatch:
		h.updateGateway(w, r, fullName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) createGateway(w http.ResponseWriter, r *http.Request, parent string) {
	var g model.Gateway
	json.NewDecoder(r.Body).Decode(&g)
	if g.Name == "" { g.Name = parent + "/gateways/" + r.URL.Query().Get("gatewayId") }
	result, _ := h.Backend.CreateGateway(r.Context(), &g)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getGateway(w http.ResponseWriter, r *http.Request, name string) {
	g, err := h.Backend.GetGateway(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, g)
}
func (h *Handler) listGateways(w http.ResponseWriter, r *http.Request) {
	gs, _ := h.Backend.ListGateways(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{"gateways": gs})
}
func (h *Handler) deleteGateway(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.DeleteGateway(r.Context(), name)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) updateGateway(w http.ResponseWriter, r *http.Request, name string) {
	var g model.Gateway
	json.NewDecoder(r.Body).Decode(&g)
	result, _ := h.Backend.UpdateGateway(r.Context(), name, &g)
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
func writeError(w http.ResponseWriter, s int, m string) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}}) }
