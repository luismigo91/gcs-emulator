package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/dns/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/dns/model"
)

type Handler struct {
	Backend backend.DNSBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/dns/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "projects" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project := parts[1]

	if len(parts) >= 3 && parts[2] == "managedZones" {
		zone := ""
		if len(parts) >= 4 { zone = parts[3] }
		if len(parts) == 3 {
			switch r.Method {
			case http.MethodPost: h.createZone(w, r, project)
			case http.MethodGet: h.listZones(w, r, project)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		if len(parts) >= 5 && parts[4] == "rrsets" {
			h.listRecords(w, r, project, zone)
			return
		}
		if len(parts) >= 5 && parts[4] == "changes" {
			if r.Method == http.MethodPost {
				h.createChange(w, r, project, zone)
				return
			}
		}
		switch r.Method {
		case http.MethodGet: h.getZone(w, r, project, zone)
		case http.MethodDelete: h.deleteZone(w, r, project, zone)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createZone(w http.ResponseWriter, r *http.Request, project string) {
	var z model.ManagedZone
	json.NewDecoder(r.Body).Decode(&z)
	if z.Name == "" { z.Name = r.URL.Query().Get("managedZoneId") }
	result, _ := h.Backend.CreateZone(r.Context(), project, &z)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getZone(w http.ResponseWriter, r *http.Request, project, name string) {
	z, err := h.Backend.GetZone(r.Context(), project, name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, z)
}
func (h *Handler) listZones(w http.ResponseWriter, r *http.Request, project string) {
	zs, _ := h.Backend.ListZones(r.Context(), project)
	writeJSON(w, http.StatusOK, map[string]interface{}{"managedZones": zs})
}
func (h *Handler) deleteZone(w http.ResponseWriter, r *http.Request, project, name string) {
	h.Backend.DeleteZone(r.Context(), project, name)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) createChange(w http.ResponseWriter, r *http.Request, project, zone string) {
	var ch model.Change
	json.NewDecoder(r.Body).Decode(&ch)
	rrs, _ := h.Backend.CreateChange(r.Context(), project, zone, &ch)
	writeJSON(w, http.StatusOK, map[string]interface{}{"additions": rrs})
}
func (h *Handler) listRecords(w http.ResponseWriter, r *http.Request, project, zone string) {
	rrs, _ := h.Backend.ListRecordSets(r.Context(), project, zone)
	writeJSON(w, http.StatusOK, map[string]interface{}{"rrsets": rrs})
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
