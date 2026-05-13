package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/iam/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/iam/model"
)

type Handler struct {
	Backend backend.IAMBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[0] != "projects" || parts[2] != "serviceAccounts" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project := parts[1]
	parent := "projects/" + project

	if len(parts) == 3 {
		switch r.Method {
		case http.MethodPost:
			h.createServiceAccount(w, r, parent)
		case http.MethodGet:
			h.listServiceAccounts(w, r, project)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	saName := parts[3]
	fullName := parent + "/serviceAccounts/" + saName

	if strings.HasSuffix(saName, ":signJwt") {
		h.signJwt(w, r, parent+"/serviceAccounts/"+strings.TrimSuffix(saName, ":signJwt"))
		return
	}
	if strings.HasSuffix(saName, ":generateAccessToken") {
		h.generateAccessToken(w, r, parent+"/serviceAccounts/"+strings.TrimSuffix(saName, ":generateAccessToken"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getServiceAccount(w, r, fullName)
	case http.MethodDelete:
		h.deleteServiceAccount(w, r, fullName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) createServiceAccount(w http.ResponseWriter, r *http.Request, parent string) {
	var sa model.ServiceAccount
	json.NewDecoder(r.Body).Decode(&sa)
	if sa.Name == "" && sa.Email != "" {
		sa.Name = parent + "/serviceAccounts/" + sa.Email
	}
	result, _ := h.Backend.CreateServiceAccount(r.Context(), "", &sa)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getServiceAccount(w http.ResponseWriter, r *http.Request, name string) {
	sa, err := h.Backend.GetServiceAccount(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sa)
}

func (h *Handler) listServiceAccounts(w http.ResponseWriter, r *http.Request, project string) {
	sas, _ := h.Backend.ListServiceAccounts(r.Context(), project)
	writeJSON(w, http.StatusOK, map[string]interface{}{"accounts": sas})
}

func (h *Handler) deleteServiceAccount(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.DeleteServiceAccount(r.Context(), name)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) signJwt(w http.ResponseWriter, r *http.Request, name string) {
	var req model.SignJwtRequest
	json.NewDecoder(r.Body).Decode(&req)
	signed, _ := h.Backend.SignJwt(r.Context(), name, req.Payload)
	writeJSON(w, http.StatusOK, model.SignJwtResponse{KeyID: "emulator-key", SignedJwt: signed})
}

func (h *Handler) generateAccessToken(w http.ResponseWriter, r *http.Request, name string) {
	resp, _ := h.Backend.GenerateAccessToken(r.Context(), name)
	writeJSON(w, http.StatusOK, resp)
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
