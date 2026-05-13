package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/certificatemanager/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/certificatemanager/model"
)

type Handler struct{ Backend *backend.MemoryCertManagerBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "certificates" {
		cname := ""
		if len(parts) >= 6 { cname = parts[5] }
		switch {
		case cname == "" && r.Method == http.MethodPost:
			h.create(w, r, parent)
		case cname == "" && r.Method == http.MethodGet:
			h.list(w, r, parent)
		case cname != "" && r.Method == http.MethodGet:
			h.get(w, r, parent+"/certificates/"+cname)
		case cname != "" && r.Method == http.MethodDelete:
			h.del(w, r, parent+"/certificates/"+cname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var c model.Certificate; json.NewDecoder(r.Body).Decode(&c)
	if c.Name == "" { c.Name = parent + "/certificates/" + r.URL.Query().Get("certificateId") }
	result, _ := h.Backend.Create(r.Context(), &c)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	c, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, c)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request, parent string) {
	cs, _ := h.Backend.List(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"certificates": cs})
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
