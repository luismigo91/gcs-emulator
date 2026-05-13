package backend

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
)

type pubsubWALOperation struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type WALPubSubBackend struct {
	*MemoryPubSubBackend
	storagePath      string
	walFile          *os.File
	walWriter        *bufio.Writer
	mu               sync.Mutex
	walSize          int64
	compactThreshold int64
}

func NewWALPubSubBackend(storagePath string) (*WALPubSubBackend, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, err
	}

	wb := &WALPubSubBackend{
		MemoryPubSubBackend: NewMemoryPubSubBackend(),
		storagePath:         storagePath,
		compactThreshold:    10 * 1024 * 1024,
	}

	walPath := filepath.Join(storagePath, "pubsub-operations.wal")
	f, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	wb.walFile = f
	wb.walWriter = bufio.NewWriter(f)

	if info, err := f.Stat(); err == nil {
		wb.walSize = info.Size()
	}

	if err := wb.replay(); err != nil {
		f.Close()
		return nil, err
	}

	if err := wb.loadSnapshot(); err != nil {
		if !os.IsNotExist(err) {
			f.Close()
			return nil, err
		}
	}

	return wb, nil
}

func (wb *WALPubSubBackend) appendOperation(op *pubsubWALOperation) error {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	data, err := json.Marshal(op)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := wb.walWriter.Write(data); err != nil {
		return err
	}
	if err := wb.walWriter.Flush(); err != nil {
		return err
	}
	wb.walSize += int64(len(data))
	if wb.walSize >= wb.compactThreshold {
		return wb.compact()
	}
	return nil
}

func (wb *WALPubSubBackend) CreateTopic(ctx context.Context, topic *model.Topic) (*model.Topic, error) {
	result, err := wb.MemoryPubSubBackend.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	op := &pubsubWALOperation{
		Type:      "create_topic",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"topic": topic},
	}
	if err := wb.appendOperation(op); err != nil {
		return result, err
	}
	return result, nil
}

func (wb *WALPubSubBackend) DeleteTopic(ctx context.Context, project, name string) error {
	if err := wb.MemoryPubSubBackend.DeleteTopic(ctx, project, name); err != nil {
		return err
	}
	op := &pubsubWALOperation{
		Type:      "delete_topic",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"project": project, "name": name},
	}
	return wb.appendOperation(op)
}

func (wb *WALPubSubBackend) CreateSubscription(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	result, err := wb.MemoryPubSubBackend.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}
	op := &pubsubWALOperation{
		Type:      "create_subscription",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"subscription": sub},
	}
	if err := wb.appendOperation(op); err != nil {
		return result, err
	}
	return result, nil
}

func (wb *WALPubSubBackend) DeleteSubscription(ctx context.Context, project, name string) error {
	if err := wb.MemoryPubSubBackend.DeleteSubscription(ctx, project, name); err != nil {
		return err
	}
	op := &pubsubWALOperation{
		Type:      "delete_subscription",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"project": project, "name": name},
	}
	return wb.appendOperation(op)
}

func (wb *WALPubSubBackend) CreateSchema(ctx context.Context, schema *model.Schema) (*model.Schema, error) {
	result, err := wb.MemoryPubSubBackend.CreateSchema(ctx, schema)
	if err != nil {
		return nil, err
	}
	op := &pubsubWALOperation{
		Type:      "create_schema",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"schema": schema},
	}
	if err := wb.appendOperation(op); err != nil {
		return result, err
	}
	return result, nil
}

func (wb *WALPubSubBackend) DeleteSchema(ctx context.Context, project, name string) error {
	if err := wb.MemoryPubSubBackend.DeleteSchema(ctx, project, name); err != nil {
		return err
	}
	op := &pubsubWALOperation{
		Type:      "delete_schema",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"project": project, "name": name},
	}
	return wb.appendOperation(op)
}

func (wb *WALPubSubBackend) CommitSchema(ctx context.Context, project, name string, schema *model.Schema) (*model.Schema, error) {
	result, err := wb.MemoryPubSubBackend.CommitSchema(ctx, project, name, schema)
	if err != nil {
		return nil, err
	}
	op := &pubsubWALOperation{
		Type:      "commit_schema",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"project": project, "name": name, "schema": schema},
	}
	if err := wb.appendOperation(op); err != nil {
		return result, err
	}
	return result, nil
}

func (wb *WALPubSubBackend) Shutdown() error {
	if err := wb.compact(); err != nil {
		return err
	}
	return wb.walFile.Close()
}

func (wb *WALPubSubBackend) replay() error {
	walPath := filepath.Join(wb.storagePath, "pubsub-operations.wal")
	f, err := os.Open(walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var op pubsubWALOperation
		if err := json.Unmarshal(scanner.Bytes(), &op); err != nil {
			continue
		}
		if err := wb.applyOperation(&op); err != nil {
			fmt.Fprintf(os.Stderr, "PubSub WAL replay error: %v\n", err)
		}
	}
	return scanner.Err()
}

func (wb *WALPubSubBackend) applyOperation(op *pubsubWALOperation) error {
	switch op.Type {
	case "create_topic":
		var topic model.Topic
		if data, ok := op.Data["topic"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			json.Unmarshal(jsonData, &topic)
		}
		_, err := wb.MemoryPubSubBackend.CreateTopic(context.Background(), &topic)
		return err
	case "delete_topic":
		project, _ := op.Data["project"].(string)
		name, _ := op.Data["name"].(string)
		return wb.MemoryPubSubBackend.DeleteTopic(context.Background(), project, name)
	case "create_subscription":
		var sub model.Subscription
		if data, ok := op.Data["subscription"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			json.Unmarshal(jsonData, &sub)
		}
		_, err := wb.MemoryPubSubBackend.CreateSubscription(context.Background(), &sub)
		return err
	case "delete_subscription":
		project, _ := op.Data["project"].(string)
		name, _ := op.Data["name"].(string)
		return wb.MemoryPubSubBackend.DeleteSubscription(context.Background(), project, name)
	case "create_schema":
		var schema model.Schema
		if data, ok := op.Data["schema"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			json.Unmarshal(jsonData, &schema)
		}
		_, err := wb.MemoryPubSubBackend.CreateSchema(context.Background(), &schema)
		return err
	case "delete_schema":
		project, _ := op.Data["project"].(string)
		name, _ := op.Data["name"].(string)
		return wb.MemoryPubSubBackend.DeleteSchema(context.Background(), project, name)
	case "commit_schema":
		project, _ := op.Data["project"].(string)
		name, _ := op.Data["name"].(string)
		var schema model.Schema
		if data, ok := op.Data["schema"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			json.Unmarshal(jsonData, &schema)
		}
		_, err := wb.MemoryPubSubBackend.CommitSchema(context.Background(), project, name, &schema)
		return err
	}
	return nil
}

func (wb *WALPubSubBackend) compact() error {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	if err := wb.saveSnapshot(); err != nil {
		return err
	}
	wb.walFile.Close()

	walPath := filepath.Join(wb.storagePath, "pubsub-operations.wal")
	if err := os.Remove(walPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	wb.walFile = f
	wb.walWriter = bufio.NewWriter(f)
	wb.walSize = 0
	return nil
}

func (wb *WALPubSubBackend) saveSnapshot() error {
	wb.MemoryPubSubBackend.mu.RLock()
	defer wb.MemoryPubSubBackend.mu.RUnlock()

	data := pubsubPersistentData{
		Topics:        wb.topics,
		Subscriptions: wb.subscriptions,
		Schemas:       wb.schemas,
		SchemaRevs:    wb.schemaRevisions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(wb.storagePath, "pubsub-snapshot.json"), jsonData, 0644)
}

func (wb *WALPubSubBackend) loadSnapshot() error {
	snapshotPath := filepath.Join(wb.storagePath, "pubsub-snapshot.json")
	jsonData, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}

	var data pubsubPersistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	wb.topics = data.Topics
	if wb.topics == nil {
		wb.topics = make(map[string]*model.Topic)
	}
	wb.subscriptions = data.Subscriptions
	if wb.subscriptions == nil {
		wb.subscriptions = make(map[string]*model.Subscription)
	}
	wb.schemas = data.Schemas
	if wb.schemas == nil {
		wb.schemas = make(map[string]*model.Schema)
	}
	wb.schemaRevisions = data.SchemaRevs
	if wb.schemaRevisions == nil {
		wb.schemaRevisions = make(map[string][]*model.Schema)
	}
	return nil
}
