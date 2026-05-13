package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/scheduler/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/scheduler/model"
)

type Handler struct {
	Backend backend.SchedulerBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	parent := "projects/" + parts[1] + "/locations/" + parts[3]

	if len(parts) >= 5 && parts[4] == "jobs" {
		if len(parts) == 5 {
			switch r.Method {
			case http.MethodPost:
				h.createJob(w, r, parent)
			case http.MethodGet:
				h.listJobs(w, r, parent)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		jobName := parts[5]
		fullName := parent + "/jobs/" + jobName
		if strings.HasSuffix(jobName, ":pause") {
			h.pauseJob(w, r, fullName); return
		}
		if strings.HasSuffix(jobName, ":resume") {
			h.resumeJob(w, r, fullName); return
		}
		if strings.HasSuffix(jobName, ":run") {
			h.runJob(w, r, fullName); return
		}
		switch r.Method {
		case http.MethodGet:
			h.getJob(w, r, fullName)
		case http.MethodPatch:
			h.updateJob(w, r, fullName)
		case http.MethodDelete:
			h.deleteJob(w, r, fullName)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request, parent string) {
	var job model.Job
	json.NewDecoder(r.Body).Decode(&job)
	if job.Name == "" {
		job.Name = parent + "/jobs/" + r.URL.Query().Get("jobId")
	}
	result, _ := h.Backend.CreateJob(r.Context(), &job)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request, name string) {
	j, err := h.Backend.GetJob(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request, parent string) {
	parts := strings.Split(parent, "/")
	jobs, _ := h.Backend.ListJobs(r.Context(), parts[1], parts[3])
	writeJSON(w, http.StatusOK, map[string]interface{}{"jobs": jobs})
}

func (h *Handler) updateJob(w http.ResponseWriter, r *http.Request, name string) {
	var job model.Job
	json.NewDecoder(r.Body).Decode(&job)
	result, _ := h.Backend.UpdateJob(r.Context(), name, &job)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteJob(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.DeleteJob(r.Context(), name)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) pauseJob(w http.ResponseWriter, r *http.Request, name string) {
	j, _ := h.Backend.PauseJob(r.Context(), name)
	writeJSON(w, http.StatusOK, j)
}

func (h *Handler) resumeJob(w http.ResponseWriter, r *http.Request, name string) {
	j, _ := h.Backend.ResumeJob(r.Context(), name)
	writeJSON(w, http.StatusOK, j)
}

func (h *Handler) runJob(w http.ResponseWriter, r *http.Request, name string) {
	h.Backend.RunJob(r.Context(), name)
	w.WriteHeader(http.StatusOK)
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
