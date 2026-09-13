package services

import (
	"GarageSaleAPI/infrastructure/ratelimit"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClock is a hand-wound clock, so no test has to sleep.
type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time { return c.now }

func (c *testClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

const (
	testIP       = "192.0.2.10"
	testUsername = "Edgouille"
)

func newTestThrottle(t *testing.T, usernameLimit int, ipLimit int) (*LoginThrottleService, *testClock) {
	t.Helper()

	clock := &testClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	throttle := NewLoginThrottleService(
		ratelimit.NewInMemoryAttemptStore(ratelimit.DefaultMaxEntries),
		Policy{Limit: usernameLimit, Window: 15 * time.Minute},
		Policy{Limit: ipLimit, Window: 15 * time.Minute},
	).withClock(clock.Now)

	return throttle, clock
}

func TestLoginThrottle_AllowsAttemptsUnderTheLimit(t *testing.T) {
	throttle, _ := newTestThrottle(t, 10, 30)

	for i := 0; i < 9; i++ {
		_, allowed := throttle.Check(testIP, testUsername)
		require.True(t, allowed, "attempt %d should be allowed", i+1)
		throttle.RecordFailure(testIP, testUsername)
	}

	_, allowed := throttle.Check(testIP, testUsername)
	assert.True(t, allowed, "the 10th attempt should still be allowed")
}

func TestLoginThrottle_BlocksOnReachingTheLimit(t *testing.T) {
	throttle, _ := newTestThrottle(t, 10, 30)

	for i := 0; i < 10; i++ {
		throttle.RecordFailure(testIP, testUsername)
	}

	retryAfter, allowed := throttle.Check(testIP, testUsername)
	require.False(t, allowed, "the attempt after the limit should be blocked")
	assert.Equal(t, 1*time.Minute, retryAfter, "the first offense should block for a minute")
}

func TestLoginThrottle_RetryAfterShrinksAsTimePasses(t *testing.T) {
	throttle, clock := newTestThrottle(t, 10, 30)

	for i := 0; i < 10; i++ {
		throttle.RecordFailure(testIP, testUsername)
	}

	clock.Advance(20 * time.Second)

	retryAfter, allowed := throttle.Check(testIP, testUsername)
	require.False(t, allowed)
	assert.Equal(t, 40*time.Second, retryAfter)

	clock.Advance(40 * time.Second)

	_, allowed = throttle.Check(testIP, testUsername)
	assert.True(t, allowed, "the block should lapse once its duration has passed")
}

// The property the sliding log buys over a fixed window: failures expire one at
// a time, so an old failure ageing out frees exactly one slot.
func TestLoginThrottle_FailuresAgeOutIndividually(t *testing.T) {
	throttle, clock := newTestThrottle(t, 10, 30)

	// Nine failures spread across the window, one per minute.
	for i := 0; i < 9; i++ {
		throttle.RecordFailure(testIP, testUsername)
		clock.Advance(1 * time.Minute)
	}

	// Move past the window's start so only the oldest failure has lapsed.
	clock.Advance(6*time.Minute + 1*time.Second)

	// Eight failures still count, so this one lands under the limit.
	throttle.RecordFailure(testIP, testUsername)

	_, allowed := throttle.Check(testIP, testUsername)
	assert.True(t, allowed, "an aged-out failure should free one slot, not trip the limit")
}

func TestLoginThrottle_BackoffGrowsAndCaps(t *testing.T) {
	throttle, clock := newTestThrottle(t, 3, 1000)

	want := []time.Duration{
		1 * time.Minute,
		2 * time.Minute,
		4 * time.Minute,
		8 * time.Minute,
		30 * time.Minute,
		30 * time.Minute, // capped
	}

	for offense, wantBlock := range want {
		for i := 0; i < 3; i++ {
			throttle.RecordFailure(testIP, testUsername)
		}

		retryAfter, allowed := throttle.Check(testIP, testUsername)
		require.False(t, allowed, "offense %d should block", offense+1)
		assert.Equal(t, wantBlock, retryAfter, "offense %d", offense+1)

		clock.Advance(wantBlock)
	}
}

func TestLoginThrottle_SuccessClearsBothBudgets(t *testing.T) {
	throttle, _ := newTestThrottle(t, 10, 30)

	for i := 0; i < 9; i++ {
		throttle.RecordFailure(testIP, testUsername)
	}

	throttle.RecordSuccess(testIP, testUsername)

	// The budget is fresh: a full run of failures is needed to block again.
	for i := 0; i < 9; i++ {
		throttle.RecordFailure(testIP, testUsername)
		_, allowed := throttle.Check(testIP, testUsername)
		require.True(t, allowed, "failure %d after a success should not block", i+1)
	}
}

func TestLoginThrottle_SuccessResetsTheOffenseCount(t *testing.T) {
	throttle, clock := newTestThrottle(t, 3, 1000)

	// Two offenses: the next block would be 4 minutes.
	for offense := 0; offense < 2; offense++ {
		for i := 0; i < 3; i++ {
			throttle.RecordFailure(testIP, testUsername)
		}
		_, allowed := throttle.Check(testIP, testUsername)
		require.False(t, allowed)
		clock.Advance(10 * time.Minute)
	}

	throttle.RecordSuccess(testIP, testUsername)

	for i := 0; i < 3; i++ {
		throttle.RecordFailure(testIP, testUsername)
	}

	retryAfter, allowed := throttle.Check(testIP, testUsername)
	require.False(t, allowed)
	assert.Equal(t, 1*time.Minute, retryAfter, "a success should send the backoff back to the start")
}

func TestLoginThrottle_UsernameAndIPBudgetsAreIndependent(t *testing.T) {
	t.Run("blocking a username leaves other usernames on that IP alone", func(t *testing.T) {
		throttle, _ := newTestThrottle(t, 3, 1000)

		for i := 0; i < 3; i++ {
			throttle.RecordFailure(testIP, "victim")
		}

		_, allowed := throttle.Check(testIP, "victim")
		require.False(t, allowed)

		_, allowed = throttle.Check(testIP, "bystander")
		assert.True(t, allowed, "a different username from the same IP should still be allowed")
	})

	t.Run("blocking an IP leaves that username on other IPs alone", func(t *testing.T) {
		throttle, _ := newTestThrottle(t, 1000, 3)

		for i := 0; i < 3; i++ {
			throttle.RecordFailure("198.51.100.7", testUsername)
		}

		_, allowed := throttle.Check("198.51.100.7", testUsername)
		require.False(t, allowed)

		_, allowed = throttle.Check("203.0.113.9", testUsername)
		assert.True(t, allowed, "the same username from a different IP should still be allowed")
	})
}

func TestLoginThrottle_ReportsTheLongerWaitOfTheTwoBudgets(t *testing.T) {
	throttle, clock := newTestThrottle(t, 3, 5)

	// Trip the username budget first, then let some of its block elapse.
	for i := 0; i < 3; i++ {
		throttle.RecordFailure(testIP, testUsername)
	}
	clock.Advance(30 * time.Second)

	// Trip the IP budget from a different username, so its block runs longer.
	for i := 0; i < 5; i++ {
		throttle.RecordFailure(testIP, "another")
	}

	retryAfter, allowed := throttle.Check(testIP, testUsername)
	require.False(t, allowed)
	assert.Equal(t, 1*time.Minute, retryAfter, "the longer of the two remaining blocks should win")
}

func TestLoginThrottle_UnknownKeyIsAllowed(t *testing.T) {
	throttle, _ := newTestThrottle(t, 10, 30)

	retryAfter, allowed := throttle.Check("203.0.113.1", "nobody")

	assert.True(t, allowed)
	assert.Zero(t, retryAfter)
}
