package backend

import (
	"context"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/asset/model"
)

type MemoryAssetBackend struct {
	mu     sync.RWMutex
	assets []*model.Asset
}

func NewMemoryAssetBackend() *MemoryAssetBackend { return &MemoryAssetBackend{} }

func (m *MemoryAssetBackend) Export(ctx context.Context, types []string) ([]*model.Asset, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	if len(types) == 0 { return m.assets, nil }
	var result []*model.Asset
	for _, a := range m.assets {
		for _, t := range types {
			if a.AssetType == t { result = append(result, a); break }
		}
	}
	return result, nil
}

func (m *MemoryAssetBackend) Add(ctx context.Context, a *model.Asset) {
	m.mu.Lock(); defer m.mu.Unlock()
	m.assets = append(m.assets, a)
}

func (m *MemoryAssetBackend) Shutdown() error { return nil }
