package backend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/util"
)

type MemoryBackend struct {
	mu      sync.RWMutex
	buckets map[string]*model.Bucket
	objects map[string]*model.Object
	content map[string][]byte
	iamPolicies map[string]*model.Policy
	notifications map[string][]*model.Notification
	bucketACLs map[string][]*model.BucketACL
	objectACLs map[string][]*model.ObjectACL
	defaultObjectACLs map[string][]*model.ObjectACL
	genCounter int64
	blobDir string
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		buckets: make(map[string]*model.Bucket),
		objects: make(map[string]*model.Object),
		content: make(map[string][]byte),
		iamPolicies: make(map[string]*model.Policy),
		notifications: make(map[string][]*model.Notification),
		bucketACLs: make(map[string][]*model.BucketACL),
		objectACLs: make(map[string][]*model.ObjectACL),
		defaultObjectACLs: make(map[string][]*model.ObjectACL),
	}
}

func (m *MemoryBackend) SetBlobDir(dir string) {
	m.blobDir = dir
}

func (m *MemoryBackend) writeBlob(key string, data []byte) error {
	if m.blobDir == "" {
		return nil
	}
	blobPath := filepath.Join(m.blobDir, key)
	if err := os.MkdirAll(filepath.Dir(blobPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(blobPath, data, 0644)
}

func (m *MemoryBackend) deleteBlob(key string) error {
	if m.blobDir == "" {
		return nil
	}
	blobPath := filepath.Join(m.blobDir, key)
	return os.Remove(blobPath)
}

func (m *MemoryBackend) readBlob(key string) ([]byte, error) {
	blobPath := filepath.Join(m.blobDir, key)
	return os.ReadFile(blobPath)
}

func (m *MemoryBackend) CreateBucket(ctx context.Context, bucket *model.Bucket, conds Conditions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket.Name]; exists {
		return ErrBucketAlreadyExists
	}
	if bucket.Metageneration == 0 {
		bucket.Metageneration = 1
	}
	m.buckets[bucket.Name] = bucket
	return nil
}

func (m *MemoryBackend) GetBucket(ctx context.Context, name string) (*model.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, exists := m.buckets[name]
	if !exists {
		return nil, ErrBucketNotFound
	}
	return bucket, nil
}

func (m *MemoryBackend) ListBuckets(ctx context.Context, params ListBucketsParams) ([]*model.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*model.Bucket
	for _, bucket := range m.buckets {
		if params.Project != "" && bucket.ProjectID != params.Project {
			continue
		}
		result = append(result, bucket)
	}
	return result, nil
}

func (m *MemoryBackend) DeleteBucket(ctx context.Context, name string, conds Conditions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[name]; !exists {
		return ErrBucketNotFound
	}

	for key, obj := range m.objects {
		_ = key
		if obj.Bucket == name {
			return ErrBucketNotEmpty
		}
	}

	delete(m.buckets, name)
	return nil
}

func (m *MemoryBackend) UpdateBucket(ctx context.Context, name string, attrs *model.BucketUpdateAttrs, conds Conditions) (*model.Bucket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, exists := m.buckets[name]
	if !exists {
		return nil, ErrBucketNotFound
	}

	if !conds.IsZero() {
		if err := m.checkBucketConditions(bucket, conds); err != nil {
			return nil, err
		}
	}

	if attrs.Versioning != nil {
		bucket.Versioning = attrs.Versioning
	}
	if attrs.Lifecycle != nil {
		bucket.Lifecycle = attrs.Lifecycle
	}
	if attrs.IamConfiguration != nil {
		bucket.IamConfiguration = attrs.IamConfiguration
	}
	if attrs.Labels != nil {
		bucket.Labels = attrs.Labels
	}
	if attrs.RetentionPolicy != nil {
		bucket.RetentionPolicy = attrs.RetentionPolicy
	}
	if attrs.CORS != nil {
		bucket.CORS = attrs.CORS
	}

	bucket.Metageneration++
	return bucket, nil
}

func (m *MemoryBackend) checkBucketConditions(bucket *model.Bucket, conds Conditions) error {
	if conds.HasMetagenerationMatch && bucket.Metageneration != conds.MetagenerationMatch {
		return ErrPreconditionFailed
	}
	if conds.HasMetagenerationNotMatch && bucket.Metageneration == conds.MetagenerationNotMatch {
		return ErrPreconditionFailed
	}
	return nil
}

func (m *MemoryBackend) CreateObject(ctx context.Context, bucket string, object *model.Object, content io.Reader, conds Conditions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return ErrBucketNotFound
	}

	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	now := time.Now()
	if object.TimeCreated.IsZero() {
		object.TimeCreated = now
	}
	object.Updated = now
	object.Size = int64(len(data))
	object.CRC32C = util.CalculateCRC32C(data)
	object.MD5Hash = util.CalculateMD5(data)

	existingObj, existsExisting := m.findLatestObjectLocked(bucket, object.Name)

	if conds.DoesNotExist && existsExisting {
		return ErrPreconditionFailed
	}
	if conds.HasGenerationMatch && (!existsExisting || existingObj.Generation != conds.GenerationMatch) {
		return ErrPreconditionFailed
	}
	if conds.HasGenerationNotMatch && existsExisting && existingObj.Generation == conds.GenerationNotMatch {
		return ErrPreconditionFailed
	}

	bucketObj := m.buckets[bucket]
	versioningEnabled := bucketObj != nil && bucketObj.Versioning != nil && bucketObj.Versioning.Enabled

	if existsExisting && !versioningEnabled {
		existingKey := objectKey(bucket, object.Name, existingObj.Generation)
		delete(m.objects, existingKey)
		delete(m.content, existingKey)
		m.deleteBlob(existingKey)
	}

	m.genCounter++
	ts := now.UnixNano()
	if existsExisting && existingObj != nil && ts <= existingObj.Generation {
		ts = existingObj.Generation + m.genCounter
	}
	object.Generation = ts
	object.Metageneration = 1

	key := objectKey(bucket, object.Name, object.Generation)
	m.objects[key] = object
	m.content[key] = data

	if err := m.writeBlob(key, data); err != nil {
		return err
	}

	return nil
}

func (m *MemoryBackend) GetObject(ctx context.Context, bucket, name string, generation int64) (*model.Object, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if generation > 0 {
		key := objectKey(bucket, name, generation)
		obj, exists := m.objects[key]
		if !exists {
			return nil, ErrGenerationNotFound
		}
		return obj, nil
	}

	var latest *model.Object
	var latestGen int64
	for key, obj := range m.objects {
		if obj.Bucket == bucket && obj.Name == name {
			objKey := parseObjectKey(key)
			if objKey.generation > latestGen {
				latestGen = objKey.generation
				latest = obj
			}
		}
	}

	if latest == nil {
		return nil, ErrObjectNotFound
	}
	return latest, nil
}

func (m *MemoryBackend) GetObjectContent(ctx context.Context, bucket, name string, generation int64) (io.ReadSeeker, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var key string
	if generation > 0 {
		key = objectKey(bucket, name, generation)
	} else {
		var latestGen int64
		for k, obj := range m.objects {
			if obj.Bucket == bucket && obj.Name == name {
				objKey := parseObjectKey(k)
				if objKey.generation > latestGen {
					latestGen = objKey.generation
					key = k
				}
			}
		}
	}

	if key == "" {
		return nil, ErrObjectNotFound
	}

	data, exists := m.content[key]
	if !exists || len(data) == 0 {
		if m.blobDir != "" {
			blobData, err := m.readBlob(key)
			if err != nil {
				return nil, ErrObjectNotFound
			}
			return bytes.NewReader(blobData), nil
		}
		return nil, ErrObjectNotFound
	}
	return bytes.NewReader(data), nil
}

func (m *MemoryBackend) ListObjects(ctx context.Context, bucket string, params ListObjectsParams) (*ListObjectsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	type objectEntry struct {
		obj   *model.Object
		key   string
	}

	var allObjects []objectEntry
	for key, obj := range m.objects {
		if obj.Bucket != bucket {
			continue
		}
		if !params.Versions {
			objKey := parseObjectKey(key)
			if !isLiveVersion(obj, objKey.generation, m.objects, bucket, obj.Name) {
				continue
			}
		}
		if params.Prefix != "" && !strings.HasPrefix(obj.Name, params.Prefix) {
			continue
		}
		if params.StartOffset != "" && obj.Name < params.StartOffset {
			continue
		}
		if params.EndOffset != "" && obj.Name >= params.EndOffset {
			continue
		}
		allObjects = append(allObjects, objectEntry{obj, key})
	}

	sort.Slice(allObjects, func(i, j int) bool {
		return allObjects[i].obj.Name < allObjects[j].obj.Name
	})

	prefixSet := make(map[string]bool)
	var items []*model.Object

	for _, entry := range allObjects {
		obj := entry.obj
		nameWithoutPrefix := strings.TrimPrefix(obj.Name, params.Prefix)

		if params.Delimiter != "" {
			delimPos := strings.Index(nameWithoutPrefix, params.Delimiter)
			if delimPos >= 0 {
				prefix := params.Prefix + nameWithoutPrefix[:delimPos+len(params.Delimiter)]
				prefixSet[prefix] = true
				if params.IncludeTrailingDelimiter && obj.Name == prefix {
					items = append(items, obj)
				}
				continue
			}
		}
		items = append(items, obj)
	}

	var prefixes []string
	for p := range prefixSet {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	nextPageToken := ""
	if params.MaxResults > 0 && len(items) > params.MaxResults {
		items = items[:params.MaxResults]
		nextPageToken = items[len(items)-1].Name
	}

	return &ListObjectsResponse{
		Items:         items,
		Prefixes:      prefixes,
		NextPageToken: nextPageToken,
	}, nil
}

func (m *MemoryBackend) DeleteObject(ctx context.Context, bucket, name string, generation int64, conds Conditions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if generation > 0 {
		key := objectKey(bucket, name, generation)
		if _, exists := m.objects[key]; !exists {
			return ErrObjectNotFound
		}
		delete(m.objects, key)
		delete(m.content, key)
		m.deleteBlob(key)
		return nil
	}

	var latestKey string
	var latestGen int64
	for key, obj := range m.objects {
		if obj.Bucket == bucket && obj.Name == name {
			objKey := parseObjectKey(key)
			if objKey.generation > latestGen {
				latestGen = objKey.generation
				latestKey = key
			}
		}
	}

	if latestKey == "" {
		return ErrObjectNotFound
	}

	bucketObj := m.buckets[bucket]
	versioningEnabled := bucketObj != nil && bucketObj.Versioning != nil && bucketObj.Versioning.Enabled

	delete(m.objects, latestKey)
	delete(m.content, latestKey)
	m.deleteBlob(latestKey)

	if versioningEnabled {
		tombstone := &model.Object{
			Name:       name,
			Bucket:     bucket,
			Generation: latestGen,
			TimeDeleted: time.Now(),
		}
		tombstoneKey := objectKey(bucket, name, latestGen)
		m.objects[tombstoneKey] = tombstone
	}

	return nil
}

func (m *MemoryBackend) UpdateObject(ctx context.Context, bucket, name string, attrs *model.ObjectUpdateAttrs, conds Conditions) (*model.Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	obj, found := m.findLatestObjectLocked(bucket, name)
	if !found {
		return nil, ErrObjectNotFound
	}

	if !conds.IsZero() {
		if err := m.checkObjectConditions(obj, conds); err != nil {
			return nil, err
		}
	}

	if attrs.ContentType != nil {
		obj.ContentType = *attrs.ContentType
	}
	if attrs.ContentEncoding != nil {
		obj.ContentEncoding = *attrs.ContentEncoding
	}
	if attrs.ContentDisposition != nil {
		obj.ContentDisposition = *attrs.ContentDisposition
	}
	if attrs.ContentLanguage != nil {
		obj.ContentLanguage = *attrs.ContentLanguage
	}
	if attrs.CacheControl != nil {
		obj.CacheControl = *attrs.CacheControl
	}
	if attrs.Metadata != nil {
		obj.Metadata = attrs.Metadata
	}
	if attrs.EventBasedHold != nil {
		obj.EventBasedHold = *attrs.EventBasedHold
	}
	if attrs.TemporaryHold != nil {
		obj.TemporaryHold = *attrs.TemporaryHold
	}

	obj.Metageneration++
	return obj, nil
}

func (m *MemoryBackend) ComposeObjects(ctx context.Context, bucket, destName string, sourceNames []string, destAttrs *model.Object) (*model.Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(sourceNames) > 32 {
		return nil, ErrComposeTooManySources
	}

	var composedData []byte
	componentCount := 0

	for _, srcName := range sourceNames {
		obj, found := m.findLatestObjectLocked(bucket, srcName)
		if !found {
			return nil, ErrObjectNotFound
		}
		key := objectKey(bucket, srcName, obj.Generation)
		data, exists := m.content[key]
		if !exists {
			return nil, ErrObjectNotFound
		}
		composedData = append(composedData, data...)
		componentCount++
	}

	now := time.Now()
	destObj := &model.Object{
		Name:          destName,
		Bucket:        bucket,
		Kind:          "storage#object",
		TimeCreated:   now,
		Updated:       now,
		Generation:    now.UnixNano(),
		Metageneration: 1,
		ComponentCount: componentCount,
		Size:          int64(len(composedData)),
	}

	if destAttrs != nil {
		if destAttrs.ContentType != "" {
			destObj.ContentType = destAttrs.ContentType
		}
		if destAttrs.Metadata != nil {
			destObj.Metadata = destAttrs.Metadata
		}
	}

	key := objectKey(bucket, destName, destObj.Generation)
	m.objects[key] = destObj
	m.content[key] = composedData

	return destObj, nil
}

func (m *MemoryBackend) CopyObject(ctx context.Context, srcBucket, srcName, destBucket, destName string, conds Conditions) (*model.Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	srcObj, found := m.findLatestObjectLocked(srcBucket, srcName)
	if !found {
		return nil, ErrObjectNotFound
	}

	srcKey := objectKey(srcBucket, srcName, srcObj.Generation)
	srcData := m.content[srcKey]

	now := time.Now()
	destObj := &model.Object{
		Name:          destName,
		Bucket:        destBucket,
		Kind:          "storage#object",
		ContentType:   srcObj.ContentType,
		ContentEncoding: srcObj.ContentEncoding,
		Metadata:      srcObj.Metadata,
		TimeCreated:   now,
		Updated:       now,
		Generation:    now.UnixNano(),
		Metageneration: 1,
		Size:          int64(len(srcData)),
	}

	destKey := objectKey(destBucket, destName, destObj.Generation)
	m.objects[destKey] = destObj
	m.content[destKey] = srcData

	return destObj, nil
}

func (m *MemoryBackend) RewriteObject(ctx context.Context, srcBucket, srcName, destBucket, destName string, token string, conds Conditions) (*model.Object, string, error) {
	obj, err := m.CopyObject(ctx, srcBucket, srcName, destBucket, destName, conds)
	if err != nil {
		return nil, "", err
	}
	return obj, "", nil
}

func (m *MemoryBackend) Shutdown() error {
	return nil
}

func (m *MemoryBackend) GetBucketIAMPolicy(ctx context.Context, bucket string) (*model.Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	policy, exists := m.iamPolicies[bucket]
	if !exists {
		policy = &model.Policy{
			Kind:     "storage#policy",
			Version:  1,
			Etag:     "CAE=",
			Bindings: []model.PolicyBinding{},
		}
	}
	return policy, nil
}

func (m *MemoryBackend) SetBucketIAMPolicy(ctx context.Context, bucket string, policy *model.Policy) (*model.Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	policy.Kind = "storage#policy"
	policy.ResourceID = "projects/_/buckets/" + bucket
	policy.Version = 1
	policy.Etag = "CAI="

	m.iamPolicies[bucket] = policy
	return policy, nil
}

func (m *MemoryBackend) TestBucketIAMPermissions(ctx context.Context, bucket string, permissions []string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	return permissions, nil
}

func (m *MemoryBackend) CreateNotification(ctx context.Context, bucket string, notification *model.Notification) (*model.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	if notification.ID == "" {
		notification.ID = fmt.Sprintf("%d", m.genCounter)
		m.genCounter++
	}
	notification.Kind = "storage#notification"

	m.notifications[bucket] = append(m.notifications[bucket], notification)
	return notification, nil
}

func (m *MemoryBackend) ListNotifications(ctx context.Context, bucket string) ([]*model.Notification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	return m.notifications[bucket], nil
}

func (m *MemoryBackend) GetNotification(ctx context.Context, bucket, id string) (*model.Notification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}

	for _, n := range m.notifications[bucket] {
		if n.ID == id {
			return n, nil
		}
	}
	return nil, ErrNotificationNotFound
}

func (m *MemoryBackend) DeleteNotification(ctx context.Context, bucket, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return ErrBucketNotFound
	}

	notifications := m.notifications[bucket]
	for i, n := range notifications {
		if n.ID == id {
			m.notifications[bucket] = append(notifications[:i], notifications[i+1:]...)
			return nil
		}
	}
	return ErrNotificationNotFound
}

func (m *MemoryBackend) GetLifecycle(ctx context.Context, bucket string) (*model.Lifecycle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	if b.Lifecycle == nil {
		return &model.Lifecycle{Rule: []model.LifecycleRule{}}, nil
	}
	return b.Lifecycle, nil
}

func (m *MemoryBackend) SetLifecycle(ctx context.Context, bucket string, lifecycle *model.Lifecycle) (*model.Lifecycle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	b.Lifecycle = lifecycle
	b.Metageneration++
	return lifecycle, nil
}

func (m *MemoryBackend) DeleteLifecycle(ctx context.Context, bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.buckets[bucket]
	if !exists {
		return ErrBucketNotFound
	}
	b.Lifecycle = nil
	b.Metageneration++
	return nil
}

func (m *MemoryBackend) GetCORS(ctx context.Context, bucket string) ([]model.CORSRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	if b.CORS == nil {
		return []model.CORSRule{}, nil
	}
	return b.CORS, nil
}

func (m *MemoryBackend) SetCORS(ctx context.Context, bucket string, rules []model.CORSRule) ([]model.CORSRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	b.CORS = rules
	b.Metageneration++
	return rules, nil
}

func (m *MemoryBackend) ListBucketACL(ctx context.Context, bucket string) ([]*model.BucketACL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	return m.bucketACLs[bucket], nil
}

func (m *MemoryBackend) GetBucketACL(ctx context.Context, bucket, entity string) (*model.BucketACL, error) {
	acls, err := m.ListBucketACL(ctx, bucket)
	if err != nil {
		return nil, err
	}
	for _, acl := range acls {
		if acl.Entity == entity {
			return acl, nil
		}
	}
	return nil, ErrObjectNotFound
}

func (m *MemoryBackend) CreateBucketACL(ctx context.Context, bucket string, acl *model.BucketACL) (*model.BucketACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	acl.Kind = "storage#bucketAccessControl"
	acl.Bucket = bucket
	m.bucketACLs[bucket] = append(m.bucketACLs[bucket], acl)
	return acl, nil
}

func (m *MemoryBackend) UpdateBucketACL(ctx context.Context, bucket, entity string, acl *model.BucketACL) (*model.BucketACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	acls := m.bucketACLs[bucket]
	for i, a := range acls {
		if a.Entity == entity {
			acl.Kind = "storage#bucketAccessControl"
			acl.Bucket = bucket
			acl.Entity = entity
			acls[i] = acl
			return acl, nil
		}
	}
	return nil, ErrObjectNotFound
}

func (m *MemoryBackend) DeleteBucketACL(ctx context.Context, bucket, entity string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		return ErrBucketNotFound
	}
	acls := m.bucketACLs[bucket]
	for i, a := range acls {
		if a.Entity == entity {
			m.bucketACLs[bucket] = append(acls[:i], acls[i+1:]...)
			return nil
		}
	}
	return ErrObjectNotFound
}

func objectACLKey(bucket, object, entity string, gen int64) string {
	return bucket + "/" + object + "/" + entity
}

func (m *MemoryBackend) ListObjectACL(ctx context.Context, bucket, object string, generation int64) ([]*model.ObjectACL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := objectACLKey(bucket, object, "", generation)
	return m.objectACLs[key], nil
}

func (m *MemoryBackend) GetObjectACL(ctx context.Context, bucket, object, entity string, generation int64) (*model.ObjectACL, error) {
	acls, err := m.ListObjectACL(ctx, bucket, object, generation)
	if err != nil {
		return nil, err
	}
	for _, acl := range acls {
		if acl.Entity == entity {
			return acl, nil
		}
	}
	return nil, ErrObjectNotFound
}

func (m *MemoryBackend) CreateObjectACL(ctx context.Context, bucket, object string, acl *model.ObjectACL) (*model.ObjectACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acl.Kind = "storage#objectAccessControl"
	acl.Bucket = bucket
	acl.Object = object
	key := objectACLKey(bucket, object, "", acl.Generation)
	m.objectACLs[key] = append(m.objectACLs[key], acl)
	return acl, nil
}

func (m *MemoryBackend) UpdateObjectACL(ctx context.Context, bucket, object, entity string, acl *model.ObjectACL) (*model.ObjectACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := objectACLKey(bucket, object, "", acl.Generation)
	acls := m.objectACLs[key]
	for i, a := range acls {
		if a.Entity == entity {
			acl.Kind = "storage#objectAccessControl"
			acl.Bucket = bucket
			acl.Object = object
			acl.Entity = entity
			acls[i] = acl
			return acl, nil
		}
	}
	return nil, ErrObjectNotFound
}

func (m *MemoryBackend) DeleteObjectACL(ctx context.Context, bucket, object, entity string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := objectACLKey(bucket, object, "", 0)
	acls := m.objectACLs[key]
	for i, a := range acls {
		if a.Entity == entity {
			m.objectACLs[key] = append(acls[:i], acls[i+1:]...)
			return nil
		}
	}
	return ErrObjectNotFound
}

func (m *MemoryBackend) GetWebsiteConfig(ctx context.Context, bucket string) (*model.WebsiteConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	if b.Website == nil {
		return &model.WebsiteConfig{}, nil
	}
	return b.Website, nil
}

func (m *MemoryBackend) SetWebsiteConfig(ctx context.Context, bucket string, config *model.WebsiteConfig) (*model.WebsiteConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	b.Website = config
	b.Metageneration++
	return config, nil
}

func (m *MemoryBackend) DeleteWebsiteConfig(ctx context.Context, bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, exists := m.buckets[bucket]
	if !exists {
		return ErrBucketNotFound
	}
	b.Website = nil
	b.Metageneration++
	return nil
}

func (m *MemoryBackend) GetEncryptionConfig(ctx context.Context, bucket string) (*model.EncryptionConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	if b.Encryption == nil {
		return &model.EncryptionConfig{}, nil
	}
	return b.Encryption, nil
}

func (m *MemoryBackend) SetEncryptionConfig(ctx context.Context, bucket string, config *model.EncryptionConfig) (*model.EncryptionConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, exists := m.buckets[bucket]
	if !exists {
		return nil, ErrBucketNotFound
	}
	b.Encryption = config
	b.Metageneration++
	return config, nil
}

func (m *MemoryBackend) ListDefaultObjectACL(ctx context.Context, bucket string) ([]*model.ObjectACL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	return m.defaultObjectACLs[bucket], nil
}

func (m *MemoryBackend) CreateDefaultObjectACL(ctx context.Context, bucket string, acl *model.ObjectACL) (*model.ObjectACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	acl.Kind = "storage#objectAccessControl"
	acl.Bucket = bucket
	m.defaultObjectACLs[bucket] = append(m.defaultObjectACLs[bucket], acl)
	return acl, nil
}

func (m *MemoryBackend) UpdateDefaultObjectACL(ctx context.Context, bucket, entity string, acl *model.ObjectACL) (*model.ObjectACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.buckets[bucket]; !exists {
		return nil, ErrBucketNotFound
	}
	for i, a := range m.defaultObjectACLs[bucket] {
		if a.Entity == entity {
			acl.Kind = "storage#objectAccessControl"
			acl.Bucket = bucket
			acl.Entity = entity
			m.defaultObjectACLs[bucket][i] = acl
			return acl, nil
		}
	}
	return nil, ErrObjectNotFound
}

func (m *MemoryBackend) DeleteDefaultObjectACL(ctx context.Context, bucket, entity string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.buckets[bucket]; !exists {
		return ErrBucketNotFound
	}
	acls := m.defaultObjectACLs[bucket]
	for i, a := range acls {
		if a.Entity == entity {
			m.defaultObjectACLs[bucket] = append(acls[:i], acls[i+1:]...)
			return nil
		}
	}
	return ErrObjectNotFound
}

func (m *MemoryBackend) findLatestObject(bucket, name string) (*model.Object, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	obj, found := m.findLatestObjectLocked(bucket, name)
	if !found {
		return nil, ErrObjectNotFound
	}
	return obj, nil
}

func (m *MemoryBackend) findLatestObjectLocked(bucket, name string) (*model.Object, bool) {
	var latest *model.Object
	var latestGen int64
	for key, obj := range m.objects {
		if obj.Bucket == bucket && obj.Name == name {
			objKey := parseObjectKey(key)
			if objKey.generation > latestGen {
				latestGen = objKey.generation
				latest = obj
			}
		}
	}
	if latest == nil {
		return nil, false
	}
	return latest, true
}

func (m *MemoryBackend) checkObjectConditions(obj *model.Object, conds Conditions) error {
	if conds.DoesNotExist {
		return ErrPreconditionFailed
	}
	if conds.HasGenerationMatch && obj.Generation != conds.GenerationMatch {
		return ErrPreconditionFailed
	}
	if conds.HasGenerationNotMatch && obj.Generation == conds.GenerationNotMatch {
		return ErrPreconditionFailed
	}
	if conds.HasMetagenerationMatch && obj.Metageneration != conds.MetagenerationMatch {
		return ErrPreconditionFailed
	}
	if conds.HasMetagenerationNotMatch && obj.Metageneration == conds.MetagenerationNotMatch {
		return ErrPreconditionFailed
	}
	return nil
}

type objectKeyParts struct {
	bucket     string
	name       string
	generation int64
}

func objectKey(bucket, name string, generation int64) string {
	if generation > 0 {
		return bucket + "/" + name + "#" + strconv.FormatInt(generation, 10)
	}
	return bucket + "/" + name
}

func parseObjectKey(key string) objectKeyParts {
	var parts objectKeyParts
	bucketEnd := strings.Index(key, "/")
	if bucketEnd < 0 {
		parts.bucket = key
		return parts
	}
	parts.bucket = key[:bucketEnd]
	rest := key[bucketEnd+1:]
	hashPos := strings.LastIndex(rest, "#")
	if hashPos >= 0 {
		parts.name = rest[:hashPos]
		parts.generation, _ = strconv.ParseInt(rest[hashPos+1:], 10, 64)
	} else {
		parts.name = rest
	}
	return parts
}

func isLiveVersion(obj *model.Object, gen int64, objects map[string]*model.Object, bucket, name string) bool {
	var latestGen int64
	for key, o := range objects {
		if o.Bucket == bucket && o.Name == name {
			parts := parseObjectKey(key)
			if parts.generation > latestGen {
				latestGen = parts.generation
			}
		}
	}
	return gen == latestGen && obj.TimeDeleted.IsZero()
}

func (m *MemoryBackend) CleanupBlobs() error {
	if m.blobDir == "" {
		return nil
	}

	m.mu.RLock()
	knownKeys := make(map[string]bool, len(m.objects))
	for key := range m.objects {
		knownKeys[key] = true
	}
	m.mu.RUnlock()

	return filepath.Walk(m.blobDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(m.blobDir, path)
		if err != nil {
			return err
		}
		if !knownKeys[relPath] {
			os.Remove(path)
		}
		return nil
	})
}
