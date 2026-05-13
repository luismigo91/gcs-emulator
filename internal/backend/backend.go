package backend

import (
	"context"
	"io"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

type Conditions struct {
	DoesNotExist         bool
	GenerationMatch      int64
	GenerationNotMatch   int64
	MetagenerationMatch  int64
	MetagenerationNotMatch int64
	HasGenerationMatch   bool
	HasGenerationNotMatch bool
	HasMetagenerationMatch bool
	HasMetagenerationNotMatch bool
}

func (c Conditions) IsZero() bool {
	return !c.DoesNotExist && !c.HasGenerationMatch && !c.HasGenerationNotMatch && !c.HasMetagenerationMatch && !c.HasMetagenerationNotMatch
}

type ListObjectsParams struct {
	Prefix                   string
	Delimiter                string
	MaxResults               int
	PageToken                string
	Versions                 bool
	StartOffset              string
	EndOffset                string
	IncludeTrailingDelimiter bool
}

type ListBucketsParams struct {
	Project string
}

type ListObjectsResponse struct {
	Items         []*model.Object
	Prefixes      []string
	NextPageToken string
}

type Backend interface {
	CreateBucket(ctx context.Context, bucket *model.Bucket, conds Conditions) error
	GetBucket(ctx context.Context, name string) (*model.Bucket, error)
	ListBuckets(ctx context.Context, params ListBucketsParams) ([]*model.Bucket, error)
	DeleteBucket(ctx context.Context, name string, conds Conditions) error
	UpdateBucket(ctx context.Context, name string, attrs *model.BucketUpdateAttrs, conds Conditions) (*model.Bucket, error)

	CreateObject(ctx context.Context, bucket string, object *model.Object, content io.Reader, conds Conditions) error
	GetObject(ctx context.Context, bucket, name string, generation int64) (*model.Object, error)
	GetObjectContent(ctx context.Context, bucket, name string, generation int64) (io.ReadSeeker, error)
	ListObjects(ctx context.Context, bucket string, params ListObjectsParams) (*ListObjectsResponse, error)
	DeleteObject(ctx context.Context, bucket, name string, generation int64, conds Conditions) error
	UpdateObject(ctx context.Context, bucket, name string, attrs *model.ObjectUpdateAttrs, conds Conditions) (*model.Object, error)
	ComposeObjects(ctx context.Context, bucket, destName string, sourceNames []string, destAttrs *model.Object) (*model.Object, error)
	CopyObject(ctx context.Context, srcBucket, srcName, destBucket, destName string, conds Conditions) (*model.Object, error)
	RewriteObject(ctx context.Context, srcBucket, srcName, destBucket, destName string, token string, conds Conditions) (*model.Object, string, error)

	GetBucketIAMPolicy(ctx context.Context, bucket string) (*model.Policy, error)
	SetBucketIAMPolicy(ctx context.Context, bucket string, policy *model.Policy) (*model.Policy, error)
	TestBucketIAMPermissions(ctx context.Context, bucket string, permissions []string) ([]string, error)

	CreateNotification(ctx context.Context, bucket string, notification *model.Notification) (*model.Notification, error)
	ListNotifications(ctx context.Context, bucket string) ([]*model.Notification, error)
	GetNotification(ctx context.Context, bucket, id string) (*model.Notification, error)
	DeleteNotification(ctx context.Context, bucket, id string) error

	GetLifecycle(ctx context.Context, bucket string) (*model.Lifecycle, error)
	SetLifecycle(ctx context.Context, bucket string, lifecycle *model.Lifecycle) (*model.Lifecycle, error)
	DeleteLifecycle(ctx context.Context, bucket string) error

	GetCORS(ctx context.Context, bucket string) ([]model.CORSRule, error)
	SetCORS(ctx context.Context, bucket string, rules []model.CORSRule) ([]model.CORSRule, error)

	ListBucketACL(ctx context.Context, bucket string) ([]*model.BucketACL, error)
	GetBucketACL(ctx context.Context, bucket, entity string) (*model.BucketACL, error)
	CreateBucketACL(ctx context.Context, bucket string, acl *model.BucketACL) (*model.BucketACL, error)
	UpdateBucketACL(ctx context.Context, bucket, entity string, acl *model.BucketACL) (*model.BucketACL, error)
	DeleteBucketACL(ctx context.Context, bucket, entity string) error

	ListObjectACL(ctx context.Context, bucket, object string, generation int64) ([]*model.ObjectACL, error)
	GetObjectACL(ctx context.Context, bucket, object, entity string, generation int64) (*model.ObjectACL, error)
	CreateObjectACL(ctx context.Context, bucket, object string, acl *model.ObjectACL) (*model.ObjectACL, error)
	UpdateObjectACL(ctx context.Context, bucket, object, entity string, acl *model.ObjectACL) (*model.ObjectACL, error)
	DeleteObjectACL(ctx context.Context, bucket, object, entity string) error

	GetWebsiteConfig(ctx context.Context, bucket string) (*model.WebsiteConfig, error)
	SetWebsiteConfig(ctx context.Context, bucket string, config *model.WebsiteConfig) (*model.WebsiteConfig, error)
	DeleteWebsiteConfig(ctx context.Context, bucket string) error

	GetEncryptionConfig(ctx context.Context, bucket string) (*model.EncryptionConfig, error)
	SetEncryptionConfig(ctx context.Context, bucket string, config *model.EncryptionConfig) (*model.EncryptionConfig, error)

	ListDefaultObjectACL(ctx context.Context, bucket string) ([]*model.ObjectACL, error)
	CreateDefaultObjectACL(ctx context.Context, bucket string, acl *model.ObjectACL) (*model.ObjectACL, error)
	UpdateDefaultObjectACL(ctx context.Context, bucket, entity string, acl *model.ObjectACL) (*model.ObjectACL, error)
	DeleteDefaultObjectACL(ctx context.Context, bucket, entity string) error

	Shutdown() error
}
