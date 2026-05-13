package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/workflows/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/workflows/model"
)

type Handler struct{ Backend *backend.MemoryWorkflowsBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "workflows" {
		wfname := ""
		if len(parts) >= 6 { wfname = parts[5] }
		if strings.HasSuffix(wfname, ":execute") {
			h.execute(w, r, parent+"/workflows/"+strings.TrimSuffix(wfname, ":execute"))
			return
		}
		switch {
		case wfname == "" && r.Method == http.MethodPost: h.create(w, r, parent)
		case wfname == "" && r.Method == http.MethodGet: h.list(w, r, parent)
		case wfname != "" && r.Method == http.MethodGet: h.get(w, r, parent+"/workflows/"+wfname)
		case wfname != "" && r.Method == http.MethodDelete: h.del(w, r, parent+"/workflows/"+wfname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var wf model.Workflow; json.NewDecoder(r.Body).Decode(&wf)
	if wf.Name == "" { wf.Name = parent + "/workflows/" + r.URL.Query().Get("workflowId") }
	result, _ := h.Backend.Create(r.Context(), &wf)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	wf, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, wf)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request, parent string) {
	wfs, _ := h.Backend.List(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"workflows": wfs})
}
func (h *Handler) del(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.Delete(r.Context(), name); w.WriteHeader(http.StatusOK)
}
func (h *Handler) execute(w http.ResponseWriter, r *http.Request, name string) {
	exec, _ := h.Backend.Execute(r.Context(), name)
	writeJSON(w, http.StatusOK, exec)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d)
}
func writeError(w http.ResponseWriter, s int, m string) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}})
}
