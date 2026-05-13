package backend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func TestListObjectsWithPrefixes(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	objects := []string{
		"logs/2024/jan.txt",
		"logs/2024/feb.txt",
		"logs/2025/mar.txt",
		"docs/readme.md",
		"docs/api/guide.md",
	}
	for _, name := range objects {
		obj := &model.Object{Name: name, Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("content"), backend.Conditions{})
	}

	t.Run("with delimiter returns prefixes", func(t *testing.T) {
		resp, err := b.ListObjects(ctx, "test-bucket", backend.ListObjectsParams{
			Prefix:    "",
			Delimiter: "/",
		})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(resp.Prefixes) != 2 {
			t.Errorf("Expected 2 prefixes, got %d: %v", len(resp.Prefixes), resp.Prefixes)
		}

		prefixSet := make(map[string]bool)
		for _, p := range resp.Prefixes {
			prefixSet[p] = true
		}
		if !prefixSet["logs/"] {
			t.Error("Expected prefix 'logs/'")
		}
		if !prefixSet["docs/"] {
			t.Error("Expected prefix 'docs/'")
		}
	})

	t.Run("nested prefixes", func(t *testing.T) {
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

		prefixSet := make(map[string]bool)
		for _, p := range resp.Prefixes {
			prefixSet[p] = true
		}
		if !prefixSet["logs/2024/"] {
			t.Error("Expected prefix 'logs/2024/'")
		}
		if !prefixSet["logs/2025/"] {
			t.Error("Expected prefix 'logs/2025/'")
		}
	})

	t.Run("includeTrailingDelimiter", func(t *testing.T) {
		b2 := backend.NewMemoryBackend()
		ctx2 := context.Background()
		b2.CreateBucket(ctx2, &model.Bucket{Name: "test-bucket2"}, backend.Conditions{})

		b2.CreateObject(ctx2, "test-bucket2", &model.Object{Name: "logs/2024/", Bucket: "test-bucket2"}, strings.NewReader(""), backend.Conditions{})
		b2.CreateObject(ctx2, "test-bucket2", &model.Object{Name: "logs/2024/jan.txt", Bucket: "test-bucket2"}, strings.NewReader("content"), backend.Conditions{})

		resp, err := b2.ListObjects(ctx2, "test-bucket2", backend.ListObjectsParams{
			Prefix:                   "logs/",
			Delimiter:                "/",
			IncludeTrailingDelimiter: true,
		})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		hasPrefixObject := false
		for _, obj := range resp.Items {
			if obj.Name == "logs/2024/" {
				hasPrefixObject = true
				break
			}
		}

		if !hasPrefixObject {
			t.Errorf("Expected prefix object 'logs/2024/' in items when includeTrailingDelimiter is true, got items: %v", resp.Items)
		}
	})
}

func TestListObjectsPagination(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	for i := 0; i < 20; i++ {
		name := strings.Repeat("a", i+1) + ".txt"
		obj := &model.Object{Name: name, Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("content"), backend.Conditions{})
	}

	t.Run("maxResults limits items", func(t *testing.T) {
		resp, err := b.ListObjects(ctx, "test-bucket", backend.ListObjectsParams{
			MaxResults: 5,
		})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(resp.Items) != 5 {
			t.Errorf("Expected 5 items, got %d", len(resp.Items))
		}

		if resp.NextPageToken == "" {
			t.Error("Expected nextPageToken when maxResults is set")
		}
	})

	t.Run("startOffset and endOffset", func(t *testing.T) {
		resp, err := b.ListObjects(ctx, "test-bucket", backend.ListObjectsParams{
			StartOffset: "aa",
			EndOffset:   "aaaaa",
		})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		for _, obj := range resp.Items {
			if obj.Name < "aa" || obj.Name >= "aaaaa" {
				t.Errorf("Object %q outside range [aa, aaaaa)", obj.Name)
			}
		}
	})
}

func TestConditionalRequests(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	t.Run("DoesNotExist allows create", func(t *testing.T) {
		obj := &model.Object{Name: "new.txt", Bucket: "test-bucket"}
		err := b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{DoesNotExist: true})
		if err != nil {
			t.Errorf("Expected create to succeed with DoesNotExist, got: %v", err)
		}
	})

	t.Run("DoesNotExist prevents overwrite", func(t *testing.T) {
		obj := &model.Object{Name: "new.txt", Bucket: "test-bucket"}
		err := b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data2"), backend.Conditions{DoesNotExist: true})
		if err != backend.ErrPreconditionFailed {
			t.Errorf("Expected ErrPreconditionFailed, got: %v", err)
		}
	})

	t.Run("GenerationMatch succeeds", func(t *testing.T) {
		obj := &model.Object{Name: "match.txt", Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{})

		_, err := b.UpdateObject(ctx, "test-bucket", "match.txt", &model.ObjectUpdateAttrs{
			ContentType: strPtr("text/plain"),
		}, backend.Conditions{
			GenerationMatch:      obj.Generation,
			HasGenerationMatch:   true,
		})
		if err != nil {
			t.Errorf("Expected update to succeed with matching generation, got: %v", err)
		}
	})

	t.Run("GenerationMatch fails", func(t *testing.T) {
		obj := &model.Object{Name: "nomatch.txt", Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{})

		_, err := b.UpdateObject(ctx, "test-bucket", "nomatch.txt", &model.ObjectUpdateAttrs{
			ContentType: strPtr("text/plain"),
		}, backend.Conditions{
			GenerationMatch:      obj.Generation + 1,
			HasGenerationMatch:   true,
		})
		if err != backend.ErrPreconditionFailed {
			t.Errorf("Expected ErrPreconditionFailed, got: %v", err)
		}
	})

	t.Run("GenerationNotMatch succeeds", func(t *testing.T) {
		obj := &model.Object{Name: "notmatch.txt", Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{})

		_, err := b.UpdateObject(ctx, "test-bucket", "notmatch.txt", &model.ObjectUpdateAttrs{
			ContentType: strPtr("text/plain"),
		}, backend.Conditions{
			GenerationNotMatch:    obj.Generation + 1,
			HasGenerationNotMatch: true,
		})
		if err != nil {
			t.Errorf("Expected update to succeed with non-matching generation, got: %v", err)
		}
	})

	t.Run("GenerationNotMatch fails", func(t *testing.T) {
		obj := &model.Object{Name: "notmatch2.txt", Bucket: "test-bucket"}
		b.CreateObject(ctx, "test-bucket", obj, strings.NewReader("data"), backend.Conditions{})

		_, err := b.UpdateObject(ctx, "test-bucket", "notmatch2.txt", &model.ObjectUpdateAttrs{
			ContentType: strPtr("text/plain"),
		}, backend.Conditions{
			GenerationNotMatch:    obj.Generation,
			HasGenerationNotMatch: true,
		})
		if err != backend.ErrPreconditionFailed {
			t.Errorf("Expected ErrPreconditionFailed, got: %v", err)
		}
	})

	t.Run("MetagenerationMatch", func(t *testing.T) {
		b.CreateBucket(ctx, &model.Bucket{Name: "meta-bucket"}, backend.Conditions{})

		_, err := b.UpdateBucket(ctx, "meta-bucket", &model.BucketUpdateAttrs{
			Labels: map[string]string{"key": "value"},
		}, backend.Conditions{
			MetagenerationMatch:      1,
			HasMetagenerationMatch:   true,
		})
		if err != nil {
			t.Errorf("Expected update to succeed with matching metageneration, got: %v", err)
		}
	})

	t.Run("MetagenerationMatch fails", func(t *testing.T) {
		_, err := b.UpdateBucket(ctx, "meta-bucket", &model.BucketUpdateAttrs{
			Labels: map[string]string{"key": "value2"},
		}, backend.Conditions{
			MetagenerationMatch:      999,
			HasMetagenerationMatch:   true,
		})
		if err != backend.ErrPreconditionFailed {
			t.Errorf("Expected ErrPreconditionFailed, got: %v", err)
		}
	})
}

func TestComposeObjects(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	src1 := &model.Object{Name: "part1.txt", Bucket: "test-bucket"}
	b.CreateObject(ctx, "test-bucket", src1, strings.NewReader("hello"), backend.Conditions{})

	src2 := &model.Object{Name: "part2.txt", Bucket: "test-bucket"}
	b.CreateObject(ctx, "test-bucket", src2, strings.NewReader(" world"), backend.Conditions{})

	t.Run("compose two objects", func(t *testing.T) {
		result, err := b.ComposeObjects(ctx, "test-bucket", "combined.txt", []string{"part1.txt", "part2.txt"}, nil)
		if err != nil {
			t.Fatalf("Failed to compose objects: %v", err)
		}

		if result.ComponentCount != 2 {
			t.Errorf("Expected componentCount 2, got %d", result.ComponentCount)
		}

		content, err := b.GetObjectContent(ctx, "test-bucket", "combined.txt", 0)
		if err != nil {
			t.Fatalf("Failed to get composed object content: %v", err)
		}

		data := make([]byte, 11)
		content.Read(data)
		if string(data) != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", string(data))
		}
	})

	t.Run("compose with destination metadata", func(t *testing.T) {
		dest := &model.Object{
			ContentType: "application/octet-stream",
			Metadata:    map[string]string{"composed": "true"},
		}
		result, err := b.ComposeObjects(ctx, "test-bucket", "combined2.txt", []string{"part1.txt", "part2.txt"}, dest)
		if err != nil {
			t.Fatalf("Failed to compose objects: %v", err)
		}

		if result.ContentType != "application/octet-stream" {
			t.Errorf("Expected contentType 'application/octet-stream', got '%s'", result.ContentType)
		}
		if result.Metadata["composed"] != "true" {
			t.Errorf("Expected metadata composed=true, got %v", result.Metadata)
		}
	})

	t.Run("compose too many sources", func(t *testing.T) {
		sources := make([]string, 33)
		for i := range sources {
			name := strings.Repeat("x", i+1) + ".txt"
			src := &model.Object{Name: name, Bucket: "test-bucket"}
			b.CreateObject(ctx, "test-bucket", src, strings.NewReader("x"), backend.Conditions{})
			sources[i] = name
		}

		_, err := b.ComposeObjects(ctx, "test-bucket", "too-many.txt", sources, nil)
		if err != backend.ErrComposeTooManySources {
			t.Errorf("Expected ErrComposeTooManySources, got: %v", err)
		}
	})
}

func TestVersioning(t *testing.T) {
	b := backend.NewMemoryBackend()
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{
		Name:       "versioned-bucket",
		Versioning: &model.Versioning{Enabled: true},
	}, backend.Conditions{})

	t.Run("overwrite preserves old generation", func(t *testing.T) {
		obj1 := &model.Object{Name: "versioned.txt", Bucket: "versioned-bucket"}
		b.CreateObject(ctx, "versioned-bucket", obj1, strings.NewReader("v1"), backend.Conditions{})

		obj2 := &model.Object{Name: "versioned.txt", Bucket: "versioned-bucket"}
		b.CreateObject(ctx, "versioned-bucket", obj2, strings.NewReader("v2"), backend.Conditions{})

		if obj1.Generation == obj2.Generation {
			t.Errorf("Expected different generations, got %d", obj1.Generation)
		}

		resp, err := b.ListObjects(ctx, "versioned-bucket", backend.ListObjectsParams{Versions: true})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(resp.Items) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(resp.Items))
		}
	})

	t.Run("delete without generation preserves version", func(t *testing.T) {
		b.DeleteObject(ctx, "versioned-bucket", "versioned.txt", 0, backend.Conditions{})

		resp, err := b.ListObjects(ctx, "versioned-bucket", backend.ListObjectsParams{Versions: true})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(resp.Items) < 1 {
			t.Errorf("Expected at least 1 version after delete, got %d", len(resp.Items))
		}

		respLive, _ := b.ListObjects(ctx, "versioned-bucket", backend.ListObjectsParams{})
		if len(respLive.Items) != 0 {
			t.Errorf("Expected 0 live versions after delete, got %d", len(respLive.Items))
		}
	})

	t.Run("get specific generation", func(t *testing.T) {
		obj := &model.Object{Name: "gen-test.txt", Bucket: "versioned-bucket"}
		b.CreateObject(ctx, "versioned-bucket", obj, strings.NewReader("data"), backend.Conditions{})

		got, err := b.GetObject(ctx, "versioned-bucket", "gen-test.txt", obj.Generation)
		if err != nil {
			t.Fatalf("Failed to get object by generation: %v", err)
		}

		if got.Generation != obj.Generation {
			t.Errorf("Expected generation %d, got %d", obj.Generation, got.Generation)
		}
	})
}

func strPtr(s string) *string {
	return &s
}
