package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/servicedirectory/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/servicedirectory/model"
)

type Handler struct{ Backend *backend.MemoryServiceDirectoryBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project, location := parts[1], parts[3]
	if len(parts) >= 5 && parts[4] == "namespaces" {
		ns := ""
		if len(parts) >= 6 { ns = parts[5] }
		if ns == "" {
			switch r.Method {
			case http.MethodPost: h.createNS(w, r, project, location)
			case http.MethodGet: h.listNS(w, r, project, location)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		nsKey := project + "/" + location + "/" + ns
		if len(parts) >= 7 && parts[6] == "services" {
			switch r.Method {
			case http.MethodPost: h.createSvc(w, r, project, location, ns)
			case http.MethodGet: h.listSvc(w, r, project, location, ns)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		switch r.Method {
		case http.MethodGet: h.getNS(w, r, nsKey)
		case http.MethodDelete: h.deleteNS(w, r, nsKey)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

func (h *Handler) createNS(w http.ResponseWriter, r *http.Request, p, l string) {
	var ns model.Namespace; json.NewDecoder(r.Body).Decode(&ns)
	if ns.Name == "" { ns.Name = r.URL.Query().Get("namespaceId") }
	result, _ := h.Backend.CreateNamespace(r.Context(), p, l, &ns)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getNS(w http.ResponseWriter, r *http.Request, k string) {
	ns, err := h.Backend.GetNamespace(r.Context(), k)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, ns)
}
func (h *Handler) listNS(w http.ResponseWriter, r *http.Request, p, l string) {
	nss, _ := h.Backend.ListNamespaces(r.Context(), p, l)
	writeJSON(w, http.StatusOK, map[string]interface{}{"namespaces": nss})
}
func (h *Handler) deleteNS(w http.ResponseWriter, r *http.Request, k string) {
	h.Backend.DeleteNamespace(r.Context(), k)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) createSvc(w http.ResponseWriter, r *http.Request, p, l, ns string) {
	var svc model.Service; json.NewDecoder(r.Body).Decode(&svc)
	if svc.Name == "" { svc.Name = r.URL.Query().Get("serviceId") }
	result, _ := h.Backend.CreateService(r.Context(), p, l, ns, &svc)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) listSvc(w http.ResponseWriter, r *http.Request, p, l, ns string) {
	svcs, _ := h.Backend.ListServices(r.Context(), p, l, ns)
	writeJSON(w, http.StatusOK, map[string]interface{}{"services": svcs})
}

func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
func writeError(w http.ResponseWriter, s int, m string) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}}) }
