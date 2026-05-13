package backend

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
)

var (
	ErrTopicNotFound        = errors.New("topic not found")
	ErrTopicAlreadyExists   = errors.New("topic already exists")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSchemaNotFound       = errors.New("schema not found")
	ErrSchemaInvalid        = errors.New("schema validation failed")
)

type queuedMessage struct {
	msg              *model.PubSubMessage
	ackID            string
	delivered        time.Time
	deadline         time.Duration
	deliveryAttempts int
}

type MemoryPubSubBackend struct {
	mu            sync.RWMutex
	topics        map[string]*model.Topic
	subscriptions map[string]*model.Subscription
	messages      map[string][]*queuedMessage
	schemas       map[string]*model.Schema
	schemaRevisions map[string][]*model.Schema
	snapshots     map[string]*model.Snapshot
	msgIDCounter  int64
	ackIDCounter  int64
}

func NewMemoryPubSubBackend() *MemoryPubSubBackend {
	return &MemoryPubSubBackend{
		topics:          make(map[string]*model.Topic),
		subscriptions:   make(map[string]*model.Subscription),
		messages:        make(map[string][]*queuedMessage),
		schemas:         make(map[string]*model.Schema),
		schemaRevisions: make(map[string][]*model.Schema),
		snapshots:       make(map[string]*model.Snapshot),
	}
}

func topicKey(project, name string) string {
	return "projects/" + project + "/topics/" + name
}

func subscriptionKey(project, name string) string {
	return "projects/" + project + "/subscriptions/" + name
}

func schemaKey(project, name string) string {
	return "projects/" + project + "/schemas/" + name
}

func (m *MemoryPubSubBackend) resName(project, resource, name string) string {
	return "projects/" + project + "/" + resource + "/" + name
}

func (m *MemoryPubSubBackend) CreateTopic(ctx context.Context, topic *model.Topic) (*model.Topic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.topics[topic.Name]; exists {
		return nil, ErrTopicAlreadyExists
	}
	m.topics[topic.Name] = topic
	m.messages[topic.Name] = make([]*queuedMessage, 0)
	return topic, nil
}

func (m *MemoryPubSubBackend) GetTopic(ctx context.Context, project, name string) (*model.Topic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := topicKey(project, name)
	t, exists := m.topics[key]
	if !exists {
		return nil, ErrTopicNotFound
	}
	return t, nil
}

func (m *MemoryPubSubBackend) ListTopics(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Topic, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := "projects/" + project + "/"
	var result []*model.Topic
	for name, t := range m.topics {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			result = append(result, t)
		}
	}

	nextToken := ""
	if pageSize > 0 && len(result) > pageSize {
		result = result[:pageSize]
		nextToken = result[len(result)-1].Name
	}

	return result, nextToken, nil
}

func (m *MemoryPubSubBackend) DeleteTopic(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := topicKey(project, name)
	if _, exists := m.topics[key]; !exists {
		return ErrTopicNotFound
	}
	delete(m.topics, key)
	delete(m.messages, key)
	return nil
}

func (m *MemoryPubSubBackend) Publish(ctx context.Context, project, topic string, messages []*model.PubSubMessage) ([]string, error) {
	m.mu.Lock()
	ids, err := m.publishLocked(project, topic, messages)

	pushTargets := make(map[string]string)
	for subKey, sub := range m.subscriptions {
		if sub.Topic == topicKey(project, topic) && sub.PushConfig != nil && sub.PushConfig.PushEndpoint != "" {
			pushTargets[subKey] = sub.PushConfig.PushEndpoint
		}
	}
	m.mu.Unlock()

	if err != nil {
		return nil, err
	}

	for _, msg := range messages {
		pushBody, _ := json.Marshal(map[string]interface{}{
			"message": map[string]interface{}{
				"data":       base64.StdEncoding.EncodeToString(msg.Data),
				"messageId":  msg.MessageID,
				"attributes": msg.Attributes,
			},
			"subscription": "",
		})
		for _, endpoint := range pushTargets {
			go func(ep string, body []byte) {
				http.Post(ep, "application/json", bytes.NewReader(body))
			}(endpoint, pushBody)
		}
	}

	return ids, nil
}

func (m *MemoryPubSubBackend) publishLocked(project, topic string, messages []*model.PubSubMessage) ([]string, error) {
	key := topicKey(project, topic)
	if _, exists := m.topics[key]; !exists {
		return nil, ErrTopicNotFound
	}

	now := time.Now()
	msgIDs := make([]string, len(messages))

	for i, pubMsg := range messages {
		m.msgIDCounter++
		msgID := fmt.Sprintf("%d", m.msgIDCounter)
		pubMsg.MessageID = msgID
		pubMsg.PublishTime = now
		msgIDs[i] = msgID

		qm := &queuedMessage{
			msg:              pubMsg,
			delivered:        now,
			deadline:         10 * time.Second,
			deliveryAttempts: 0,
		}

		for subKey, sub := range m.subscriptions {
			if sub.Topic != key {
				continue
			}

			if sub.Filter != "" && !matchFilter(sub.Filter, pubMsg.Attributes) {
				continue
			}

			m.ackIDCounter++
			qmCopy := *qm
			qmCopy.ackID = fmt.Sprintf("ack-%d", m.ackIDCounter)
			if sub.AckDeadlineSeconds > 0 {
				qmCopy.deadline = time.Duration(sub.AckDeadlineSeconds) * time.Second
			}
			m.messages[subKey] = append(m.messages[subKey], &qmCopy)
		}
	}

	return msgIDs, nil
}

func (m *MemoryPubSubBackend) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[sub.Name]; exists {
		return nil, errors.New("subscription already exists")
	}
	if _, exists := m.topics[sub.Topic]; !exists {
		return nil, ErrTopicNotFound
	}
	if sub.AckDeadlineSeconds == 0 {
		sub.AckDeadlineSeconds = 10
	}
	m.subscriptions[sub.Name] = sub
	if m.messages[sub.Name] == nil {
		m.messages[sub.Name] = make([]*queuedMessage, 0)
	}
	return sub, nil
}

func (m *MemoryPubSubBackend) GetSubscription(ctx context.Context, project, name string) (*model.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := subscriptionKey(project, name)
	sub, exists := m.subscriptions[key]
	if !exists {
		return nil, ErrSubscriptionNotFound
	}
	return sub, nil
}

func (m *MemoryPubSubBackend) ListSubscriptions(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Subscription, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := "projects/" + project + "/"
	var result []*model.Subscription
	for name, s := range m.subscriptions {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			result = append(result, s)
		}
	}

	nextToken := ""
	if pageSize > 0 && len(result) > pageSize {
		result = result[:pageSize]
		nextToken = result[len(result)-1].Name
	}

	return result, nextToken, nil
}

func (m *MemoryPubSubBackend) DeleteSubscription(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := subscriptionKey(project, name)
	if _, exists := m.subscriptions[key]; !exists {
		return ErrSubscriptionNotFound
	}
	delete(m.subscriptions, key)
	delete(m.messages, key)
	return nil
}

func (m *MemoryPubSubBackend) Pull(ctx context.Context, project, subscription string, maxMessages int) ([]*model.ReceivedMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := subscriptionKey(project, subscription)
	sub, subExists := m.subscriptions[key]
	if !subExists {
		return nil, ErrSubscriptionNotFound
	}

	queue := m.messages[key]

	var result []*model.ReceivedMessage
	count := 0
	var deadLetterMsgs []*model.PubSubMessage
	remaining := make([]*queuedMessage, 0, len(queue))

	for _, qm := range queue {
		now := time.Now()
		if now.After(qm.delivered.Add(qm.deadline)) {
			qm.deliveryAttempts++
			qm.delivered = now
			if sub.DeadLetterPolicy != nil && sub.DeadLetterPolicy.MaxDeliveryAttempts > 0 && qm.deliveryAttempts >= sub.DeadLetterPolicy.MaxDeliveryAttempts {
				deadLetterMsgs = append(deadLetterMsgs, qm.msg)
				continue
			}
		}
		if maxMessages > 0 && count >= maxMessages {
			remaining = append(remaining, qm)
			continue
		}
		result = append(result, &model.ReceivedMessage{
			AckID:   qm.ackID,
			Message: qm.msg,
		})
		count++
		remaining = append(remaining, qm)
	}
	m.messages[key] = remaining

	if len(deadLetterMsgs) > 0 && sub.DeadLetterPolicy != nil {
		dlTopic := sub.DeadLetterPolicy.DeadLetterTopic
		dlTopicParts := strings.Split(dlTopic, "/")
		if len(dlTopicParts) >= 4 {
			dlProject := dlTopicParts[1]
			dlName := dlTopicParts[3]
			if _, exists := m.topics[topicKey(dlProject, dlName)]; exists {
				m.publishLocked(dlProject, dlName, deadLetterMsgs)
			}
		}
	}

	return result, nil
}

func (m *MemoryPubSubBackend) Acknowledge(ctx context.Context, project, subscription string, ackIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := subscriptionKey(project, subscription)
	if _, exists := m.subscriptions[key]; !exists {
		return ErrSubscriptionNotFound
	}

	ackSet := make(map[string]bool, len(ackIDs))
	for _, id := range ackIDs {
		ackSet[id] = true
	}

	queue := m.messages[key]
	remaining := make([]*queuedMessage, 0, len(queue))
	for _, qm := range queue {
		if !ackSet[qm.ackID] {
			remaining = append(remaining, qm)
		}
	}
	m.messages[key] = remaining

	return nil
}

func (m *MemoryPubSubBackend) ModifyAckDeadline(ctx context.Context, project, subscription string, ackIDs []string, seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := subscriptionKey(project, subscription)
	if _, exists := m.subscriptions[key]; !exists {
		return ErrSubscriptionNotFound
	}

	ackSet := make(map[string]bool, len(ackIDs))
	for _, id := range ackIDs {
		ackSet[id] = true
	}

	for _, qm := range m.messages[key] {
		if ackSet[qm.ackID] {
			qm.deadline = time.Duration(seconds) * time.Second
			qm.delivered = time.Now()
		}
	}
	return nil
}

func (m *MemoryPubSubBackend) CreateSchema(ctx context.Context, schema *model.Schema) (*model.Schema, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := schema.Name
	if _, exists := m.schemas[key]; exists {
		return nil, errors.New("schema already exists")
	}
	schema.RevisionID = "1"
	m.schemas[key] = schema
	m.schemaRevisions[key] = append(m.schemaRevisions[key], schema)
	return schema, nil
}

func (m *MemoryPubSubBackend) GetSchema(ctx context.Context, project, name string) (*model.Schema, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := schemaKey(project, name)
	s, exists := m.schemas[key]
	if !exists {
		return nil, ErrSchemaNotFound
	}
	return s, nil
}

func (m *MemoryPubSubBackend) ListSchemas(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Schema, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix := "projects/" + project + "/"
	var result []*model.Schema
	for name, s := range m.schemas {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			result = append(result, s)
		}
	}

	nextToken := ""
	if pageSize > 0 && len(result) > pageSize {
		result = result[:pageSize]
		nextToken = result[len(result)-1].Name
	}

	return result, nextToken, nil
}

func (m *MemoryPubSubBackend) DeleteSchema(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := schemaKey(project, name)
	if _, exists := m.schemas[key]; !exists {
		return ErrSchemaNotFound
	}
	delete(m.schemas, key)
	delete(m.schemaRevisions, key)
	return nil
}

func (m *MemoryPubSubBackend) ValidateSchema(ctx context.Context, schema *model.Schema) error {
	if schema.Name == "" || schema.Definition == "" {
		return ErrSchemaInvalid
	}
	return nil
}

func (m *MemoryPubSubBackend) CommitSchema(ctx context.Context, project, name string, schema *model.Schema) (*model.Schema, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := schemaKey(project, name)
	if _, exists := m.schemas[key]; !exists {
		return nil, ErrSchemaNotFound
	}

	rev := len(m.schemaRevisions[key]) + 1
	schema.RevisionID = fmt.Sprintf("%d", rev)
	m.schemas[key] = schema
	m.schemaRevisions[key] = append(m.schemaRevisions[key], schema)
	return schema, nil
}

func (m *MemoryPubSubBackend) CreateSnapshot(ctx context.Context, snapshot, subscription string) (*model.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[snapshot]; exists {
		return nil, errors.New("snapshot already exists")
	}
	sub, exists := m.subscriptions[subscription]
	if !exists {
		return nil, ErrSubscriptionNotFound
	}

	s := &model.Snapshot{
		Name:         snapshot,
		Topic:        sub.Topic,
		Subscription: subscription,
	}
	m.snapshots[snapshot] = s
	return s, nil
}

func (m *MemoryPubSubBackend) GetSnapshot(ctx context.Context, project, name string) (*model.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := "projects/" + project + "/snapshots/" + name
	s, exists := m.snapshots[key]
	if !exists {
		return nil, errors.New("snapshot not found")
	}
	return s, nil
}

func (m *MemoryPubSubBackend) DeleteSnapshot(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := "projects/" + project + "/snapshots/" + name
	if _, exists := m.snapshots[key]; !exists {
		return errors.New("snapshot not found")
	}
	delete(m.snapshots, key)
	return nil
}

func (m *MemoryPubSubBackend) Seek(ctx context.Context, project, subscription string, snapshot string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := subscriptionKey(project, subscription)
	if _, exists := m.subscriptions[key]; !exists {
		return ErrSubscriptionNotFound
	}
	snapshotKey := "projects/" + project + "/snapshots/" + snapshot
	if _, exists := m.snapshots[snapshotKey]; !exists {
		return errors.New("snapshot not found")
	}
	m.messages[key] = make([]*queuedMessage, 0)
	return nil
}

func (m *MemoryPubSubBackend) Shutdown() error {
	return nil
}

func matchFilter(filter string, attrs map[string]string) bool {
	if filter == "" {
		return true
	}
	filter = strings.TrimSpace(filter)
	if strings.HasPrefix(filter, "attributes.") {
		filter = strings.TrimPrefix(filter, "attributes.")
	}
	parts := strings.SplitN(filter, "=", 2)
	if len(parts) != 2 {
		parts = strings.SplitN(filter, "!=", 2)
		if len(parts) != 2 {
			return true
		}
		key := strings.Trim(parts[0], ` "'`)
		val := strings.Trim(parts[1], ` "'`)
		return attrs[key] != val
	}
	key := strings.Trim(parts[0], ` "'`)
	val := strings.Trim(parts[1], ` "'`)
	return attrs[key] == val
}
