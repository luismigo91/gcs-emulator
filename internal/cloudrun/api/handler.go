package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrun/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrun/model"
)

type Handler struct{ Backend *backend.MemoryCloudRunBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "services" {
		sname := ""
		if len(parts) >= 6 { sname = parts[5] }
		switch {
		case sname == "" && r.Method == http.MethodPost: h.create(w, r, parent)
		case sname == "" && r.Method == http.MethodGet: h.list(w, r, parent)
		case sname != "" && r.Method == http.MethodGet: h.get(w, r, parent+"/services/"+sname)
		case sname != "" && r.Method == http.MethodDelete: h.del(w, r, parent+"/services/"+sname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var s model.Service; json.NewDecoder(r.Body).Decode(&s)
	if s.Name == "" { s.Name = parent + "/services/" + r.URL.Query().Get("serviceId") }
	result, _ := h.Backend.Create(r.Context(), &s)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	s, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, s)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request, parent string) {
	ss, _ := h.Backend.List(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"services": ss})
}
func (h *Handler) del(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.Delete(r.Context(), name); w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d)
}
func writeError(w http.ResponseWriter, s int, m string) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}})
}
