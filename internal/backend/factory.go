package backend

import (
	"fmt"
	"time"
)

type BackendMode string

const (
	ModeMemory     BackendMode = "memory"
	ModePersistent BackendMode = "persistent"
	ModeHybrid     BackendMode = "hybrid"
	ModeWAL        BackendMode = "wal"
)

type BackendConfig struct {
	Mode          BackendMode
	StoragePath   string
	FlushInterval time.Duration
}

func NewBackend(config BackendConfig) (Backend, error) {
	switch config.Mode {
	case ModeMemory:
		return NewMemoryBackend(), nil

	case ModePersistent:
		return NewPersistentBackend(config.StoragePath)

	case ModeHybrid:
		interval := config.FlushInterval
		if interval == 0 {
			interval = 5 * time.Second
		}
		return NewHybridBackend(config.StoragePath, interval)

	case ModeWAL:
		return NewWALBackend(config.StoragePath)

	default:
		return nil, fmt.Errorf("unknown backend mode: %s", config.Mode)
	}
}
