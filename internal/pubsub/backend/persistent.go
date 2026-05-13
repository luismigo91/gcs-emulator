package backend

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
)

type pubsubPersistentData struct {
	Topics        map[string]*model.Topic             `json:"topics"`
	Subscriptions map[string]*model.Subscription      `json:"subscriptions"`
	Schemas       map[string]*model.Schema            `json:"schemas"`
	SchemaRevs    map[string][]*model.Schema          `json:"schemaRevisions"`
}

type PersistentPubSubBackend struct {
	*MemoryPubSubBackend
	mu          sync.Mutex
	storagePath string
}

func NewPersistentPubSubBackend(storagePath string) (*PersistentPubSubBackend, error) {
	pb := &PersistentPubSubBackend{
		MemoryPubSubBackend: NewMemoryPubSubBackend(),
		storagePath:         storagePath,
	}

	if err := pb.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return pb, nil
}

func (p *PersistentPubSubBackend) save() error {
	p.mu.Lock()
	p.MemoryPubSubBackend.mu.RLock()
	defer p.mu.Unlock()
	defer p.MemoryPubSubBackend.mu.RUnlock()

	if err := os.MkdirAll(p.storagePath, 0755); err != nil {
		return err
	}

	data := pubsubPersistentData{
		Topics:        p.topics,
		Subscriptions: p.subscriptions,
		Schemas:       p.schemas,
		SchemaRevs:    p.schemaRevisions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(p.storagePath, "pubsub-data.json"), jsonData, 0644)
}

func (p *PersistentPubSubBackend) load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dataFile := filepath.Join(p.storagePath, "pubsub-data.json")
	jsonData, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	var data pubsubPersistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	p.topics = data.Topics
	if p.topics == nil {
		p.topics = make(map[string]*model.Topic)
	}
	p.subscriptions = data.Subscriptions
	if p.subscriptions == nil {
		p.subscriptions = make(map[string]*model.Subscription)
	}
	p.schemas = data.Schemas
	if p.schemas == nil {
		p.schemas = make(map[string]*model.Schema)
	}
	p.schemaRevisions = data.SchemaRevs
	if p.schemaRevisions == nil {
		p.schemaRevisions = make(map[string][]*model.Schema)
	}
	p.messages = make(map[string][]*queuedMessage) // messages don't survive restart

	return nil
}

func (p *PersistentPubSubBackend) CreateTopic(ctx context.Context, topic *model.Topic) (*model.Topic, error) {
	result, err := p.MemoryPubSubBackend.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return result, p.save()
}

func (p *PersistentPubSubBackend) DeleteTopic(ctx context.Context, project, name string) error {
	if err := p.MemoryPubSubBackend.DeleteTopic(ctx, project, name); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentPubSubBackend) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	result, err := p.MemoryPubSubBackend.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}
	return result, p.save()
}

func (p *PersistentPubSubBackend) DeleteSubscription(ctx context.Context, project, name string) error {
	if err := p.MemoryPubSubBackend.DeleteSubscription(ctx, project, name); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentPubSubBackend) CreateSchema(ctx context.Context, schema *model.Schema) (*model.Schema, error) {
	result, err := p.MemoryPubSubBackend.CreateSchema(ctx, schema)
	if err != nil {
		return nil, err
	}
	return result, p.save()
}

func (p *PersistentPubSubBackend) DeleteSchema(ctx context.Context, project, name string) error {
	if err := p.MemoryPubSubBackend.DeleteSchema(ctx, project, name); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentPubSubBackend) CommitSchema(ctx context.Context, project, name string, schema *model.Schema) (*model.Schema, error) {
	result, err := p.MemoryPubSubBackend.CommitSchema(ctx, project, name, schema)
	if err != nil {
		return nil, err
	}
	return result, p.save()
}

func (p *PersistentPubSubBackend) Shutdown() error {
	return p.save()
}
