package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cdn/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cdn/model"
)

type Handler struct{ Backend *backend.MemoryCDNBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/compute/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "projects" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project := parts[1]
	resource := ""
	if len(parts) >= 4 && parts[2] == "global" { resource = parts[3] }
	name := ""
	if len(parts) >= 5 { name = parts[4] }

	if resource == "backendServices" {
		switch {
		case name == "" && r.Method == http.MethodPost: h.createBackend(w, r, project)
		case name == "" && r.Method == http.MethodGet: h.listBackends(w, r, project)
		case name != "" && r.Method == http.MethodGet: h.getBackend(w, r, "projects/"+project+"/global/backendServices/"+name)
		case name != "" && r.Method == http.MethodDelete: h.deleteBackend(w, r, "projects/"+project+"/global/backendServices/"+name)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	if resource == "urlMaps" {
		switch {
		case name == "" && r.Method == http.MethodPost: h.createURLMap(w, r, project)
		case name == "" && r.Method == http.MethodGet: h.listURLMaps(w, r, project)
		case name != "" && r.Method == http.MethodGet: h.getURLMap(w, r, "projects/"+project+"/global/urlMaps/"+name)
		case name != "" && r.Method == http.MethodDelete: h.deleteURLMap(w, r, "projects/"+project+"/global/urlMaps/"+name)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid resource")
}

func (h *Handler) createBackend(w http.ResponseWriter, r *http.Request, p string) {
	var b model.BackendService; json.NewDecoder(r.Body).Decode(&b)
	if b.Name == "" { b.Name = "projects/" + p + "/global/backendServices/" + r.URL.Query().Get("name") }
	result, _ := h.Backend.CreateBackend(r.Context(), &b)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getBackend(w http.ResponseWriter, r *http.Request, n string) {
	b, err := h.Backend.GetBackend(r.Context(), n)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, b)
}
func (h *Handler) listBackends(w http.ResponseWriter, r *http.Request, p string) {
	bs, _ := h.Backend.ListBackends(r.Context(), p)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": bs})
}
func (h *Handler) deleteBackend(w http.ResponseWriter, r *http.Request, n string) {
	h.Backend.DeleteBackend(r.Context(), n)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) createURLMap(w http.ResponseWriter, r *http.Request, p string) {
	var um model.URLMap; json.NewDecoder(r.Body).Decode(&um)
	if um.Name == "" { um.Name = "projects/" + p + "/global/urlMaps/" + r.URL.Query().Get("name") }
	result, _ := h.Backend.CreateURLMap(r.Context(), &um)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getURLMap(w http.ResponseWriter, r *http.Request, n string) {
	um, err := h.Backend.GetURLMap(r.Context(), n)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, um)
}
func (h *Handler) listURLMaps(w http.ResponseWriter, r *http.Request, p string) {
	ums, _ := h.Backend.ListURLMaps(r.Context(), p)
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": ums})
}
func (h *Handler) deleteURLMap(w http.ResponseWriter, r *http.Request, n string) {
	h.Backend.DeleteURLMap(r.Context(), n)
	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
func writeError(w http.ResponseWriter, s int, m string) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}}) }
