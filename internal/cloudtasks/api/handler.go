package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/model"
)

type Handler struct {
	Backend backend.CloudTasksBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/")
	parts := strings.Split(path, "/")

	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "locations" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	project := parts[1]
	location := parts[3]
	base := "projects/" + project + "/locations/" + location

	if len(parts) >= 5 && parts[4] == "queues" {
		if len(parts) == 5 {
			switch r.Method {
			case http.MethodPost:
				h.createQueue(w, r, base)
			case http.MethodGet:
				h.listQueues(w, r, project, location)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		queueName := parts[5]
		fullName := base + "/queues/" + queueName

		if strings.HasSuffix(queueName, ":pause") {
			h.pauseQueue(w, r, fullName)
			return
		}
		if strings.HasSuffix(queueName, ":resume") {
			h.resumeQueue(w, r, fullName)
			return
		}
		if strings.HasSuffix(queueName, ":purge") {
			h.purgeQueue(w, r, fullName)
			return
		}

		if len(parts) >= 7 && parts[6] == "tasks" {
			if len(parts) == 7 {
				switch r.Method {
				case http.MethodPost:
					h.createTask(w, r, fullName)
				case http.MethodGet:
					h.listTasks(w, r, fullName)
				default:
					writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
				}
				return
			}

			taskName := parts[7]
			taskFull := fullName + "/tasks/" + taskName

			if strings.HasSuffix(taskName, ":run") {
				h.runTask(w, r, taskFull)
				return
			}

			switch r.Method {
			case http.MethodGet:
				h.getTask(w, r, taskFull)
			case http.MethodDelete:
				h.deleteTask(w, r, taskFull)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getQueue(w, r, fullName)
		case http.MethodPatch:
			h.updateQueue(w, r, fullName)
		case http.MethodDelete:
			h.deleteQueue(w, r, fullName)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createQueue(w http.ResponseWriter, r *http.Request, parent string) {
	var queue model.Queue
	if err := json.NewDecoder(r.Body).Decode(&queue); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if queue.Name == "" {
		queue.Name = parent + "/queues/" + r.URL.Query().Get("queueId")
	}
	result, err := h.Backend.CreateQueue(r.Context(), &queue)
	if err != nil {
		if err == backend.ErrQueueAlreadyExists {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getQueue(w http.ResponseWriter, r *http.Request, name string) {
	q, err := h.Backend.GetQueue(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handler) listQueues(w http.ResponseWriter, r *http.Request, project, location string) {
	queues, err := h.Backend.ListQueues(r.Context(), project, location)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"queues": queues})
}

func (h *Handler) updateQueue(w http.ResponseWriter, r *http.Request, name string) {
	var queue model.Queue
	if err := json.NewDecoder(r.Body).Decode(&queue); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.UpdateQueue(r.Context(), name, &queue)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteQueue(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.Backend.DeleteQueue(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) pauseQueue(w http.ResponseWriter, r *http.Request, name string) {
	q, err := h.Backend.PauseQueue(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handler) resumeQueue(w http.ResponseWriter, r *http.Request, name string) {
	q, err := h.Backend.ResumeQueue(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handler) purgeQueue(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.Backend.PurgeQueue(r.Context(), name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{})
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request, parent string) {
	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	task, err := h.Backend.CreateTask(r.Context(), parent, req.Task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, name string) {
	task, err := h.Backend.GetTask(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request, parent string) {
	tasks, err := h.Backend.ListTasks(r.Context(), parent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.Backend.DeleteTask(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) runTask(w http.ResponseWriter, r *http.Request, name string) {
	task, err := h.Backend.RunTask(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
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
