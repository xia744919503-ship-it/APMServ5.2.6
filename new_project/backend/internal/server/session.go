package server

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const sessionCookieName = "rxsg_session"

type sessionRecord struct {
	UID       int
	CreatedAt time.Time
}

type sessionStore struct {
	mu    sync.RWMutex
	items map[string]sessionRecord
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		items: make(map[string]sessionRecord),
	}
}

func (s *sessionStore) create(uid int) (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	token := hex.EncodeToString(buffer)
	s.mu.Lock()
	s.items[token] = sessionRecord{
		UID:       uid,
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()
	return token, nil
}

func (s *sessionStore) get(token string) (sessionRecord, bool) {
	s.mu.RLock()
	item, ok := s.items[token]
	s.mu.RUnlock()
	return item, ok
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	delete(s.items, token)
	s.mu.Unlock()
}
