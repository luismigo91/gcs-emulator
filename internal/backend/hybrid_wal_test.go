package backend_test

import (
	"context"
	"testing"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func TestHybridBackend_FlushIncludesIAMAndNotifications(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewHybridBackend(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create hybrid backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "hybrid-bucket"}, backend.Conditions{})

	policy := &model.Policy{
		Bindings: []model.PolicyBinding{
			{Role: "roles/storage.admin", Members: []string{"user:admin@example.com"}},
		},
	}
	b.SetBucketIAMPolicy(ctx, "hybrid-bucket", policy)

	notif := &model.Notification{
		Topic:     "projects/test/topics/hybrid-topic",
		EventType: []string{"OBJECT_FINALIZE"},
	}
	b.CreateNotification(ctx, "hybrid-bucket", notif)

	time.Sleep(200 * time.Millisecond)
	b.Shutdown()

	b2, err := backend.NewHybridBackend(dir, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to recreate hybrid backend: %v", err)
	}
	defer b2.Shutdown()

	restoredPolicy, err := b2.GetBucketIAMPolicy(ctx, "hybrid-bucket")
	if err != nil {
		t.Fatalf("Failed to get IAM policy: %v", err)
	}
	if len(restoredPolicy.Bindings) != 1 || restoredPolicy.Bindings[0].Role != "roles/storage.admin" {
		t.Errorf("IAM policy not restored correctly: %+v", restoredPolicy)
	}

	notifs, err := b2.ListNotifications(ctx, "hybrid-bucket")
	if err != nil {
		t.Fatalf("Failed to list notifications: %v", err)
	}
	if len(notifs) != 1 || notifs[0].Topic != "projects/test/topics/hybrid-topic" {
		t.Errorf("Notification not restored correctly: %+v", notifs)
	}
}

func TestWALBackend_ReplayIAMAndNotifications(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewWALBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create WAL backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "wal-bucket"}, backend.Conditions{})

	policy := &model.Policy{
		Bindings: []model.PolicyBinding{
			{Role: "roles/storage.objectCreator", Members: []string{"user:creator@example.com"}},
		},
	}
	b.SetBucketIAMPolicy(ctx, "wal-bucket", policy)

	notif := &model.Notification{
		Topic:     "projects/test/topics/wal-topic",
		EventType: []string{"OBJECT_DELETE"},
	}
	b.CreateNotification(ctx, "wal-bucket", notif)
	b.Shutdown()

	b2, err := backend.NewWALBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate WAL backend: %v", err)
	}
	defer b2.Shutdown()

	restoredPolicy, err := b2.GetBucketIAMPolicy(ctx, "wal-bucket")
	if err != nil {
		t.Fatalf("Failed to get IAM policy after WAL replay: %v", err)
	}
	if len(restoredPolicy.Bindings) != 1 {
		t.Errorf("Expected 1 IAM binding after replay, got %d", len(restoredPolicy.Bindings))
	}

	notifs, err := b2.ListNotifications(ctx, "wal-bucket")
	if err != nil {
		t.Fatalf("Failed to list notifications after WAL replay: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification after replay, got %d", len(notifs))
	}

	err = b2.DeleteNotification(ctx, "wal-bucket", notifs[0].ID)
	if err != nil {
		t.Fatalf("Failed to delete notification: %v", err)
	}
	b2.Shutdown()

	b3, err := backend.NewWALBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate WAL backend 2: %v", err)
	}
	defer b3.Shutdown()

	notifs3, err := b3.ListNotifications(ctx, "wal-bucket")
	if err != nil {
		t.Fatalf("Failed to list notifications after 2nd replay: %v", err)
	}
	if len(notifs3) != 0 {
		t.Errorf("Expected 0 notifications after delete replay, got %d", len(notifs3))
	}
}

func TestWALBackend_CompactionIncludesIAM(t *testing.T) {
	dir := t.TempDir()
	b, err := backend.NewWALBackend(dir)
	if err != nil {
		t.Fatalf("Failed to create WAL backend: %v", err)
	}

	ctx := context.Background()
	b.CreateBucket(ctx, &model.Bucket{Name: "compact-bucket"}, backend.Conditions{})

	policy := &model.Policy{
		Bindings: []model.PolicyBinding{
			{Role: "roles/storage.legacyBucketReader", Members: []string{"allUsers"}},
		},
	}
	b.SetBucketIAMPolicy(ctx, "compact-bucket", policy)
	b.Shutdown()

	b2, err := backend.NewWALBackend(dir)
	if err != nil {
		t.Fatalf("Failed to recreate WAL backend: %v", err)
	}
	defer b2.Shutdown()

	restored, err := b2.GetBucketIAMPolicy(ctx, "compact-bucket")
	if err != nil {
		t.Fatalf("Failed to get IAM policy: %v", err)
	}
	if len(restored.Bindings) != 1 {
		t.Errorf("Expected 1 IAM binding after compaction, got %d", len(restored.Bindings))
	}
}
