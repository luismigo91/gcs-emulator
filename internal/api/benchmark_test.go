package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func newBenchmarkHandler(b *testing.B) http.Handler {
	b.Helper()
	mem := backend.NewMemoryBackend()
	return router.New(mem, nil, nil, nil, nil, nil, "test-project")
}

func BenchmarkBucketOperations(b *testing.B) {
	handler := newBenchmarkHandler(b)

	b.Run("CreateBucket", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			name := "bench-bucket-" + string(rune('a'+i%26)) + "-" + string(rune(i%10+'0'))
			body := bytes.NewBufferString(`{"name":"` + name + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", body)
			req.Header.Set("X-Goog-User-Project", "test-project")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkObjectOperations(b *testing.B) {
	handler := newBenchmarkHandler(b)
	bkd := backend.NewMemoryBackend()

	bkd.CreateBucket(b.Context(), &model.Bucket{Name: "bench-bucket"}, backend.Conditions{})
	b.ResetTimer()

	b.Run("CreateObject", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			name := "obj-" + string(rune('a'+i%26)) + ".txt"
			req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/bench-bucket/o?name="+name, strings.NewReader("benchmark content"))
			req.Header.Set("X-Goog-User-Project", "test-project")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})

	b.Run("GetObject", func(b *testing.B) {
		bkd.CreateObject(b.Context(), "bench-bucket", &model.Object{Name: "bench-obj.txt", Bucket: "bench-bucket"}, strings.NewReader("benchmark content"), backend.Conditions{})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/bench-bucket/o/bench-obj.txt", nil)
			req.Header.Set("X-Goog-User-Project", "test-project")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})

	b.Run("ListObjects", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			name := "list-obj-" + string(rune('a'+i%26)) + ".txt"
			bkd.CreateObject(b.Context(), "bench-bucket", &model.Object{Name: name, Bucket: "bench-bucket"}, strings.NewReader("content"), backend.Conditions{})
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/bench-bucket/o", nil)
			req.Header.Set("X-Goog-User-Project", "test-project")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkConcurrentOperations(b *testing.B) {
	handler := newBenchmarkHandler(b)
	bkd := backend.NewMemoryBackend()

	bkd.CreateBucket(b.Context(), &model.Bucket{Name: "concurrent-bench-bucket"}, backend.Conditions{})
	b.ResetTimer()

	b.Run("ConcurrentObjectCreation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				name := "concurrent-obj-" + string(rune('a'+i%26)) + "-" + string(rune(i%10+'0')) + ".txt"
				req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/concurrent-bench-bucket/o?name="+name, strings.NewReader("concurrent content"))
				req.Header.Set("X-Goog-User-Project", "test-project")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				i++
			}
		})
	})
}
