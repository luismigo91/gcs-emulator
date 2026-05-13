package util

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"net/http"
	"path/filepath"
	"strings"
)

var crc32cTable = crc32.MakeTable(0x82F63B78)

func CalculateCRC32C(data []byte) string {
	crc := crc32.Checksum(data, crc32cTable)
	return base64.StdEncoding.EncodeToString([]byte{
		byte(crc >> 24),
		byte(crc >> 16),
		byte(crc >> 8),
		byte(crc),
	})
}

func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateETag(generation int64) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%d", generation)))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func DetectContentTypeFromName(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
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
	return "application/octet-stream"
}

func DetectContentType(data []byte) string {
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
