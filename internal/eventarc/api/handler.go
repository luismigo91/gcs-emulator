package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/eventarc/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/eventarc/model"
)

type Handler struct{ Backend *backend.MemoryEventarcBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "triggers" {
		tname := ""
		if len(parts) >= 6 { tname = parts[5] }
		switch {
		case tname == "" && r.Method == http.MethodPost: h.create(w, r, parent)
		case tname == "" && r.Method == http.MethodGet: h.list(w, r, parent)
		case tname != "" && r.Method == http.MethodGet: h.get(w, r, parent+"/triggers/"+tname)
		case tname != "" && r.Method == http.MethodDelete: h.del(w, r, parent+"/triggers/"+tname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var t model.Trigger; json.NewDecoder(r.Body).Decode(&t)
	if t.Name == "" { t.Name = parent + "/triggers/" + r.URL.Query().Get("triggerId") }
	result, _ := h.Backend.Create(r.Context(), &t)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	t, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, t)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request, parent string) {
	ts, _ := h.Backend.List(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"triggers": ts})
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
