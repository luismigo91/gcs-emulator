package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudfunctions/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudfunctions/model"
)

type Handler struct{ Backend *backend.MemoryCloudFunctionsBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "functions" {
		fname := ""
		if len(parts) >= 6 { fname = parts[5] }
		if strings.HasSuffix(fname, ":call") {
			h.call(w, r, parent+"/functions/"+strings.TrimSuffix(fname, ":call"))
			return
		}
		switch {
		case fname == "" && r.Method == http.MethodPost: h.create(w, r, parent)
		case fname == "" && r.Method == http.MethodGet: h.list(w, r, parent)
		case fname != "" && r.Method == http.MethodGet: h.get(w, r, parent+"/functions/"+fname)
		case fname != "" && r.Method == http.MethodPatch: h.update(w, r, parent+"/functions/"+fname)
		case fname != "" && r.Method == http.MethodDelete: h.del(w, r, parent+"/functions/"+fname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var f model.Function; json.NewDecoder(r.Body).Decode(&f)
	if f.Name == "" { f.Name = parent + "/functions/" + r.URL.Query().Get("functionId") }
	result, _ := h.Backend.Create(r.Context(), &f)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	f, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, f)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request, parent string) {
	fs, _ := h.Backend.List(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"functions": fs})
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request, name string) {
	var f model.Function; json.NewDecoder(r.Body).Decode(&f)
	result, _ := h.Backend.Update(r.Context(), name, &f)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) del(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.Delete(r.Context(), name); w.WriteHeader(http.StatusOK)
}
func (h *Handler) call(w http.ResponseWriter, r *http.Request, name string) {
	var req model.CallRequest; json.NewDecoder(r.Body).Decode(&req)
	result, _ := h.Backend.Call(r.Context(), name, req.Data)
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d)
}
func writeError(w http.ResponseWriter, s int, m string) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}})
}
