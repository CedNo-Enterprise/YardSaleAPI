package services

import (
	"GarageSaleAPI/domain/loginattempt"
	"time"
)

type Policy struct {
	Limit  int           // failures allowed inside Window
	Window time.Duration // how long a failure keeps counting
}

// DefaultBackoff is the block duration by offense number: the first block lasts
// a minute, and repeat offenders climb to the capped final entry.
var DefaultBackoff = []time.Duration{
	1 * time.Minute,
	2 * time.Minute,
	4 * time.Minute,
	8 * time.Minute,
	30 * time.Minute,
}

type LoginThrottleService struct {
	attempts    loginattempt.AttemptStore
	emailPolicy Policy
	ipPolicy    Policy
	backoff     []time.Duration
	now         func() time.Time
}

func NewLoginThrottleService(attempts loginattempt.AttemptStore, emailPolicy Policy, ipPolicy Policy) *LoginThrottleService {
	return &LoginThrottleService{
		attempts:    attempts,
		emailPolicy: emailPolicy,
		ipPolicy:    ipPolicy,
		backoff:     DefaultBackoff,
		now:         time.Now,
	}
}

// withClock lets tests drive time directly, so verifying window expiry and
// backoff growth costs no wall-clock waiting.
func (service *LoginThrottleService) withClock(now func() time.Time) *LoginThrottleService {
	service.now = now
	return service
}

func emailKey(email string) string { return "email:" + email }

func ipKey(ip string) string { return "ip:" + ip }

// Check reports whether a login attempt may proceed. When it may not, it also
// reports how long the caller should wait. It must be consulted before the
// password is verified: bcrypt costs hundreds of milliseconds, so letting a
// blocked attempt reach it would defeat the point.
func (service *LoginThrottleService) Check(ip string, email string) (time.Duration, bool) {
	now := service.now()

	var retryAfter time.Duration
	for _, key := range []string{emailKey(email), ipKey(ip)} {
		attempts, found := service.attempts.Get(key)
		if !found {
			continue
		}

		if attempts.BlockedUntil.After(now) {
			// Both keys are consulted so the caller is told the longer wait.
			if remaining := attempts.BlockedUntil.Sub(now); remaining > retryAfter {
				retryAfter = remaining
			}
		}
	}

	if retryAfter > 0 {
		return retryAfter, false
	}
	return 0, true
}

func (service *LoginThrottleService) RecordFailure(ip string, email string) {
	service.recordFailure(emailKey(email), service.emailPolicy)
	service.recordFailure(ipKey(ip), service.ipPolicy)
}

func (service *LoginThrottleService) recordFailure(key string, policy Policy) {
	now := service.now()

	attempts, _ := service.attempts.Get(key)
	attempts.Failures = append(pruneLapsed(attempts.Failures, now.Add(-policy.Window)), now)

	if len(attempts.Failures) >= policy.Limit {
		attempts.Offenses++
		attempts.BlockedUntil = now.Add(service.blockFor(attempts.Offenses))
		attempts.Failures = nil
	}

	service.attempts.Put(key, attempts)
}

// RecordSuccess clears both budgets. Proving you know the password retires the
// offense count with them, so a user's occasional typos never accumulate.
func (service *LoginThrottleService) RecordSuccess(ip string, email string) {
	service.attempts.Delete(emailKey(email))
	service.attempts.Delete(ipKey(ip))
}

func (service *LoginThrottleService) blockFor(offense int) time.Duration {
	if offense < 1 {
		offense = 1
	}
	if offense > len(service.backoff) {
		offense = len(service.backoff)
	}
	return service.backoff[offense-1]
}

// pruneLapsed drops failures that have aged out of the window. Each one expires
// on its own schedule, so there is no boundary for an attacker to time guesses
// around and no lump reset of a legitimate user's count.
func pruneLapsed(failures []time.Time, cutoff time.Time) []time.Time {
	kept := failures[:0]
	for _, failure := range failures {
		if failure.After(cutoff) {
			kept = append(kept, failure)
		}
	}
	return kept
}
