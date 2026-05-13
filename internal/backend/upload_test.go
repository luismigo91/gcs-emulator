package backend_test

import (
	"testing"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
)

func TestUploadSessionStore(t *testing.T) {
	store := backend.NewUploadSessionStore()

	t.Run("create and get session", func(t *testing.T) {
		session := &backend.UploadSession{
			ID:         "test-session-1",
			Bucket:     "test-bucket",
			ObjectName: "test.txt",
			Chunks:     make(map[int64][]byte),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		}
		store.Create(session)

		got, exists := store.Get("test-session-1")
		if !exists {
			t.Fatal("Expected session to exist")
		}
		if got.Bucket != "test-bucket" {
			t.Errorf("Expected bucket 'test-bucket', got '%s'", got.Bucket)
		}
		if got.ObjectName != "test.txt" {
			t.Errorf("Expected object 'test.txt', got '%s'", got.ObjectName)
		}
	})

	t.Run("get non-existent session", func(t *testing.T) {
		_, exists := store.Get("non-existent")
		if exists {
			t.Error("Expected session to not exist")
		}
	})

	t.Run("store and merge chunks", func(t *testing.T) {
		session := &backend.UploadSession{
			ID:         "test-session-2",
			Bucket:     "test-bucket",
			ObjectName: "chunked.txt",
			Chunks:     make(map[int64][]byte),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		}
		store.Create(session)

		store.StoreChunk("test-session-2", 0, []byte("hello"))
		store.StoreChunk("test-session-2", 5, []byte(" world"))

		data := store.MergeChunks("test-session-2")
		if string(data) != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", string(data))
		}
	})

	t.Run("delete session", func(t *testing.T) {
		store.Delete("test-session-1")
		_, exists := store.Get("test-session-1")
		if exists {
			t.Error("Expected session to be deleted")
		}
	})

	t.Run("expired session not returned", func(t *testing.T) {
		session := &backend.UploadSession{
			ID:         "expired-session",
			Bucket:     "test-bucket",
			ObjectName: "expired.txt",
			Chunks:     make(map[int64][]byte),
			CreatedAt:  time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt:  time.Now().Add(-1 * time.Hour),
		}
		store.Create(session)

		_, exists := store.Get("expired-session")
		if exists {
			t.Error("Expected expired session to not be returned")
		}
	})

	t.Run("out of order chunks merge correctly", func(t *testing.T) {
		session := &backend.UploadSession{
			ID:         "test-session-3",
			Bucket:     "test-bucket",
			ObjectName: "unordered.txt",
			Chunks:     make(map[int64][]byte),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		}
		store.Create(session)

		store.StoreChunk("test-session-3", 10, []byte("world"))
		store.StoreChunk("test-session-3", 0, []byte("hello "))

		data := store.MergeChunks("test-session-3")
		if string(data) != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", string(data))
		}
	})
}
