package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/artifactregistry/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/artifactregistry/model"
)

type Handler struct {
	Backend backend.ArtifactRegistryBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "repositories" {
		if len(parts) == 5 {
			switch r.Method {
			case http.MethodPost: h.createRepo(w, r, parent)
			case http.MethodGet: h.listRepos(w, r, parent)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		repoName := parts[5]
		fullName := parent + "/repositories/" + repoName
		if len(parts) >= 7 && parts[6] == "dockerImages" {
			switch r.Method {
			case http.MethodGet: h.listImages(w, r, fullName)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		switch r.Method {
		case http.MethodGet: h.getRepo(w, r, fullName)
		case http.MethodDelete: h.deleteRepo(w, r, fullName)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createRepo(w http.ResponseWriter, r *http.Request, parent string) {
	var repo model.Repository
	json.NewDecoder(r.Body).Decode(&repo)
	if repo.Name == "" { repo.Name = parent + "/repositories/" + r.URL.Query().Get("repositoryId") }
	result, _ := h.Backend.CreateRepository(r.Context(), &repo)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getRepo(w http.ResponseWriter, r *http.Request, name string) {
	repo, err := h.Backend.GetRepository(r.Context(), name)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, repo)
}
func (h *Handler) listRepos(w http.ResponseWriter, r *http.Request, parent string) {
	repos, _ := h.Backend.ListRepositories(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"repositories": repos})
}
func (h *Handler) deleteRepo(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.DeleteRepository(r.Context(), name)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) listImages(w http.ResponseWriter, r *http.Request, parent string) {
	images, _ := h.Backend.ListDockerImages(r.Context(), parent)
	writeJSON(w, http.StatusOK, map[string]interface{}{"dockerImages": images})
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
