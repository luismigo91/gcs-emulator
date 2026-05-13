package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
)

type PubSubBackend interface {
	CreateTopic(ctx context.Context, topic *model.Topic) (*model.Topic, error)
	GetTopic(ctx context.Context, project, name string) (*model.Topic, error)
	ListTopics(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Topic, string, error)
	DeleteTopic(ctx context.Context, project, name string) error

	Publish(ctx context.Context, project, topic string, messages []*model.PubSubMessage) ([]string, error)

	CreateSubscription(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error)
	GetSubscription(ctx context.Context, project, name string) (*model.Subscription, error)
	ListSubscriptions(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Subscription, string, error)
	DeleteSubscription(ctx context.Context, project, name string) error

	Pull(ctx context.Context, project, subscription string, maxMessages int) ([]*model.ReceivedMessage, error)
	Acknowledge(ctx context.Context, project, subscription string, ackIDs []string) error
	ModifyAckDeadline(ctx context.Context, project, subscription string, ackIDs []string, seconds int) error

	CreateSnapshot(ctx context.Context, snapshot, subscription string) (*model.Snapshot, error)
	GetSnapshot(ctx context.Context, project, name string) (*model.Snapshot, error)
	DeleteSnapshot(ctx context.Context, project, name string) error
	Seek(ctx context.Context, project, subscription string, snapshot string) error

	CreateSchema(ctx context.Context, schema *model.Schema) (*model.Schema, error)
	GetSchema(ctx context.Context, project, name string) (*model.Schema, error)
	ListSchemas(ctx context.Context, project string, pageSize int, pageToken string) ([]*model.Schema, string, error)
	DeleteSchema(ctx context.Context, project, name string) error
	ValidateSchema(ctx context.Context, schema *model.Schema) error
	CommitSchema(ctx context.Context, project, name string, schema *model.Schema) (*model.Schema, error)

	Shutdown() error
}
