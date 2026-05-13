package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/apikeys/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/apikeys/model"
)

type Handler struct{ Backend *backend.MemoryAPIKeysBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path"); return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "keys" {
		keyName := ""
		if len(parts) >= 6 { keyName = parts[5] }
		switch {
		case keyName == "" && r.Method == http.MethodPost: h.create(w, r, parent)
		case keyName == "" && r.Method == http.MethodGet: h.list(w, r)
		case keyName != "" && r.Method == http.MethodGet: h.get(w, r, parent+"/keys/"+keyName)
		case keyName != "" && r.Method == http.MethodDelete: h.del(w, r, parent+"/keys/"+keyName)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, parent string) {
	var k model.Key; json.NewDecoder(r.Body).Decode(&k)
	if k.Name == "" { k.Name = parent + "/keys/" + r.URL.Query().Get("keyId") }
	result, _ := h.Backend.Create(r.Context(), &k)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, name string) {
	k, err := h.Backend.Get(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, k)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ks, _ := h.Backend.List(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{"keys": ks})
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
