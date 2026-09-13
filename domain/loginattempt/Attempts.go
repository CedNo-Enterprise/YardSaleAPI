package loginattempt

import "time"

// Attempts is the throttling state held against one key. Unlike the other
// domain types it is a plain record with exported fields: it carries no
// invariants of its own, and only ever moves between the throttle and its store.
type Attempts struct {
	Failures     []time.Time // one entry per recent failure, newest last
	Offenses     int         // consecutive blocks, drives the backoff
	BlockedUntil time.Time
}
