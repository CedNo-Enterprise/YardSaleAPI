package ratelimit

import (
	"GarageSaleAPI/domain/loginattempt"
	"fmt"
	"sync"
	"testing"
	"time"
)

var testTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestInMemoryAttemptStore_PutGetDelete(t *testing.T) {
	store := NewInMemoryAttemptStore(DefaultMaxEntries)

	if _, found := store.Get("user:absent"); found {
		t.Error("expected an unknown key to be absent")
	}

	want := loginattempt.Attempts{
		Failures:     []time.Time{testTime},
		Offenses:     2,
		BlockedUntil: testTime.Add(time.Minute),
	}
	store.Put("user:someone", want)

	got, found := store.Get("user:someone")
	if !found {
		t.Fatal("expected the key to be present")
	}
	if got.Offenses != want.Offenses || !got.BlockedUntil.Equal(want.BlockedUntil) || len(got.Failures) != 1 {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}

	store.Delete("user:someone")
	if _, found := store.Get("user:someone"); found {
		t.Error("expected the key to be gone after Delete")
	}
}

// The store must hand out copies; otherwise a caller could mutate the slice the
// map is holding, under no lock at all.
func TestInMemoryAttemptStore_DoesNotAliasStoredSlices(t *testing.T) {
	store := NewInMemoryAttemptStore(DefaultMaxEntries)

	original := []time.Time{testTime, testTime.Add(time.Minute)}
	store.Put("user:someone", loginattempt.Attempts{Failures: original})

	// Mutating what we passed in must not reach the store.
	original[0] = testTime.Add(99 * time.Hour)

	stored, _ := store.Get("user:someone")
	if !stored.Failures[0].Equal(testTime) {
		t.Error("the store aliased the slice it was given")
	}

	// Mutating what we got back must not reach the store either.
	stored.Failures[0] = testTime.Add(99 * time.Hour)

	again, _ := store.Get("user:someone")
	if !again.Failures[0].Equal(testTime) {
		t.Error("the store handed out a slice aliased into the map")
	}
}

func TestInMemoryAttemptStore_DeleteStale(t *testing.T) {
	store := NewInMemoryAttemptStore(DefaultMaxEntries)

	store.Put("user:blocked", loginattempt.Attempts{BlockedUntil: testTime.Add(10 * time.Minute)})
	store.Put("user:recent", loginattempt.Attempts{Failures: []time.Time{testTime.Add(5 * time.Minute)}})
	store.Put("user:lapsed", loginattempt.Attempts{
		Failures:     []time.Time{testTime.Add(-time.Hour)},
		BlockedUntil: testTime.Add(-30 * time.Minute),
	})
	store.Put("user:empty", loginattempt.Attempts{})

	store.DeleteStale(testTime)

	for _, key := range []string{"user:blocked", "user:recent"} {
		if _, found := store.Get(key); !found {
			t.Errorf("%s should have been kept", key)
		}
	}
	for _, key := range []string{"user:lapsed", "user:empty"} {
		if _, found := store.Get(key); found {
			t.Errorf("%s should have been swept", key)
		}
	}
}

func TestInMemoryAttemptStore_EnforcesTheEntryCap(t *testing.T) {
	store := NewInMemoryAttemptStore(3)

	for i := 0; i < 10; i++ {
		store.Put(fmt.Sprintf("user:%d", i), loginattempt.Attempts{Offenses: 1})
	}

	if store.Len() != 3 {
		t.Errorf("Len() = %d, want 3", store.Len())
	}
	if store.Dropped() != 7 {
		t.Errorf("Dropped() = %d, want 7", store.Dropped())
	}

	// A key already being tracked must keep updating even when the store is full.
	store.Put("user:0", loginattempt.Attempts{Offenses: 5})
	got, found := store.Get("user:0")
	if !found || got.Offenses != 5 {
		t.Errorf("an existing key should still update when full, got %+v (found=%v)", got, found)
	}
}

// Run under -race: this is the project's first concurrent state.
func TestInMemoryAttemptStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryAttemptStore(DefaultMaxEntries)

	const workers = 8
	const iterations = 200

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			key := fmt.Sprintf("user:%d", w%3) // deliberate contention on shared keys
			for i := 0; i < iterations; i++ {
				attempts, _ := store.Get(key)
				attempts.Failures = append(attempts.Failures, testTime)
				attempts.Offenses++
				store.Put(key, attempts)
				store.DeleteStale(testTime.Add(-time.Hour))
				store.Len()
			}
		}(w)
	}
	wg.Wait()
}
