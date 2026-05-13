package util

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
)

type contextKey string

const ProjectIDKey contextKey = "projectID"

func GetProjectID(r *http.Request) string {
	if projectID, ok := r.Context().Value(ProjectIDKey).(string); ok {
		return projectID
	}
	return ""
}

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, ProjectIDKey, projectID)
}

func ParseInt64Param(param string, name string) (int64, error) {
	if param == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, param)
	}
	return v, nil
}

func GenerateProjectNumber(projectID string) uint64 {
	if projectID == "" {
		return 0
	}
	hash := md5.Sum([]byte(projectID))
	return uint64(hash[0])<<56 | uint64(hash[1])<<48 | uint64(hash[2])<<40 | uint64(hash[3])<<32 |
		uint64(hash[4])<<24 | uint64(hash[5])<<16 | uint64(hash[6])<<8 | uint64(hash[7])
}
