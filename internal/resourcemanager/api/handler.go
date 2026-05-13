package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/resourcemanager/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/resourcemanager/model"
)

type Handler struct{ Backend *backend.MemoryResourceManagerBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "projects" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		h.list(w, r); return
	}

	projectID := parts[1]
	switch r.Method {
	case http.MethodGet: h.get(w, r, projectID)
	case http.MethodPost: h.create(w, r)
	case http.MethodDelete: h.del(w, r, projectID)
	default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var p model.Project; json.NewDecoder(r.Body).Decode(&p)
	if p.ProjectID == "" { p.ProjectID = r.URL.Query().Get("projectId") }
	result, _ := h.Backend.Create(r.Context(), &p)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, id string) {
	p, _ := h.Backend.Get(r.Context(), id)
	writeJSON(w, http.StatusOK, p)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ps, _ := h.Backend.List(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{"projects": ps})
}
func (h *Handler) del(w http.ResponseWriter, r *http.Request, id string) {
	h.Backend.Delete(r.Context(), id); w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d)
}
func writeError(w http.ResponseWriter, s int, m string) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}})
}
