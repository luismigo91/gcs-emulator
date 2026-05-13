package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudbuild/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudbuild/model"
)

type Handler struct{ Backend backend.CloudBuildBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[0] != "projects" || parts[2] != "triggers" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project := parts[1]
	parent := "projects/" + project

	if len(parts) == 3 {
		switch r.Method {
		case http.MethodPost: h.createTrigger(w, r, parent)
		case http.MethodGet: h.listTriggers(w, r, project)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	tName := parts[3]
	full := parent + "/triggers/" + tName
	if strings.HasSuffix(tName, ":run") {
		h.runTrigger(w, r, parent+"/triggers/"+strings.TrimSuffix(tName, ":run"))
		return
	}
	switch r.Method {
	case http.MethodGet: h.getTrigger(w, r, full)
	case http.MethodDelete: h.deleteTrigger(w, r, full)
	default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) createTrigger(w http.ResponseWriter, r *http.Request, parent string) {
	var t model.BuildTrigger
	json.NewDecoder(r.Body).Decode(&t)
	if t.Name == "" { t.Name = parent + "/triggers/" + r.URL.Query().Get("triggerId") }
	result, _ := h.Backend.CreateTrigger(r.Context(), &t)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getTrigger(w http.ResponseWriter, r *http.Request, name string) {
	t, err := h.Backend.GetTrigger(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, t)
}
func (h *Handler) listTriggers(w http.ResponseWriter, r *http.Request, project string) {
	ts, _ := h.Backend.ListTriggers(r.Context(), project)
	writeJSON(w, http.StatusOK, map[string]interface{}{"triggers": ts})
}
func (h *Handler) deleteTrigger(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.DeleteTrigger(r.Context(), name)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) runTrigger(w http.ResponseWriter, r *http.Request, name string) {
	b, _ := h.Backend.RunTrigger(r.Context(), name)
	writeJSON(w, http.StatusOK, b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": status, "message": message},
	})
}
