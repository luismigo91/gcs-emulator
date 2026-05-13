package emulator_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func newSDKClient(t *testing.T, srv *httptest.Server) *storage.Client {
	t.Helper()
	client, err := storage.NewClient(
		context.Background(),
		option.WithEndpoint(srv.URL+"/storage/v1/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("Failed to create storage client: %v", err)
	}
	return client
}

func newSDKTestServer(t *testing.T) (*httptest.Server, *storage.Client, backend.Backend) {
	t.Helper()
	b := backend.NewMemoryBackend()
	mux := router.New(b, nil, nil, nil, nil, "test-project")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := newSDKClient(t, srv)
	t.Cleanup(func() { client.Close() })
	return srv, client, b
}

func TestSDKBucketCRUD(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	t.Run("Create bucket", func(t *testing.T) {
		b := client.Bucket("sdk-bucket-create")
		if err := b.Create(ctx, "test-project", nil); err != nil {
			t.Fatalf("Failed to create bucket: %v", err)
		}
	})

	t.Run("Get bucket metadata", func(t *testing.T) {
		b := client.Bucket("sdk-bucket-create")
		attrs, err := b.Attrs(ctx)
		if err != nil {
			t.Fatalf("Failed to get bucket attrs: %v", err)
		}
		if attrs.Name != "sdk-bucket-create" {
			t.Errorf("Expected name sdk-bucket-create, got %s", attrs.Name)
		}
	})

	t.Run("List buckets", func(t *testing.T) {
		it := client.Buckets(ctx, "test-project")
		count := 0
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Failed to list buckets: %v", err)
			}
			count++
		}
		if count < 1 {
			t.Errorf("Expected at least 1 bucket, got %d", count)
		}
	})

	t.Run("Delete bucket", func(t *testing.T) {
		b := client.Bucket("sdk-bucket-create")
		if err := b.Delete(ctx); err != nil {
			t.Fatalf("Failed to delete bucket: %v", err)
		}
	})
}

func TestSDKObjectCRUD(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	client.Bucket("sdk-bucket").Create(ctx, "test-project", nil)

	t.Run("Create object", func(t *testing.T) {
		obj := client.Bucket("sdk-bucket").Object("test.txt")
		w := obj.NewWriter(ctx)
		if _, err := w.Write([]byte("hello from sdk")); err != nil {
			t.Fatalf("Failed to write object: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}
	})

	t.Run("Get object metadata", func(t *testing.T) {
		obj := client.Bucket("sdk-bucket").Object("test.txt")
		attrs, err := obj.Attrs(ctx)
		if err != nil {
			t.Fatalf("Failed to get object attrs: %v", err)
		}
		if attrs.Name != "test.txt" {
			t.Errorf("Expected name test.txt, got %s", attrs.Name)
		}
		if attrs.Bucket != "sdk-bucket" {
			t.Errorf("Expected bucket sdk-bucket, got %s", attrs.Bucket)
		}
	})

	t.Run("Download object", func(t *testing.T) {
		obj := client.Bucket("sdk-bucket").Object("test.txt")
		r, err := obj.NewReader(ctx)
		if err != nil {
			t.Fatalf("Failed to create reader: %v", err)
		}
		defer r.Close()

		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("Failed to read object: %v", err)
		}
		if string(data) != "hello from sdk" {
			t.Errorf("Expected 'hello from sdk', got '%s'", string(data))
		}
	})

	t.Run("Update object metadata", func(t *testing.T) {
		obj := client.Bucket("sdk-bucket").Object("test.txt")
		_, err := obj.Update(ctx, storage.ObjectAttrsToUpdate{
			ContentType: "text/plain",
			Metadata:    map[string]string{"key": "value"},
		})
		if err != nil {
			t.Fatalf("Failed to update object: %v", err)
		}

		attrs, err := obj.Attrs(ctx)
		if err != nil {
			t.Fatalf("Failed to get updated attrs: %v", err)
		}
		if attrs.ContentType != "text/plain" {
			t.Errorf("Expected content-type text/plain, got %s", attrs.ContentType)
		}
		if attrs.Metadata["key"] != "value" {
			t.Errorf("Expected metadata key=value, got %v", attrs.Metadata)
		}
	})

	t.Run("Delete object", func(t *testing.T) {
		obj := client.Bucket("sdk-bucket").Object("test.txt")
		if err := obj.Delete(ctx); err != nil {
			t.Fatalf("Failed to delete object: %v", err)
		}

		_, err := obj.Attrs(ctx)
		if err == nil {
			t.Error("Expected error after delete, got nil")
		}
	})
}

func TestSDKListObjects(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	client.Bucket("sdk-list-bucket").Create(ctx, "test-project", nil)

	names := []string{"a.txt", "b.txt", "c.txt", "dir/1.txt", "dir/2.txt"}
	for _, name := range names {
		obj := client.Bucket("sdk-list-bucket").Object(name)
		w := obj.NewWriter(ctx)
		w.Write([]byte("content"))
		w.Close()
	}

	t.Run("List all objects", func(t *testing.T) {
		it := client.Bucket("sdk-list-bucket").Objects(ctx, nil)
		count := 0
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Failed to list objects: %v", err)
			}
			count++
		}
		if count != 5 {
			t.Errorf("Expected 5 objects, got %d", count)
		}
	})

	t.Run("List with prefix", func(t *testing.T) {
		it := client.Bucket("sdk-list-bucket").Objects(ctx, &storage.Query{Prefix: "dir/"})
		count := 0
		for {
			obj, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Failed to list objects: %v", err)
			}
			count++
			if !strings.HasPrefix(obj.Name, "dir/") {
				t.Errorf("Expected prefix dir/, got %s", obj.Name)
			}
		}
		if count != 2 {
			t.Errorf("Expected 2 objects with prefix dir/, got %d", count)
		}
	})

	t.Run("List with delimiter", func(t *testing.T) {
		it := client.Bucket("sdk-list-bucket").Objects(ctx, &storage.Query{Delimiter: "/"})
		objCount := 0
		prefixCount := 0
		for {
			obj, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Failed to list objects: %v", err)
			}
			if obj.Prefix != "" {
				prefixCount++
			} else {
				objCount++
			}
		}
		if objCount != 3 {
			t.Errorf("Expected 3 objects at root, got %d", objCount)
		}
		if prefixCount != 1 {
			t.Errorf("Expected 1 prefix (dir/), got %d", prefixCount)
		}
	})
}

func TestSDKUploadDownload(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	client.Bucket("sdk-io-bucket").Create(ctx, "test-project", nil)

	t.Run("Upload and download", func(t *testing.T) {
		content := bytes.Repeat([]byte("x"), 1024*100)
		obj := client.Bucket("sdk-io-bucket").Object("large.bin")

		w := obj.NewWriter(ctx)
		w.ContentType = "application/octet-stream"
		if _, err := w.Write(content); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Failed to close: %v", err)
		}

		r, err := obj.NewReader(ctx)
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}
		defer r.Close()

		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("Failed to read all: %v", err)
		}
		if len(data) != len(content) {
			t.Errorf("Expected %d bytes, got %d", len(content), len(data))
		}
	})

	t.Run("Copy object", func(t *testing.T) {
		src := client.Bucket("sdk-io-bucket").Object("large.bin")
		
		attrs, err := src.Attrs(ctx)
		if err != nil {
			t.Fatalf("Source object doesn't exist: %v", err)
		}
		t.Logf("Source object: %s, size: %d", attrs.Name, attrs.Size)
		
		dst := client.Bucket("sdk-io-bucket").Object("copy.bin")

		if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
			t.Fatalf("Failed to copy object: %v", err)
		}

		r, err := dst.NewReader(ctx)
		if err != nil {
			t.Fatalf("Failed to read copy: %v", err)
		}
		defer r.Close()

		data, _ := io.ReadAll(r)
		if len(data) != 1024*100 {
			t.Errorf("Expected 102400 bytes, got %d", len(data))
		}
	})

	t.Run("Compose objects", func(t *testing.T) {
		obj1 := client.Bucket("sdk-io-bucket").Object("part1.txt")
		w1 := obj1.NewWriter(ctx)
		w1.Write([]byte("hello"))
		w1.Close()

		obj2 := client.Bucket("sdk-io-bucket").Object("part2.txt")
		w2 := obj2.NewWriter(ctx)
		w2.Write([]byte(" world"))
		w2.Close()

		dst := client.Bucket("sdk-io-bucket").Object("composed.txt")
		c := dst.ComposerFrom(obj1, obj2)
		c.ContentType = "text/plain"
		if _, err := c.Run(ctx); err != nil {
			t.Fatalf("Failed to compose: %v", err)
		}

		r, err := dst.NewReader(ctx)
		if err != nil {
			t.Fatalf("Failed to read composed: %v", err)
		}
		defer r.Close()

		data, _ := io.ReadAll(r)
		if string(data) != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", string(data))
		}
	})
}

func TestSDKVersioning(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	client.Bucket("sdk-version-bucket").Create(ctx, "test-project", &storage.BucketAttrs{
		VersioningEnabled: true,
	})

	t.Run("Object versions", func(t *testing.T) {
		obj := client.Bucket("sdk-version-bucket").Object("versioned.txt")

		w := obj.NewWriter(ctx)
		w.Write([]byte("v1"))
		w.Close()

		w = obj.NewWriter(ctx)
		w.Write([]byte("v2"))
		w.Close()

		it := client.Bucket("sdk-version-bucket").Objects(ctx, &storage.Query{
			Versions: true,
		})
		count := 0
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Failed to list versions: %v", err)
			}
			count++
		}
		if count < 2 {
			t.Errorf("Expected at least 2 versions, got %d", count)
		}
	})
}

func TestSDKCompatibilityReport(t *testing.T) {
	_, client, _ := newSDKTestServer(t)
	ctx := context.Background()

	client.Bucket("sdk-report-bucket").Create(ctx, "test-project", nil)

	report := make(map[string]string)

	t.Run("Bucket.Create", func(t *testing.T) {
		err := client.Bucket("sdk-report-bucket").Create(ctx, "test-project", nil)
		if err != nil {
			report["Bucket.Create"] = "FAIL: " + err.Error()
		} else {
			report["Bucket.Create"] = "PASS"
		}
	})

	t.Run("Bucket.Attrs", func(t *testing.T) {
		_, err := client.Bucket("sdk-report-bucket").Attrs(ctx)
		if err != nil {
			report["Bucket.Attrs"] = "FAIL: " + err.Error()
		} else {
			report["Bucket.Attrs"] = "PASS"
		}
	})

	t.Run("Object.NewWriter", func(t *testing.T) {
		obj := client.Bucket("sdk-report-bucket").Object("report.txt")
		w := obj.NewWriter(ctx)
		w.Write([]byte("report"))
		err := w.Close()
		if err != nil {
			report["Object.NewWriter"] = "FAIL: " + err.Error()
		} else {
			report["Object.NewWriter"] = "PASS"
		}
	})

	t.Run("Object.NewReader", func(t *testing.T) {
		obj := client.Bucket("sdk-report-bucket").Object("report.txt")
		r, err := obj.NewReader(ctx)
		if err != nil {
			report["Object.NewReader"] = "FAIL: " + err.Error()
		} else {
			r.Close()
			report["Object.NewReader"] = "PASS"
		}
	})

	t.Run("Object.Attrs", func(t *testing.T) {
		_, err := client.Bucket("sdk-report-bucket").Object("report.txt").Attrs(ctx)
		if err != nil {
			report["Object.Attrs"] = "FAIL: " + err.Error()
		} else {
			report["Object.Attrs"] = "PASS"
		}
	})

	t.Run("Object.Update", func(t *testing.T) {
		_, err := client.Bucket("sdk-report-bucket").Object("report.txt").Update(ctx, storage.ObjectAttrsToUpdate{
			ContentType: "text/plain",
		})
		if err != nil {
			report["Object.Update"] = "FAIL: " + err.Error()
		} else {
			report["Object.Update"] = "PASS"
		}
	})

	t.Run("Object.Delete", func(t *testing.T) {
		err := client.Bucket("sdk-report-bucket").Object("report.txt").Delete(ctx)
		if err != nil {
			report["Object.Delete"] = "FAIL: " + err.Error()
		} else {
			report["Object.Delete"] = "PASS"
		}
	})

	t.Run("Objects.List", func(t *testing.T) {
		it := client.Bucket("sdk-report-bucket").Objects(ctx, nil)
		count := 0
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				report["Objects.List"] = "FAIL: " + err.Error()
				return
			}
			count++
		}
		report["Objects.List"] = "PASS"
	})

	t.Run("Bucket.Delete", func(t *testing.T) {
		err := client.Bucket("sdk-report-bucket").Delete(ctx)
		if err != nil {
			report["Bucket.Delete"] = "FAIL: " + err.Error()
		} else {
			report["Bucket.Delete"] = "PASS"
		}
	})

	t.Log("=== SDK Compatibility Report ===")
	for op, status := range report {
		t.Logf("  %s: %s", op, status)
	}

	passCount := 0
	failCount := 0
	for _, status := range report {
		if strings.HasPrefix(status, "PASS") {
			passCount++
		} else {
			failCount++
		}
	}
	t.Logf("=== Summary: %d passed, %d failed ===", passCount, failCount)
}
