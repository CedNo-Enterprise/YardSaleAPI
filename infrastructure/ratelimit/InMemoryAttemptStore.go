package ratelimit

import (
	"GarageSaleAPI/domain/loginattempt"
	"sync"
	"time"
)

// DefaultMaxEntries bounds the map. Keys come from attacker-supplied input, so
// without a ceiling a spray across distinct usernames would grow it without
// limit.
const DefaultMaxEntries = 100_000

type InMemoryAttemptStore struct {
	mu         sync.Mutex
	entries    map[string]loginattempt.Attempts
	maxEntries int
	dropped    int
}

func NewInMemoryAttemptStore(maxEntries int) *InMemoryAttemptStore {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxEntries
	}
	return &InMemoryAttemptStore{
		entries:    make(map[string]loginattempt.Attempts),
		maxEntries: maxEntries,
	}
}

func (s *InMemoryAttemptStore) Get(key string) (loginattempt.Attempts, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attempts, found := s.entries[key]
	if !found {
		return loginattempt.Attempts{}, false
	}

	// Hand back a copy: the caller must not be able to mutate what the map holds.
	return copyAttempts(attempts), true
}

func (s *InMemoryAttemptStore) Put(key string, attempts loginattempt.Attempts) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Once full, stop tracking keys we have not seen before. Existing keys keep
	// updating, so a budget already in progress is never forgotten halfway.
	if _, found := s.entries[key]; !found && len(s.entries) >= s.maxEntries {
		s.dropped++
		return
	}

	s.entries[key] = copyAttempts(attempts)
}

func (s *InMemoryAttemptStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
}

// DeleteStale drops entries whose failure log and block have both lapsed.
func (s *InMemoryAttemptStore) DeleteStale(before time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, attempts := range s.entries {
		if attempts.BlockedUntil.After(before) {
			continue
		}
		if len(attempts.Failures) > 0 && attempts.Failures[len(attempts.Failures)-1].After(before) {
			continue
		}
		delete(s.entries, key)
	}
}

func (s *InMemoryAttemptStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.entries)
}

// Dropped reports how many writes were refused because the store was full.
func (s *InMemoryAttemptStore) Dropped() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.dropped
}

func copyAttempts(attempts loginattempt.Attempts) loginattempt.Attempts {
	if attempts.Failures == nil {
		return attempts
	}

	failures := make([]time.Time, len(attempts.Failures))
	copy(failures, attempts.Failures)
	attempts.Failures = failures

	return attempts
}
