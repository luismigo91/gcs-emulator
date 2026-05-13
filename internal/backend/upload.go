package backend

import (
	"sort"
	"sync"
	"time"
)

type UploadSession struct {
	ID         string
	Bucket     string
	ObjectName string
	Metadata   map[string]interface{}
	Chunks     map[int64][]byte
	TotalSize  int64
	Offset     int64
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type UploadSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*UploadSession
}

var DefaultSessionStore = NewUploadSessionStore()

func NewUploadSessionStore() *UploadSessionStore {
	s := &UploadSessionStore{
		sessions: make(map[string]*UploadSession),
	}
	go s.cleanupLoop()
	return s
}

func (s *UploadSessionStore) Create(session *UploadSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

func (s *UploadSessionStore) Get(id string) (*UploadSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	if !exists {
		return nil, false
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

func (s *UploadSessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *UploadSessionStore) StoreChunk(id string, offset int64, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[id]
	if !exists {
		return ErrUploadSessionNotFound
	}

	session.Chunks[offset] = data
	session.Offset = offset + int64(len(data))
	return nil
}

func (s *UploadSessionStore) MergeChunks(id string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil
	}

	offsets := make([]int64, 0, len(session.Chunks))
	for offset := range session.Chunks {
		offsets = append(offsets, offset)
	}
	sort.Slice(offsets, func(i, j int) bool {
		return offsets[i] < offsets[j]
	})

	var result []byte
	for _, offset := range offsets {
		result = append(result, session.Chunks[offset]...)
	}

	return result
}

func (s *UploadSessionStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

func (s *UploadSessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}
