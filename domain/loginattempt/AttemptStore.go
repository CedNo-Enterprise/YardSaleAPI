package loginattempt

import "time"

type AttemptStore interface {
	Get(key string) (Attempts, bool)
	Put(key string, attempts Attempts)
	Delete(key string)
	DeleteStale(before time.Time)
	Len() int
}
