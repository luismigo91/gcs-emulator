package backend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func BenchmarkBucketCreation(b *testing.B) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "bench-bucket-" + string(rune(i))
		be.CreateBucket(ctx, &model.Bucket{Name: name}, backend.Conditions{})
	}
}

func BenchmarkObjectUpload(b *testing.B) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	be.CreateBucket(ctx, &model.Bucket{Name: "bench-bucket"}, backend.Conditions{})

	data := strings.Repeat("x", 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "bench-obj-" + string(rune(i))
		be.CreateObject(ctx, "bench-bucket", &model.Object{Name: name, Bucket: "bench-bucket"}, strings.NewReader(data), backend.Conditions{})
	}
}

func BenchmarkObjectDownload(b *testing.B) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	be.CreateBucket(ctx, &model.Bucket{Name: "bench-bucket"}, backend.Conditions{})
	be.CreateObject(ctx, "bench-bucket", &model.Object{Name: "bench-obj", Bucket: "bench-bucket"}, strings.NewReader(strings.Repeat("y", 1024)), backend.Conditions{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = be.GetObjectContent(ctx, "bench-bucket", "bench-obj", 0)
	}
}

func BenchmarkListObjects(b *testing.B) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	be.CreateBucket(ctx, &model.Bucket{Name: "bench-list"}, backend.Conditions{})

	for j := 0; j < 100; j++ {
		name := "logs/2024/" + string(rune('a'+j%26)) + ".txt"
		be.CreateObject(ctx, "bench-list", &model.Object{Name: name, Bucket: "bench-list"}, strings.NewReader("x"), backend.Conditions{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = be.ListObjects(ctx, "bench-list", backend.ListObjectsParams{
			Prefix:    "logs/2024/",
			Delimiter: "/",
		})
	}
}
