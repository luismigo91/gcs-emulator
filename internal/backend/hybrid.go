package backend

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

type HybridBackend struct {
	*MemoryBackend
	storagePath   string
	flushInterval time.Duration
	stopCh        chan struct{}
	done          chan struct{}
	mu            sync.Mutex
	dirty         bool
}

func NewHybridBackend(storagePath string, flushInterval time.Duration) (*HybridBackend, error) {
	hb := &HybridBackend{
		MemoryBackend: NewMemoryBackend(),
		storagePath:   storagePath,
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
		done:          make(chan struct{}),
	}

	hb.MemoryBackend.SetBlobDir(filepath.Join(storagePath, "blobs"))

	if err := hb.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	go hb.flushLoop()

	return hb, nil
}

func (hb *HybridBackend) markDirty() {
	hb.mu.Lock()
	hb.dirty = true
	hb.mu.Unlock()
}

func (hb *HybridBackend) flushLoop() {
	ticker := time.NewTicker(hb.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hb.mu.Lock()
			if hb.dirty {
				hb.dirty = false
				hb.mu.Unlock()
				_ = hb.save()
			} else {
				hb.mu.Unlock()
			}
		case <-hb.stopCh:
			hb.mu.Lock()
			d := hb.dirty
			hb.dirty = false
			hb.mu.Unlock()
			if d {
				_ = hb.save()
			}
			close(hb.done)
			return
		}
	}
}

func (hb *HybridBackend) CreateBucket(ctx context.Context, bucket *model.Bucket, conds Conditions) error {
	if err := hb.MemoryBackend.CreateBucket(ctx, bucket, conds); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridBackend) CreateObject(ctx context.Context, bucket string, object *model.Object, content io.Reader, conds Conditions) error {
	if err := hb.MemoryBackend.CreateObject(ctx, bucket, object, content, conds); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridBackend) DeleteBucket(ctx context.Context, name string, conds Conditions) error {
	if err := hb.MemoryBackend.DeleteBucket(ctx, name, conds); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridBackend) DeleteObject(ctx context.Context, bucket, name string, generation int64, conds Conditions) error {
	if err := hb.MemoryBackend.DeleteObject(ctx, bucket, name, generation, conds); err != nil {
		return err
	}
	hb.markDirty()
	return nil
}

func (hb *HybridBackend) UpdateBucket(ctx context.Context, name string, attrs *model.BucketUpdateAttrs, conds Conditions) (*model.Bucket, error) {
	bucket, err := hb.MemoryBackend.UpdateBucket(ctx, name, attrs, conds)
	if err == nil {
		hb.markDirty()
	}
	return bucket, err
}

func (hb *HybridBackend) UpdateObject(ctx context.Context, bucket, name string, attrs *model.ObjectUpdateAttrs, conds Conditions) (*model.Object, error) {
	obj, err := hb.MemoryBackend.UpdateObject(ctx, bucket, name, attrs, conds)
	if err == nil {
		hb.markDirty()
	}
	return obj, err
}

func (hb *HybridBackend) SetBucketIAMPolicy(ctx context.Context, bucket string, policy *model.Policy) (*model.Policy, error) {
	result, err := hb.MemoryBackend.SetBucketIAMPolicy(ctx, bucket, policy)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridBackend) CreateNotification(ctx context.Context, bucket string, notification *model.Notification) (*model.Notification, error) {
	result, err := hb.MemoryBackend.CreateNotification(ctx, bucket, notification)
	if err == nil {
		hb.markDirty()
	}
	return result, err
}

func (hb *HybridBackend) DeleteNotification(ctx context.Context, bucket, id string) error {
	err := hb.MemoryBackend.DeleteNotification(ctx, bucket, id)
	if err == nil {
		hb.markDirty()
	}
	return err
}

func (hb *HybridBackend) Shutdown() error {
	close(hb.stopCh)
	<-hb.done
	return nil
}

func (hb *HybridBackend) save() error {
	hb.MemoryBackend.mu.RLock()
	defer hb.MemoryBackend.mu.RUnlock()

	if err := os.MkdirAll(hb.storagePath, 0755); err != nil {
		return err
	}

	data := persistentData{
		Buckets:           hb.buckets,
		Objects:           hb.objects,
		IamPolicies:       hb.iamPolicies,
		Notifications:     hb.notifications,
		BucketACLs:        hb.bucketACLs,
		ObjectACLs:        hb.objectACLs,
		DefaultObjectACLs: hb.defaultObjectACLs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(hb.storagePath, "data.json"), jsonData, 0644)
}

func (hb *HybridBackend) load() error {
	jsonData, err := os.ReadFile(filepath.Join(hb.storagePath, "data.json"))
	if err != nil {
		return err
	}

	var data persistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	hb.buckets = data.Buckets
	if hb.buckets == nil {
		hb.buckets = make(map[string]*model.Bucket)
	}
	hb.objects = data.Objects
	if hb.objects == nil {
		hb.objects = make(map[string]*model.Object)
	}
	hb.content = make(map[string][]byte)

	hb.iamPolicies = data.IamPolicies
	if hb.iamPolicies == nil {
		hb.iamPolicies = make(map[string]*model.Policy)
	}
	hb.notifications = data.Notifications
	if hb.notifications == nil {
		hb.notifications = make(map[string][]*model.Notification)
	}

	hb.bucketACLs = data.BucketACLs
	if hb.bucketACLs == nil {
		hb.bucketACLs = make(map[string][]*model.BucketACL)
	}
	hb.objectACLs = data.ObjectACLs
	if hb.objectACLs == nil {
		hb.objectACLs = make(map[string][]*model.ObjectACL)
	}
	hb.defaultObjectACLs = data.DefaultObjectACLs
	if hb.defaultObjectACLs == nil {
		hb.defaultObjectACLs = make(map[string][]*model.ObjectACL)
	}

	return nil
}
