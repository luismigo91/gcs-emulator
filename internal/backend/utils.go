package backend

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func GenerateID() string {
	return uuid.New().String()
}

func GenerateGeneration() int64 {
	return time.Now().UnixNano()
}

func IncrementMetageneration(meta int64) int64 {
	return meta + 1
}

func ParseGeneration(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func FormatGeneration(g int64) string {
	return strconv.FormatInt(g, 10)
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
