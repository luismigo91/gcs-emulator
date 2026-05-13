package backend

import (
	"fmt"
	"time"
)

func NewPubSubBackend(mode, storagePath string) (PubSubBackend, error) {
	switch mode {
	case "memory", "":
		return NewMemoryPubSubBackend(), nil
	case "persistent":
		return NewPersistentPubSubBackend(storagePath)
	case "hybrid":
		return NewHybridPubSubBackend(storagePath, 5*time.Second)
	case "wal":
		return NewWALPubSubBackend(storagePath)
	default:
		return nil, fmt.Errorf("unknown pubsub backend mode: %s", mode)
	}
}
