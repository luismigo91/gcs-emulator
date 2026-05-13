package backend

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

type WALOperation struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type WALBackend struct {
	*MemoryBackend
	storagePath      string
	walFile          *os.File
	walWriter        *bufio.Writer
	mu               sync.Mutex
	walSize          int64
	compactThreshold int64
}

const defaultCompactThreshold = 10 * 1024 * 1024

func NewWALBackend(storagePath string) (*WALBackend, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, err
	}

	wb := &WALBackend{
		MemoryBackend:    NewMemoryBackend(),
		storagePath:      storagePath,
		compactThreshold: defaultCompactThreshold,
	}

	wb.MemoryBackend.SetBlobDir(filepath.Join(storagePath, "blobs"))

	walPath := filepath.Join(storagePath, "operations.wal")
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

func (wb *WALBackend) appendOperation(op *WALOperation) error {
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

func (wb *WALBackend) CreateBucket(ctx context.Context, bucket *model.Bucket, conds Conditions) error {
	if err := wb.MemoryBackend.CreateBucket(ctx, bucket, conds); err != nil {
		return err
	}

	op := &WALOperation{
		Type:      "create_bucket",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name":   bucket.Name,
			"bucket": bucket,
		},
	}
	return wb.appendOperation(op)
}

func (wb *WALBackend) CreateObject(ctx context.Context, bucket string, object *model.Object, content io.Reader, conds Conditions) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	if err := wb.MemoryBackend.CreateObject(ctx, bucket, object, bytes.NewReader(data), conds); err != nil {
		return err
	}

	op := &WALOperation{
		Type:      "create_object",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket":     bucket,
			"object":     object,
			"contentLen": len(data),
		},
	}
	return wb.appendOperation(op)
}

func (wb *WALBackend) DeleteBucket(ctx context.Context, name string, conds Conditions) error {
	if err := wb.MemoryBackend.DeleteBucket(ctx, name, conds); err != nil {
		return err
	}

	op := &WALOperation{
		Type:      "delete_bucket",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name": name,
		},
	}
	return wb.appendOperation(op)
}

func (wb *WALBackend) DeleteObject(ctx context.Context, bucket, name string, generation int64, conds Conditions) error {
	if err := wb.MemoryBackend.DeleteObject(ctx, bucket, name, generation, conds); err != nil {
		return err
	}

	op := &WALOperation{
		Type:      "delete_object",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket":     bucket,
			"name":       name,
			"generation": generation,
		},
	}
	return wb.appendOperation(op)
}

func (wb *WALBackend) UpdateBucket(ctx context.Context, name string, attrs *model.BucketUpdateAttrs, conds Conditions) (*model.Bucket, error) {
	bucket, err := wb.MemoryBackend.UpdateBucket(ctx, name, attrs, conds)
	if err != nil {
		return nil, err
	}

	op := &WALOperation{
		Type:      "update_bucket",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name":  name,
			"attrs": attrs,
		},
	}
	if appendErr := wb.appendOperation(op); appendErr != nil {
		return bucket, appendErr
	}
	return bucket, nil
}

func (wb *WALBackend) UpdateObject(ctx context.Context, bucket, name string, attrs *model.ObjectUpdateAttrs, conds Conditions) (*model.Object, error) {
	obj, err := wb.MemoryBackend.UpdateObject(ctx, bucket, name, attrs, conds)
	if err != nil {
		return nil, err
	}

	op := &WALOperation{
		Type:      "update_object",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket": bucket,
			"name":   name,
			"attrs":  attrs,
		},
	}
	if appendErr := wb.appendOperation(op); appendErr != nil {
		return obj, appendErr
	}
	return obj, nil
}

func (wb *WALBackend) SetBucketIAMPolicy(ctx context.Context, bucket string, policy *model.Policy) (*model.Policy, error) {
	result, err := wb.MemoryBackend.SetBucketIAMPolicy(ctx, bucket, policy)
	if err != nil {
		return nil, err
	}

	op := &WALOperation{
		Type:      "set_iam_policy",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket": bucket,
			"policy": policy,
		},
	}
	if appendErr := wb.appendOperation(op); appendErr != nil {
		return result, appendErr
	}
	return result, nil
}

func (wb *WALBackend) CreateNotification(ctx context.Context, bucket string, notification *model.Notification) (*model.Notification, error) {
	result, err := wb.MemoryBackend.CreateNotification(ctx, bucket, notification)
	if err != nil {
		return nil, err
	}

	op := &WALOperation{
		Type:      "create_notification",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket":       bucket,
			"notification": notification,
		},
	}
	if appendErr := wb.appendOperation(op); appendErr != nil {
		return result, appendErr
	}
	return result, nil
}

func (wb *WALBackend) DeleteNotification(ctx context.Context, bucket, id string) error {
	err := wb.MemoryBackend.DeleteNotification(ctx, bucket, id)
	if err != nil {
		return err
	}

	op := &WALOperation{
		Type:      "delete_notification",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"bucket": bucket,
			"id":     id,
		},
	}
	return wb.appendOperation(op)
}

func (wb *WALBackend) Shutdown() error {
	if err := wb.compact(); err != nil {
		return err
	}
	return wb.walFile.Close()
}

func (wb *WALBackend) replay() error {
	walPath := filepath.Join(wb.storagePath, "operations.wal")
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
		var op WALOperation
		if err := json.Unmarshal(scanner.Bytes(), &op); err != nil {
			continue
		}

		if err := wb.applyOperation(&op); err != nil {
			fmt.Fprintf(os.Stderr, "WAL replay error: %v\n", err)
		}
	}

	return scanner.Err()
}

func (wb *WALBackend) applyOperation(op *WALOperation) error {
	switch op.Type {
	case "create_bucket":
		var bucket model.Bucket
		if data, ok := op.Data["bucket"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			if err := json.Unmarshal(jsonData, &bucket); err != nil {
				return err
			}
		}
		return wb.MemoryBackend.CreateBucket(context.Background(), &bucket, Conditions{})

	case "delete_bucket":
		if name, ok := op.Data["name"].(string); ok {
			return wb.MemoryBackend.DeleteBucket(context.Background(), name, Conditions{})
		}

	case "create_object":
		var object model.Object
		if data, ok := op.Data["object"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			if err := json.Unmarshal(jsonData, &object); err != nil {
				return err
			}
		}
		if bucket, ok := op.Data["bucket"].(string); ok {
			return wb.MemoryBackend.CreateObject(context.Background(), bucket, &object, nil, Conditions{})
		}

	case "delete_object":
		if bucket, ok := op.Data["bucket"].(string); ok {
			if name, ok := op.Data["name"].(string); ok {
				generation, _ := op.Data["generation"].(float64)
				return wb.MemoryBackend.DeleteObject(context.Background(), bucket, name, int64(generation), Conditions{})
			}
		}

	case "set_iam_policy":
		var policy model.Policy
		if data, ok := op.Data["policy"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			if err := json.Unmarshal(jsonData, &policy); err != nil {
				return err
			}
		}
		if bucket, ok := op.Data["bucket"].(string); ok {
			_, err := wb.MemoryBackend.SetBucketIAMPolicy(context.Background(), bucket, &policy)
			return err
		}

	case "create_notification":
		var notification model.Notification
		if data, ok := op.Data["notification"].(map[string]interface{}); ok {
			jsonData, _ := json.Marshal(data)
			if err := json.Unmarshal(jsonData, &notification); err != nil {
				return err
			}
		}
		if bucket, ok := op.Data["bucket"].(string); ok {
			_, err := wb.MemoryBackend.CreateNotification(context.Background(), bucket, &notification)
			return err
		}

	case "delete_notification":
		if bucket, ok := op.Data["bucket"].(string); ok {
			if id, ok := op.Data["id"].(string); ok {
				return wb.MemoryBackend.DeleteNotification(context.Background(), bucket, id)
			}
		}
	}
	return nil
}

func (wb *WALBackend) compact() error {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	if err := wb.saveSnapshot(); err != nil {
		return err
	}

	wb.walFile.Close()

	walPath := filepath.Join(wb.storagePath, "operations.wal")
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

func (wb *WALBackend) saveSnapshot() error {
	wb.MemoryBackend.mu.RLock()
	defer wb.MemoryBackend.mu.RUnlock()

	data := persistentData{
		Buckets:           wb.buckets,
		Objects:           wb.objects,
		IamPolicies:       wb.iamPolicies,
		Notifications:     wb.notifications,
		BucketACLs:        wb.bucketACLs,
		ObjectACLs:        wb.objectACLs,
		DefaultObjectACLs: wb.defaultObjectACLs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(wb.storagePath, "snapshot.json"), jsonData, 0644)
}

func (wb *WALBackend) loadSnapshot() error {
	snapshotPath := filepath.Join(wb.storagePath, "snapshot.json")
	jsonData, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}

	var data persistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	wb.buckets = data.Buckets
	if wb.buckets == nil {
		wb.buckets = make(map[string]*model.Bucket)
	}
	wb.objects = data.Objects
	if wb.objects == nil {
		wb.objects = make(map[string]*model.Object)
	}
	wb.content = make(map[string][]byte)

	wb.iamPolicies = data.IamPolicies
	if wb.iamPolicies == nil {
		wb.iamPolicies = make(map[string]*model.Policy)
	}
	wb.notifications = data.Notifications
	if wb.notifications == nil {
		wb.notifications = make(map[string][]*model.Notification)
	}

	wb.bucketACLs = data.BucketACLs
	if wb.bucketACLs == nil {
		wb.bucketACLs = make(map[string][]*model.BucketACL)
	}
	wb.objectACLs = data.ObjectACLs
	if wb.objectACLs == nil {
		wb.objectACLs = make(map[string][]*model.ObjectACL)
	}
	wb.defaultObjectACLs = data.DefaultObjectACLs
	if wb.defaultObjectACLs == nil {
		wb.defaultObjectACLs = make(map[string][]*model.ObjectACL)
	}

	return nil
}
