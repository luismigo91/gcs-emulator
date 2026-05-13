package backend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func TestMemoryBackend_CreateAndGetBucket(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	bucket := &model.Bucket{
		Name: "test-bucket",
	}

	if err := b.CreateBucket(ctx, bucket, backend.Conditions{}); err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	got, err := b.GetBucket(ctx, "test-bucket")
	if err != nil {
		t.Fatalf("Failed to get bucket: %v", err)
	}

	if got.Name != "test-bucket" {
		t.Errorf("Expected bucket name 'test-bucket', got '%s'", got.Name)
	}
}

func TestMemoryBackend_ListBuckets(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "bucket1"}, backend.Conditions{})
	b.CreateBucket(ctx, &model.Bucket{Name: "bucket2"}, backend.Conditions{})

	buckets, err := b.ListBuckets(ctx, backend.ListBucketsParams{})
	if err != nil {
		t.Fatalf("Failed to list buckets: %v", err)
	}

	if len(buckets) != 2 {
		t.Errorf("Expected 2 buckets, got %d", len(buckets))
	}
}

func TestMemoryBackend_DeleteBucket(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	bucket := &model.Bucket{Name: "test-bucket"}
	b.CreateBucket(ctx, bucket, backend.Conditions{})

	if err := b.DeleteBucket(ctx, "test-bucket", backend.Conditions{}); err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}

	_, err := b.GetBucket(ctx, "test-bucket")
	if err != backend.ErrBucketNotFound {
		t.Errorf("Expected ErrBucketNotFound, got %v", err)
	}
}

func TestMemoryBackend_CreateAndGetObject(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	obj := &model.Object{
		Name:   "test-object",
		Bucket: "test-bucket",
	}

	content := strings.NewReader("hello world")
	if err := b.CreateObject(ctx, "test-bucket", obj, content, backend.Conditions{}); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	got, err := b.GetObject(ctx, "test-bucket", "test-object", 0)
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	if got.Name != "test-object" {
		t.Errorf("Expected object name 'test-object', got '%s'", got.Name)
	}
}

func TestMemoryBackend_ListObjectsWithPrefixes(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	objects := []string{"logs/2024/jan.txt", "logs/2024/feb.txt", "logs/2025/mar.txt"}
	for _, name := range objects {
		obj := &model.Object{Name: name, Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("content"), backend.Conditions{})
	}

	resp, err := b.ListObjects(ctx, "test-bucket", backend.ListObjectsParams{
		Prefix:    "logs/",
		Delimiter: "/",
	})
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(resp.Prefixes) != 2 {
		t.Errorf("Expected 2 prefixes, got %d: %v", len(resp.Prefixes), resp.Prefixes)
	}
}

func TestMemoryBackend_Generations(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket", Versioning: &model.Versioning{Enabled: true}}, backend.Conditions{})

	obj1 := &model.Object{Name: "test.txt", Bucket: "test-bucket"}
	b.CreateObject(ctx, "test-bucket", obj1, strings.NewReader("v1"), backend.Conditions{})

	obj2 := &model.Object{Name: "test.txt", Bucket: "test-bucket"}
	b.CreateObject(ctx, "test-bucket", obj2, strings.NewReader("v2"), backend.Conditions{})

	if obj1.Generation == obj2.Generation {
		t.Errorf("Expected different generations, got %d", obj1.Generation)
	}

	resp, err := b.ListObjects(ctx, "test-bucket", backend.ListObjectsParams{Versions: true})
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(resp.Items))
	}
}
