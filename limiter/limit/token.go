package limit

import (
	"time"
)

type LimitToken interface {
	AllowN(now time.Time, n int) bool
	Allow() bool
}
