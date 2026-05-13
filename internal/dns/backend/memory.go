package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/dns/model"
)

var ErrZoneNotFound = errors.New("zone not found")

type MemoryDNSBackend struct {
	mu     sync.RWMutex
	zones  map[string]*model.ManagedZone
	rrsets map[string][]*model.ResourceRecordSet
}

func NewMemoryDNSBackend() *MemoryDNSBackend {
	return &MemoryDNSBackend{
		zones:  make(map[string]*model.ManagedZone),
		rrsets: make(map[string][]*model.ResourceRecordSet),
	}
}

func (m *MemoryDNSBackend) CreateZone(ctx context.Context, project string, zone *model.ManagedZone) (*model.ManagedZone, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := project + "/" + zone.Name
	if _, exists := m.zones[key]; exists {
		return nil, errors.New("zone already exists")
	}
	zone.NameServers = []string{"ns-cloud-e1.emulator.local.", "ns-cloud-e2.emulator.local."}
	m.zones[key] = zone
	return zone, nil
}

func (m *MemoryDNSBackend) GetZone(ctx context.Context, project, name string) (*model.ManagedZone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	z, exists := m.zones[project+"/"+name]
	if !exists {
		return nil, ErrZoneNotFound
	}
	return z, nil
}

func (m *MemoryDNSBackend) ListZones(ctx context.Context, project string) ([]*model.ManagedZone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.ManagedZone
	for k, z := range m.zones {
		if strings.HasPrefix(k, project+"/") {
			result = append(result, z)
		}
	}
	return result, nil
}

func (m *MemoryDNSBackend) DeleteZone(ctx context.Context, project, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := project + "/" + name
	if _, exists := m.zones[key]; !exists {
		return ErrZoneNotFound
	}
	delete(m.zones, key)
	delete(m.rrsets, key)
	return nil
}

func (m *MemoryDNSBackend) CreateChange(ctx context.Context, project, zone string, change *model.Change) ([]*model.ResourceRecordSet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := project + "/" + zone
	if _, exists := m.zones[key]; !exists {
		return nil, ErrZoneNotFound
	}
	for _, a := range change.Additions {
		m.rrsets[key] = append(m.rrsets[key], a)
	}
	for _, d := range change.Deletions {
		list := m.rrsets[key]
		for i, r := range list {
			if r.Name == d.Name && r.Type == d.Type {
				m.rrsets[key] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
	return m.rrsets[key], nil
}

func (m *MemoryDNSBackend) ListRecordSets(ctx context.Context, project, zone string) ([]*model.ResourceRecordSet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := project + "/" + zone
	return m.rrsets[key], nil
}

func (m *MemoryDNSBackend) Shutdown() error { return nil }
