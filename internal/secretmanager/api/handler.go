package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/model"
)

type Handler struct {
	Backend backend.SecretManagerBackend
}

func (h *Handler) SecretsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[0] != "projects" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	project := parts[1]

	if len(parts) >= 4 && parts[2] == "secrets" {
		secretName := parts[3]

		if len(parts) >= 6 && parts[4] == "versions" {
			version := parts[5]
			if strings.HasSuffix(version, ":access") {
				h.accessVersion(w, r, project, secretName, strings.TrimSuffix(version, ":access"))
				return
			}
			if strings.HasSuffix(version, ":enable") {
				h.enableVersion(w, r, project, secretName, strings.TrimSuffix(version, ":enable"))
				return
			}
			if strings.HasSuffix(version, ":disable") {
				h.disableVersion(w, r, project, secretName, strings.TrimSuffix(version, ":disable"))
				return
			}
			if strings.HasSuffix(version, ":destroy") {
				h.destroyVersion(w, r, project, secretName, strings.TrimSuffix(version, ":destroy"))
				return
			}
			h.getVersion(w, r, project, secretName, version)
			return
		}

		if len(parts) >= 5 && parts[4] == "versions" {
			switch r.Method {
			case http.MethodGet:
				h.listVersions(w, r, project, secretName)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(secretName, ":addVersion") {
			h.addVersion(w, r, project, strings.TrimSuffix(secretName, ":addVersion"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getSecret(w, r, project, secretName)
		case http.MethodPatch:
			h.updateSecret(w, r, project, secretName)
		case http.MethodDelete:
			h.deleteSecret(w, r, project, secretName)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	if len(parts) == 3 && parts[2] == "secrets" {
		switch r.Method {
		case http.MethodPost:
			h.createSecret(w, r, project)
		case http.MethodGet:
			h.listSecrets(w, r, project)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createSecret(w http.ResponseWriter, r *http.Request, project string) {
	var secret model.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if secret.Name == "" {
		secret.Name = "projects/" + project + "/secrets/" + r.URL.Query().Get("secretId")
	}
	result, err := h.Backend.CreateSecret(r.Context(), &secret)
	if err != nil {
		if err == backend.ErrSecretAlreadyExists {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getSecret(w http.ResponseWriter, r *http.Request, project, name string) {
	secret, err := h.Backend.GetSecret(r.Context(), project, name)
	if err != nil {
		if err == backend.ErrSecretNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, secret)
}

func (h *Handler) listSecrets(w http.ResponseWriter, r *http.Request, project string) {
	secrets, err := h.Backend.ListSecrets(r.Context(), project)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"secrets": secrets})
}

func (h *Handler) updateSecret(w http.ResponseWriter, r *http.Request, project, name string) {
	var req model.UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	paths := []string{}
	if req.UpdateMask != nil {
		paths = req.UpdateMask.Paths
	}
	result, err := h.Backend.UpdateSecret(r.Context(), project, name, req.Secret, paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteSecret(w http.ResponseWriter, r *http.Request, project, name string) {
	if err := h.Backend.DeleteSecret(r.Context(), project, name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) addVersion(w http.ResponseWriter, r *http.Request, project, name string) {
	var req model.AddVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Payload == nil {
		writeError(w, http.StatusBadRequest, "Payload is required")
		return
	}
	version, err := h.Backend.AddVersion(r.Context(), project, name, req.Payload.Data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, version)
}

func (h *Handler) listVersions(w http.ResponseWriter, r *http.Request, project, name string) {
	versions, err := h.Backend.ListVersions(r.Context(), project, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"versions": versions})
}

func (h *Handler) getVersion(w http.ResponseWriter, r *http.Request, project, name, version string) {
	v, err := h.Backend.GetVersion(r.Context(), project, name, version)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) accessVersion(w http.ResponseWriter, r *http.Request, project, name, version string) {
	data, err := h.Backend.AccessVersion(r.Context(), project, name, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.AccessResponse{
		Name: secretKey(project, name) + "/versions/" + version,
		Payload: &model.Payload{Data: data},
	})
}

func (h *Handler) enableVersion(w http.ResponseWriter, r *http.Request, project, name, version string) {
	v, err := h.Backend.EnableVersion(r.Context(), project, name, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) disableVersion(w http.ResponseWriter, r *http.Request, project, name, version string) {
	v, err := h.Backend.DisableVersion(r.Context(), project, name, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) destroyVersion(w http.ResponseWriter, r *http.Request, project, name, version string) {
	v, err := h.Backend.DestroyVersion(r.Context(), project, name, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func secretKey(project, name string) string {
	return "projects/" + project + "/secrets/" + name
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
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}
