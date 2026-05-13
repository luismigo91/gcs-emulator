package model

import "time"

type Topic struct {
	Name       string            `json:"name"`
	KMSKeyName string            `json:"kmsKeyName,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Schema     *SchemaRef        `json:"schemaSettings,omitempty"`
}

type SchemaRef struct {
	Schema         string `json:"schema"`
	Encoding       string `json:"encoding"`
	FirstRevisionID string `json:"firstRevisionId,omitempty"`
}

type Subscription struct {
	Name               string            `json:"name"`
	Topic              string            `json:"topic"`
	PushConfig         *PushConfig       `json:"pushConfig,omitempty"`
	AckDeadlineSeconds int               `json:"ackDeadlineSeconds"`
	Labels             map[string]string `json:"labels,omitempty"`
	EnableMessageOrdering bool           `json:"enableMessageOrdering,omitempty"`
	Filter             string            `json:"filter,omitempty"`
	DeadLetterPolicy   *DeadLetterPolicy `json:"deadLetterPolicy,omitempty"`
}

type PushConfig struct {
	PushEndpoint string            `json:"pushEndpoint,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type DeadLetterPolicy struct {
	DeadLetterTopic     string `json:"deadLetterTopic,omitempty"`
	MaxDeliveryAttempts int    `json:"maxDeliveryAttempts,omitempty"`
}

type PubSubMessage struct {
	Data        []byte            `json:"data,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	MessageID   string            `json:"messageId"`
	PublishTime time.Time         `json:"publishTime"`
	OrderingKey string            `json:"orderingKey,omitempty"`
}

type ReceivedMessage struct {
	AckID   string         `json:"ackId"`
	Message *PubSubMessage `json:"message"`
}

type PublishRequest struct {
	Messages []*PubSubMessage `json:"messages"`
}

type PublishResponse struct {
	MessageIDs []string `json:"messageIds"`
}

type PullRequest struct {
	MaxMessages       int  `json:"maxMessages"`
	ReturnImmediately bool `json:"returnImmediately"`
}

type PullResponse struct {
	ReceivedMessages []*ReceivedMessage `json:"receivedMessages"`
}

type AcknowledgeRequest struct {
	AckIDs []string `json:"ackIds"`
}

type ModifyAckDeadlineRequest struct {
	AckIDs            []string `json:"ackIds"`
	AckDeadlineSeconds int     `json:"ackDeadlineSeconds"`
}

type Schema struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Definition string `json:"definition"`
	RevisionID string `json:"revisionId,omitempty"`
}

type ValidateSchemaRequest struct {
	Schema *Schema `json:"schema"`
}

type ValidateSchemaResponse struct {
	Valid bool `json:"valid"`
}

type CommitSchemaRequest struct {
	Schema *Schema `json:"schema"`
}

type Snapshot struct {
	Name         string    `json:"name"`
	Topic        string    `json:"topic"`
	Subscription string    `json:"subscription,omitempty"`
	ExpireTime   time.Time `json:"expireTime,omitempty"`
}
