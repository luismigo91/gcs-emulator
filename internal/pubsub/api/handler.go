package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/util"
)

type Handler struct {
	Backend backend.PubSubBackend
}

func (h *Handler) TopicHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[0] != "projects" || parts[2] != "topics" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	project := parts[1]
	topicName := ""
	isPublish := false

	if len(parts) >= 4 {
		last := parts[len(parts)-1]
		if strings.HasSuffix(last, ":publish") {
			topicName = strings.TrimSuffix(parts[3], ":publish")
			isPublish = true
		} else {
			topicName = parts[3]
		}
	}

	switch {
	case isPublish && r.Method == http.MethodPost:
		h.publishHandler(w, r, project, topicName)
	case topicName == "" && r.Method == http.MethodGet:
		h.listTopics(w, r, project)
	case topicName == "" && r.Method == http.MethodPut:
		h.createTopic(w, r, project)
	case topicName != "" && r.Method == http.MethodGet:
		h.getTopic(w, r, project, topicName)
	case topicName != "" && r.Method == http.MethodPut:
		h.createTopic(w, r, project)
	case topicName != "" && r.Method == http.MethodDelete:
		h.deleteTopic(w, r, project, topicName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) publishHandler(w http.ResponseWriter, r *http.Request, project, topic string) {
	var req model.PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ids, err := h.Backend.Publish(r.Context(), project, topic, req.Messages)
	if err != nil {
		if err == backend.ErrTopicNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.PublishResponse{MessageIDs: ids})
}

func (h *Handler) createTopic(w http.ResponseWriter, r *http.Request, project string) {
	var topic model.Topic
	if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if topic.Name == "" {
		writeError(w, http.StatusBadRequest, "Topic name is required")
		return
	}

	result, err := h.Backend.CreateTopic(r.Context(), &topic)
	if err != nil {
		if err == backend.ErrTopicAlreadyExists {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getTopic(w http.ResponseWriter, r *http.Request, project, name string) {
	topic, err := h.Backend.GetTopic(r.Context(), project, name)
	if err != nil {
		if err == backend.ErrTopicNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, topic)
}

func (h *Handler) listTopics(w http.ResponseWriter, r *http.Request, project string) {
	topics, nextToken, err := h.Backend.ListTopics(r.Context(), project, 0, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{"topics": topics}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteTopic(w http.ResponseWriter, r *http.Request, project, name string) {
	if err := h.Backend.DeleteTopic(r.Context(), project, name); err != nil {
		if err == backend.ErrTopicNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[0] != "projects" || parts[2] != "subscriptions" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	project := parts[1]
	subName := ""
	isPull := false
	isAck := false
	isModifyAck := false

	if len(parts) >= 4 {
		last := parts[len(parts)-1]
		if strings.HasSuffix(last, ":pull") {
			subName = strings.TrimSuffix(parts[3], ":pull")
			isPull = true
		} else if strings.HasSuffix(last, ":acknowledge") {
			subName = strings.TrimSuffix(parts[3], ":acknowledge")
			isAck = true
		} else if strings.HasSuffix(last, ":modifyAckDeadline") {
			subName = strings.TrimSuffix(parts[3], ":modifyAckDeadline")
			isModifyAck = true
		} else {
			subName = parts[3]
		}
	}

	switch {
	case isPull:
		h.pullHandler(w, r, project, subName)
	case isAck:
		h.acknowledgeHandler(w, r, project, subName)
	case isModifyAck:
		h.modifyAckDeadlineHandler(w, r, project, subName)
	case subName == "" && r.Method == http.MethodGet:
		h.listSubscriptions(w, r, project)
	case subName != "" && r.Method == http.MethodGet:
		h.getSubscription(w, r, project, subName)
	case subName != "" && r.Method == http.MethodPut:
		h.createSubscription(w, r, project)
	case subName != "" && r.Method == http.MethodDelete:
		h.deleteSubscription(w, r, project, subName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) pullHandler(w http.ResponseWriter, r *http.Request, project, subscription string) {
	var req model.PullRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.MaxMessages == 0 {
		req.MaxMessages = 10
	}

	msgs, err := h.Backend.Pull(r.Context(), project, subscription, req.MaxMessages)
	if err != nil {
		if err == backend.ErrSubscriptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if msgs == nil {
		msgs = []*model.ReceivedMessage{}
	}
	writeJSON(w, http.StatusOK, model.PullResponse{ReceivedMessages: msgs})
}

func (h *Handler) acknowledgeHandler(w http.ResponseWriter, r *http.Request, project, subscription string) {
	var req model.AcknowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.Backend.Acknowledge(r.Context(), project, subscription, req.AckIDs); err != nil {
		if err == backend.ErrSubscriptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) modifyAckDeadlineHandler(w http.ResponseWriter, r *http.Request, project, subscription string) {
	var req model.ModifyAckDeadlineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.Backend.ModifyAckDeadline(r.Context(), project, subscription, req.AckIDs, req.AckDeadlineSeconds); err != nil {
		if err == backend.ErrSubscriptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request, project string) {
	var sub model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if sub.Name == "" {
		writeError(w, http.StatusBadRequest, "Subscription name is required")
		return
	}
	if sub.Topic == "" {
		writeError(w, http.StatusBadRequest, "Topic is required")
		return
	}

	result, err := h.Backend.CreateSubscription(r.Context(), &sub)
	if err != nil {
		if err == backend.ErrTopicNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request, project, name string) {
	sub, err := h.Backend.GetSubscription(r.Context(), project, name)
	if err != nil {
		if err == backend.ErrSubscriptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request, project string) {
	subs, nextToken, err := h.Backend.ListSubscriptions(r.Context(), project, 0, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{"subscriptions": subs}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request, project, name string) {
	if err := h.Backend.DeleteSubscription(r.Context(), project, name); err != nil {
		if err == backend.ErrSubscriptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SchemaHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[0] != "projects" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	project := parts[1]

	if len(parts) >= 3 && parts[2] == "schemas" {
		if len(parts) == 3 {
			switch r.Method {
			case http.MethodPost:
				h.createSchema(w, r, project)
			case http.MethodGet:
				h.listSchemas(w, r, project)
			default:
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		schemaName := parts[3]
		if strings.HasSuffix(schemaName, ":validate") {
			h.validateSchema(w, r, project)
			return
		}
		if strings.HasSuffix(schemaName, ":commit") {
			h.commitSchema(w, r, project, strings.TrimSuffix(schemaName, ":commit"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getSchema(w, r, project, schemaName)
		case http.MethodDelete:
			h.deleteSchema(w, r, project, schemaName)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createSchema(w http.ResponseWriter, r *http.Request, project string) {
	var schema model.Schema
	if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.Backend.CreateSchema(r.Context(), &schema)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getSchema(w http.ResponseWriter, r *http.Request, project, name string) {
	schema, err := h.Backend.GetSchema(r.Context(), project, name)
	if err != nil {
		if err == backend.ErrSchemaNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schema)
}

func (h *Handler) listSchemas(w http.ResponseWriter, r *http.Request, project string) {
	schemas, nextToken, err := h.Backend.ListSchemas(r.Context(), project, 0, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{"schemas": schemas}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteSchema(w http.ResponseWriter, r *http.Request, project, name string) {
	if err := h.Backend.DeleteSchema(r.Context(), project, name); err != nil {
		if err == backend.ErrSchemaNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) validateSchema(w http.ResponseWriter, r *http.Request, project string) {
	var req model.ValidateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.Backend.ValidateSchema(r.Context(), req.Schema); err != nil {
		writeJSON(w, http.StatusOK, model.ValidateSchemaResponse{Valid: false})
		return
	}
	writeJSON(w, http.StatusOK, model.ValidateSchemaResponse{Valid: true})
}

func (h *Handler) commitSchema(w http.ResponseWriter, r *http.Request, project, name string) {
	var req model.CommitSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.Backend.CommitSchema(r.Context(), project, name, req.Schema)
	if err != nil {
		if err == backend.ErrSchemaNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) SnapshotHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "projects" || parts[2] != "snapshots" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	project := parts[1]
	snapName := ""
	if len(parts) >= 4 {
		snapName = parts[3]
	}
	switch {
	case snapName == "" && r.Method == http.MethodPost:
		h.createSnapshot(w, r, project)
	case snapName != "" && r.Method == http.MethodPut:
		h.createSnapshotNamed(w, r, project, snapName)
	case snapName != "" && r.Method == http.MethodGet:
		h.getSnapshot(w, r, project, snapName)
	case snapName != "" && r.Method == http.MethodDelete:
		h.deleteSnapshot(w, r, project, snapName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) createSnapshot(w http.ResponseWriter, r *http.Request, project string) {
	var snap model.Snapshot
	json.NewDecoder(r.Body).Decode(&snap)
	if snap.Name == "" {
		snap.Name = r.URL.Query().Get("snapshotId")
	}
	result, err := h.Backend.CreateSnapshot(r.Context(), snap.Name, snap.Subscription)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) createSnapshotNamed(w http.ResponseWriter, r *http.Request, project, name string) {
	var snap model.Snapshot
	json.NewDecoder(r.Body).Decode(&snap)
	snap.Name = "projects/" + project + "/snapshots/" + name
	if snap.Subscription == "" {
		snap.Subscription = r.URL.Query().Get("subscription")
	}
	result, err := h.Backend.CreateSnapshot(r.Context(), snap.Name, snap.Subscription)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getSnapshot(w http.ResponseWriter, r *http.Request, project, name string) {
	s, err := h.Backend.GetSnapshot(r.Context(), project, name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) deleteSnapshot(w http.ResponseWriter, r *http.Request, project, name string) {
	if err := h.Backend.DeleteSnapshot(r.Context(), project, name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
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
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}

var _ = util.GetProjectID
