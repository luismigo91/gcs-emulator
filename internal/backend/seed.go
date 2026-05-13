package backend

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
)

func SeedFromDirectory(b Backend, seedPath string) error {
	if seedPath == "" {
		return nil
	}

	info, err := os.Stat(seedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	ctx := context.Background()
	entries, err := os.ReadDir(seedPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		bucketName := entry.Name()
		bucket := &model.Bucket{
			Name:          bucketName,
			Kind:          "storage#bucket",
			Metageneration: 1,
			TimeCreated:   time.Now(),
			Updated:       time.Now(),
		}

		if err := b.CreateBucket(ctx, bucket, Conditions{}); err != nil {
			return err
		}

		bucketPath := filepath.Join(seedPath, bucketName)
		if err := seedBucket(ctx, b, bucketPath, bucketName); err != nil {
			return err
		}
	}

	return nil
}

func seedBucket(ctx context.Context, b Backend, bucketPath, bucketName string) error {
	return filepath.Walk(bucketPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(bucketPath, path)
		if err != nil {
			return err
		}

		objectName := filepath.ToSlash(relPath)

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		contentType := detectContentTypeFromData(path, data)

		obj := &model.Object{
			Name:        objectName,
			Bucket:      bucketName,
			Kind:        "storage#object",
			ContentType: contentType,
		}

		return b.CreateObject(ctx, bucketName, obj, bytes.NewReader(data), Conditions{})
	})
}

func detectContentTypeFromData(path string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".log":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".gz":
		return "application/gzip"
	case ".tar":
		return "application/x-tar"
	case ".yaml", ".yml":
		return "application/x-yaml"
	case ".toml":
		return "application/toml"
	case ".md":
		return "text/markdown"
	case ".csv":
		return "text/csv"
	}

	if len(data) > 0 {
		return http.DetectContentType(data[:min(len(data), 512)])
	}

	return "application/octet-stream"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
