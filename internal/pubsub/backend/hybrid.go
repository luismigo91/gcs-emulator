package backend

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
)

type HybridPubSubBackend struct {
	*MemoryPubSubBackend
	storagePath   string
	flushInterval time.Duration
	stopCh        chan struct{}
	mu            sync.Mutex
	dirty         bool
}

func NewHybridPubSubBackend(storagePath string, flushInterval time.Duration) (*HybridPubSubBackend, error) {
	hb := &HybridPubSubBackend{
		MemoryPubSubBackend: NewMemoryPubSubBackend(),
		storagePath:         storagePath,
		flushInterval:       flushInterval,
		stopCh:              make(chan struct{}),
	}

	if err := hb.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	go hb.flushLoop()

	return hb, nil
}

func (hb *HybridPubSubBackend) markDirty() {
	hb.mu.Lock()
	hb.dirty = true
	hb.mu.Unlock()
}

func (hb *HybridPubSubBackend) flushLoop() {
	ticker := time.NewTicker(hb.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hb.mu.Lock()
			d := hb.dirty
			hb.dirty = false
			hb.mu.Unlock()
			if d {
				_ = hb.save()
			}
		case <-hb.stopCh:
			_ = hb.save()
			return
		}
	}
}

func (hb *HybridPubSubBackend) save() error {
	hb.MemoryPubSubBackend.mu.RLock()
	defer hb.MemoryPubSubBackend.mu.RUnlock()

	if err := os.MkdirAll(hb.storagePath, 0755); err != nil {
		return err
	}

	data := pubsubPersistentData{
		Topics:        hb.topics,
		Subscriptions: hb.subscriptions,
		Schemas:       hb.schemas,
		SchemaRevs:    hb.schemaRevisions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(hb.storagePath, "pubsub-data.json"), jsonData, 0644)
}

func (hb *HybridPubSubBackend) load() error {
	jsonData, err := os.ReadFile(filepath.Join(hb.storagePath, "pubsub-data.json"))
	if err != nil {
		return err
	}

	var data pubsubPersistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	hb.topics = data.Topics
	if hb.topics == nil {
		hb.topics = make(map[string]*model.Topic)
	}
	hb.subscriptions = data.Subscriptions
	if hb.subscriptions == nil {
		hb.subscriptions = make(map[string]*model.Subscription)
	}
	hb.schemas = data.Schemas
	if hb.schemas == nil {
		hb.schemas = make(map[string]*model.Schema)
	}
	hb.schemaRevisions = data.SchemaRevs
	if hb.schemaRevisions == nil {
		hb.schemaRevisions = make(map[string][]*model.Schema)
	}
	hb.messages = make(map[string][]*queuedMessage)

	return nil
}

func (hb *HybridPubSubBackend) CreateTopic(ctx context.Context, topic *model.Topic) (*model.Topic, error) {
	result, err := hb.MemoryPubSubBackend.CreateTopic(ctx, topic)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridPubSubBackend) DeleteTopic(ctx context.Context, project, name string) error {
	if err := hb.MemoryPubSubBackend.DeleteTopic(ctx, project, name); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridPubSubBackend) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	result, err := hb.MemoryPubSubBackend.CreateSubscription(ctx, sub)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridPubSubBackend) DeleteSubscription(ctx context.Context, project, name string) error {
	if err := hb.MemoryPubSubBackend.DeleteSubscription(ctx, project, name); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridPubSubBackend) CreateSchema(ctx context.Context, schema *model.Schema) (*model.Schema, error) {
	result, err := hb.MemoryPubSubBackend.CreateSchema(ctx, schema)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridPubSubBackend) DeleteSchema(ctx context.Context, project, name string) error {
	if err := hb.MemoryPubSubBackend.DeleteSchema(ctx, project, name); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridPubSubBackend) CommitSchema(ctx context.Context, project, name string, schema *model.Schema) (*model.Schema, error) {
	result, err := hb.MemoryPubSubBackend.CommitSchema(ctx, project, name, schema)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridPubSubBackend) Shutdown() error {
	close(hb.stopCh)
	return hb.save()
}
