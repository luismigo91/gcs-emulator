package backend

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/kms/model"
)

var (
	ErrKeyRingNotFound      = errors.New("key ring not found")
	ErrCryptoKeyNotFound    = errors.New("crypto key not found")
	ErrVersionNotFound      = errors.New("version not found")
)

type MemoryKMSBackend struct {
	mu        sync.RWMutex
	keyRings  map[string]*model.KeyRing
	cryptoKeys map[string]*model.CryptoKey
	versions  map[string][]*model.CryptoKeyVersion
}

func NewMemoryKMSBackend() *MemoryKMSBackend {
	return &MemoryKMSBackend{
		keyRings:   make(map[string]*model.KeyRing),
		cryptoKeys: make(map[string]*model.CryptoKey),
		versions:   make(map[string][]*model.CryptoKeyVersion),
	}
}

func (m *MemoryKMSBackend) CreateKeyRing(ctx context.Context, project, location, name string) (*model.KeyRing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := "projects/" + project + "/locations/" + location + "/keyRings/" + name
	if _, exists := m.keyRings[n]; exists {
		return nil, errors.New("key ring already exists")
	}
	kr := &model.KeyRing{Name: n, CreateTime: time.Now()}
	m.keyRings[n] = kr
	return kr, nil
}

func (m *MemoryKMSBackend) GetKeyRing(ctx context.Context, name string) (*model.KeyRing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	kr, exists := m.keyRings[name]
	if !exists {
		return nil, ErrKeyRingNotFound
	}
	return kr, nil
}

func (m *MemoryKMSBackend) ListKeyRings(ctx context.Context, project, location string) ([]*model.KeyRing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefix := "projects/" + project + "/locations/" + location + "/"
	var result []*model.KeyRing
	for n, kr := range m.keyRings {
		if len(n) > len(prefix) && n[:len(prefix)] == prefix {
			result = append(result, kr)
		}
	}
	return result, nil
}

func (m *MemoryKMSBackend) DeleteKeyRing(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.keyRings[name]; !exists {
		return ErrKeyRingNotFound
	}
	delete(m.keyRings, name)
	return nil
}

func (m *MemoryKMSBackend) CreateCryptoKey(ctx context.Context, parent string, key *model.CryptoKey) (*model.CryptoKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if key.Name == "" {
		key.Name = parent + "/cryptoKeys/" + key.Purpose
	}
	if key.Purpose == "" {
		key.Purpose = "ENCRYPT_DECRYPT"
	}
	key.CreateTime = time.Now()
	m.cryptoKeys[key.Name] = key
	m.versions[key.Name] = make([]*model.CryptoKeyVersion, 0)
	return key, nil
}

func (m *MemoryKMSBackend) GetCryptoKey(ctx context.Context, name string) (*model.CryptoKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key, exists := m.cryptoKeys[name]
	if !exists {
		return nil, ErrCryptoKeyNotFound
	}
	return key, nil
}

func (m *MemoryKMSBackend) ListCryptoKeys(ctx context.Context, parent string) ([]*model.CryptoKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.CryptoKey
	for n, k := range m.cryptoKeys {
		if len(n) > len(parent) && n[:len(parent)] == parent {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *MemoryKMSBackend) UpdateCryptoKey(ctx context.Context, name string, key *model.CryptoKey) (*model.CryptoKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, exists := m.cryptoKeys[name]
	if !exists {
		return nil, ErrCryptoKeyNotFound
	}
	if key.NextRotationTime != nil {
		existing.NextRotationTime = key.NextRotationTime
	}
	if key.RotationPeriod != nil {
		existing.RotationPeriod = key.RotationPeriod
	}
	if key.Labels != nil {
		existing.Labels = key.Labels
	}
	return existing, nil
}

func (m *MemoryKMSBackend) CreateCryptoKeyVersion(ctx context.Context, parent string) (*model.CryptoKeyVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.cryptoKeys[parent]; !exists {
		return nil, ErrCryptoKeyNotFound
	}
	v := &model.CryptoKeyVersion{
		Name:       parent + "/cryptoKeyVersions/" + fmt.Sprintf("%d", len(m.versions[parent])+1),
		State:      "ENABLED",
		Algorithm:  "GOOGLE_SYMMETRIC_ENCRYPTION",
		CreateTime: time.Now(),
	}
	m.versions[parent] = append(m.versions[parent], v)
	return v, nil
}

func (m *MemoryKMSBackend) GetCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, vs := range m.versions {
		for _, v := range vs {
			if v.Name == name {
				return v, nil
			}
		}
	}
	return nil, ErrVersionNotFound
}

func (m *MemoryKMSBackend) ListCryptoKeyVersions(ctx context.Context, parent string) ([]*model.CryptoKeyVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.versions[parent], nil
}

func (m *MemoryKMSBackend) DestroyCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, err := m.findVersionLocked(name)
	if err != nil {
		return nil, err
	}
	v.State = "DESTROYED"
	v.DestroyTime = time.Now()
	return v, nil
}

func (m *MemoryKMSBackend) RestoreCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, err := m.findVersionLocked(name)
	if err != nil {
		return nil, err
	}
	v.State = "DISABLED"
	return v, nil
}

func (m *MemoryKMSBackend) findVersionLocked(name string) (*model.CryptoKeyVersion, error) {
	for _, vs := range m.versions {
		for _, v := range vs {
			if v.Name == name {
				return v, nil
			}
		}
	}
	return nil, ErrVersionNotFound
}

func (m *MemoryKMSBackend) Encrypt(ctx context.Context, name string, plaintext []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.cryptoKeys[name]; !exists {
		return nil, ErrCryptoKeyNotFound
	}
	return append([]byte("ENCRYPTED:"), plaintext...), nil
}

func (m *MemoryKMSBackend) Decrypt(ctx context.Context, name string, ciphertext []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.cryptoKeys[name]; !exists {
		return nil, ErrCryptoKeyNotFound
	}
	if len(ciphertext) > 10 && string(ciphertext[:10]) == "ENCRYPTED:" {
		return ciphertext[10:], nil
	}
	return ciphertext, nil
}

func (m *MemoryKMSBackend) Shutdown() error { return nil }
