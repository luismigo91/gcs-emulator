package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/kms/model"
)

type Handler struct {
	Backend backend.KMSBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	parent := strings.Join(parts[:4], "/")

	if parts[4] == "keyRings" {
		if len(parts) == 5 {
			switch r.Method {
			case http.MethodPost:
				h.createKeyRing(w, r, parent)
			case http.MethodGet:
				h.listKeyRings(w, r, parent)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		krParent := parent + "/keyRings/" + parts[5]

		if len(parts) >= 7 && parts[6] == "cryptoKeys" {
			if len(parts) == 7 {
				switch r.Method {
				case http.MethodPost:
					h.createCryptoKey(w, r, krParent)
				case http.MethodGet:
					h.listCryptoKeys(w, r, krParent)
				default:
					writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
				}
				return
			}
			ckName := krParent + "/cryptoKeys/" + parts[7]
			if strings.HasSuffix(parts[7], ":encrypt") {
				h.encrypt(w, r, ckName)
				return
			}
			if strings.HasSuffix(parts[7], ":decrypt") {
				h.decrypt(w, r, ckName)
				return
			}

			if len(parts) >= 9 && parts[8] == "cryptoKeyVersions" {
				if len(parts) == 9 {
					switch r.Method {
					case http.MethodPost:
						h.createCryptoKeyVersion(w, r, ckName)
					case http.MethodGet:
						h.listCryptoKeyVersions(w, r, ckName)
					default:
						writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
					}
					return
				}
				vName := ckName + "/cryptoKeyVersions/" + parts[9]
				if strings.HasSuffix(parts[9], ":destroy") {
					h.destroyVersion(w, r, vName)
					return
				}
				if strings.HasSuffix(parts[9], ":restore") {
					h.restoreVersion(w, r, vName)
					return
				}
				h.getVersion(w, r, vName)
				return
			}

			switch r.Method {
			case http.MethodGet:
				h.getCryptoKey(w, r, ckName)
			case http.MethodPatch:
				h.updateCryptoKey(w, r, ckName)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getKeyRing(w, r, krParent)
		case http.MethodDelete:
			h.deleteKeyRing(w, r, krParent)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createKeyRing(w http.ResponseWriter, r *http.Request, parent string) {
	parts := strings.Split(parent, "/")
	project, location := parts[1], parts[3]
	name := r.URL.Query().Get("keyRingId")
	kr, err := h.Backend.CreateKeyRing(r.Context(), project, location, name)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	kr.Name = parent + "/keyRings/" + name
	writeJSON(w, http.StatusOK, kr)
}

func (h *Handler) getKeyRing(w http.ResponseWriter, r *http.Request, name string) {
	kr, err := h.Backend.GetKeyRing(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, kr)
}

func (h *Handler) listKeyRings(w http.ResponseWriter, r *http.Request, parent string) {
	parts := strings.Split(parent, "/")
	krs, _ := h.Backend.ListKeyRings(r.Context(), parts[1], parts[3])
	writeJSON(w, http.StatusOK, map[string]interface{}{"keyRings": krs})
}

func (h *Handler) deleteKeyRing(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.Backend.DeleteKeyRing(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createCryptoKey(w http.ResponseWriter, r *http.Request, parent string) {
	var key model.CryptoKey
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if key.Name == "" {
		key.Name = parent + "/cryptoKeys/" + r.URL.Query().Get("cryptoKeyId")
	}
	result, err := h.Backend.CreateCryptoKey(r.Context(), parent, &key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getCryptoKey(w http.ResponseWriter, r *http.Request, name string) {
	key, err := h.Backend.GetCryptoKey(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, key)
}

func (h *Handler) listCryptoKeys(w http.ResponseWriter, r *http.Request, parent string) {
	keys, _ := h.Backend.ListCryptoKeys(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"cryptoKeys": keys})
}

func (h *Handler) updateCryptoKey(w http.ResponseWriter, r *http.Request, name string) {
	var key model.CryptoKey
	json.NewDecoder(r.Body).Decode(&key)
	result, _ := h.Backend.UpdateCryptoKey(r.Context(), name, &key)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) createCryptoKeyVersion(w http.ResponseWriter, r *http.Request, parent string) {
	v, err := h.Backend.CreateCryptoKeyVersion(r.Context(), parent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) getVersion(w http.ResponseWriter, r *http.Request, name string) {
	v, err := h.Backend.GetCryptoKeyVersion(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) listCryptoKeyVersions(w http.ResponseWriter, r *http.Request, parent string) {
	vs, _ := h.Backend.ListCryptoKeyVersions(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"cryptoKeyVersions": vs})
}

func (h *Handler) destroyVersion(w http.ResponseWriter, r *http.Request, name string) {
	v, _ := h.Backend.DestroyCryptoKeyVersion(r.Context(), name)
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) restoreVersion(w http.ResponseWriter, r *http.Request, name string) {
	v, _ := h.Backend.RestoreCryptoKeyVersion(r.Context(), name)
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) encrypt(w http.ResponseWriter, r *http.Request, name string) {
	var req model.EncryptRequest
	json.NewDecoder(r.Body).Decode(&req)
	ct, _ := h.Backend.Encrypt(r.Context(), name, req.Plaintext)
	writeJSON(w, http.StatusOK, model.EncryptResponse{Ciphertext: ct})
}

func (h *Handler) decrypt(w http.ResponseWriter, r *http.Request, name string) {
	var req model.DecryptRequest
	json.NewDecoder(r.Body).Decode(&req)
	pt, _ := h.Backend.Decrypt(r.Context(), name, req.Ciphertext)
	writeJSON(w, http.StatusOK, model.DecryptResponse{Plaintext: pt})
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
