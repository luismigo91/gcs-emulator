package backend_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func TestPersistentBackend_BlobRoundtrip(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create persistent backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	obj := &model.Object{Name: "test.txt", Bucket: "test-bucket"}
	content := strings.NewReader("hello persistent world")
	if err := b.CreateObject(ctx, "test-bucket", obj, content, backend.Conditions{}); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	rc, err := b.GetObjectContent(ctx, "test-bucket", "test.txt", 0)
	if err != nil {
		t.Fatalf("Failed to get object content: %v", err)
	}
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}
	if string(data) != "hello persistent world" {
		t.Errorf("Expected 'hello persistent world', got '%s'", string(data))
	}

	b.Shutdown()

	b2, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate backend: %v", err)
	}
	defer b2.Shutdown()

	rc2, err := b2.GetObjectContent(ctx, "test-bucket", "test.txt", 0)
	if err != nil {
		t.Fatalf("Failed to get object content after restart: %v", err)
	}
	data2, err := io.ReadAll(rc2)
	if err != nil {
		t.Fatalf("Failed to read content after restart: %v", err)
	}
	if string(data2) != "hello persistent world" {
		t.Errorf("Content after restart: expected 'hello persistent world', got '%s'", string(data2))
	}
}

func TestPersistentBackend_BlobCleanupOnDelete(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create persistent backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	obj := &model.Object{Name: "cleanup.txt", Bucket: "test-bucket"}
	if err := b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{}); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	if err := b.DeleteObject(ctx, "test-bucket", "cleanup.txt", 0, backend.Conditions{}); err != nil {
		t.Fatalf("Failed to delete object: %v", err)
	}

	b.Shutdown()

	b2, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate backend: %v", err)
	}
	defer b2.Shutdown()

	_, err = b2.GetObject(ctx, "test-bucket", "cleanup.txt", 0)
	if err != backend.ErrObjectNotFound {
		t.Errorf("Expected ErrObjectNotFound after delete and restart, got %v", err)
	}
}

func TestPersistentBackend_IAMPolicySurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create persistent backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "iam-bucket"}, backend.Conditions{})

	policy := &model.Policy{
		Bindings: []model.PolicyBinding{
			{Role: "roles/storage.objectViewer", Members: []string{"user:test@example.com"}},
		},
	}
	b.SetBucketIAMPolicy(ctx, "iam-bucket", policy)
	b.Shutdown()

	b2, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate backend: %v", err)
	}
	defer b2.Shutdown()

	restored, err := b2.GetBucketIAMPolicy(ctx, "iam-bucket")
	if err != nil {
		t.Fatalf("Failed to get IAM policy after restart: %v", err)
	}

	if len(restored.Bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(restored.Bindings))
	}
	if restored.Bindings[0].Role != "roles/storage.objectViewer" {
		t.Errorf("Expected role 'roles/storage.objectViewer', got '%s'", restored.Bindings[0].Role)
	}
}

func TestPersistentBackend_NotificationSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create persistent backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "notif-bucket"}, backend.Conditions{})

	notif := &model.Notification{
		Topic:     "projects/test/topics/test-topic",
		EventType: []string{"OBJECT_FINALIZE"},
	}
	b.CreateNotification(ctx, "notif-bucket", notif)
	b.Shutdown()

	b2, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate backend: %v", err)
	}
	defer b2.Shutdown()

	notifs, err := b2.ListNotifications(ctx, "notif-bucket")
	if err != nil {
		t.Fatalf("Failed to list notifications after restart: %v", err)
	}

	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
	if len(notifs[0].EventType) != 1 || notifs[0].EventType[0] != "OBJECT_FINALIZE" {
		t.Errorf("Unexpected notification event types: %v", notifs[0].EventType)
	}
}

func TestBackwardCompatibleLoad_OldFormat(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	oldData := `{"buckets":{"bucket1":{"kind":"storage#bucket","name":"bucket1"}},"objects":{}}`
	if err := os.WriteFile(dir+"/data.json", []byte(oldData), 0644); err != nil {
		t.Fatalf("Failed to write old data: %v", err)
	}

	b, err := backend.NewPersistentBackend(dir)
	if err != nil {
		t.Fatalf("Failed to load old format: %v", err)
	}
	defer b.Shutdown()

	bucket, err := b.GetBucket(context.Background(), "bucket1")
	if err != nil {
		t.Fatalf("Failed to get bucket from old format: %v", err)
	}
	if bucket.Name != "bucket1" {
		t.Errorf("Expected 'bucket1', got '%s'", bucket.Name)
	}
}
