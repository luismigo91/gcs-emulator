package backend

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

type PersistentBackend struct {
	*MemoryBackend
	mu         sync.Mutex
	storagePath string
}

func NewPersistentBackend(storagePath string) (*PersistentBackend, error) {
	pb := &PersistentBackend{
		MemoryBackend: NewMemoryBackend(),
		storagePath:   storagePath,
	}

	pb.MemoryBackend.SetBlobDir(filepath.Join(storagePath, "blobs"))

	if err := pb.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return pb, nil
}

func (p *PersistentBackend) CreateBucket(ctx context.Context, bucket *model.Bucket, conds Conditions) error {
	if err := p.MemoryBackend.CreateBucket(ctx, bucket, conds); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentBackend) CreateObject(ctx context.Context, bucket string, object *model.Object, content io.Reader, conds Conditions) error {
	if err := p.MemoryBackend.CreateObject(ctx, bucket, object, content, conds); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentBackend) DeleteBucket(ctx context.Context, name string, conds Conditions) error {
	if err := p.MemoryBackend.DeleteBucket(ctx, name, conds); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentBackend) DeleteObject(ctx context.Context, bucket, name string, generation int64, conds Conditions) error {
	if err := p.MemoryBackend.DeleteObject(ctx, bucket, name, generation, conds); err != nil {
		return err
	}
	return p.save()
}

func (p *PersistentBackend) UpdateBucket(ctx context.Context, name string, attrs *model.BucketUpdateAttrs, conds Conditions) (*model.Bucket, error) {
	bucket, err := p.MemoryBackend.UpdateBucket(ctx, name, attrs, conds)
	if err != nil {
		return nil, err
	}
	if saveErr := p.save(); saveErr != nil {
		return bucket, saveErr
	}
	return bucket, nil
}

func (p *PersistentBackend) UpdateObject(ctx context.Context, bucket, name string, attrs *model.ObjectUpdateAttrs, conds Conditions) (*model.Object, error) {
	obj, err := p.MemoryBackend.UpdateObject(ctx, bucket, name, attrs, conds)
	if err != nil {
		return nil, err
	}
	if saveErr := p.save(); saveErr != nil {
		return obj, saveErr
	}
	return obj, nil
}

func (p *PersistentBackend) Shutdown() error {
	return p.save()
}

type persistentData struct {
	Buckets            map[string]*model.Bucket         `json:"buckets"`
	Objects            map[string]*model.Object         `json:"objects"`
	Content            map[string][]byte                `json:"-"`
	IamPolicies        map[string]*model.Policy         `json:"iamPolicies,omitempty"`
	Notifications      map[string][]*model.Notification `json:"notifications,omitempty"`
	BucketACLs         map[string][]*model.BucketACL    `json:"bucketACLs,omitempty"`
	ObjectACLs         map[string][]*model.ObjectACL    `json:"objectACLs,omitempty"`
	DefaultObjectACLs  map[string][]*model.ObjectACL    `json:"defaultObjectACLs,omitempty"`
}

func (p *PersistentBackend) save() error {
	p.mu.Lock()
	p.MemoryBackend.mu.RLock()
	defer p.mu.Unlock()
	defer p.MemoryBackend.mu.RUnlock()

	if err := os.MkdirAll(p.storagePath, 0755); err != nil {
		return err
	}

	data := persistentData{
		Buckets:           p.buckets,
		Objects:           p.objects,
		IamPolicies:       p.iamPolicies,
		Notifications:     p.notifications,
		BucketACLs:        p.bucketACLs,
		ObjectACLs:        p.objectACLs,
		DefaultObjectACLs: p.defaultObjectACLs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(p.storagePath, "data.json"), jsonData, 0644)
}

func (p *PersistentBackend) load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dataFile := filepath.Join(p.storagePath, "data.json")
	jsonData, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	var data persistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	p.buckets = data.Buckets
	if p.buckets == nil {
		p.buckets = make(map[string]*model.Bucket)
	}
	p.objects = data.Objects
	if p.objects == nil {
		p.objects = make(map[string]*model.Object)
	}
	p.content = make(map[string][]byte)

	p.iamPolicies = data.IamPolicies
	if p.iamPolicies == nil {
		p.iamPolicies = make(map[string]*model.Policy)
	}
	p.notifications = data.Notifications
	if p.notifications == nil {
		p.notifications = make(map[string][]*model.Notification)
	}

	p.bucketACLs = data.BucketACLs
	if p.bucketACLs == nil {
		p.bucketACLs = make(map[string][]*model.BucketACL)
	}
	p.objectACLs = data.ObjectACLs
	if p.objectACLs == nil {
		p.objectACLs = make(map[string][]*model.ObjectACL)
	}
	p.defaultObjectACLs = data.DefaultObjectACLs
	if p.defaultObjectACLs == nil {
		p.defaultObjectACLs = make(map[string][]*model.ObjectACL)
	}

	return nil
}
