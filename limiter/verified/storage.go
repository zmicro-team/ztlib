package verified

import (
	"context"
	"time"
)

// StoreArgs store arguments
type StoreArgs struct {
	DisableOneTime bool
	Key            string
	KeyExpires     time.Duration
	MaxErrQuota    int
	Answer         string
}

// VerifyArgs verify arguments
type VerifyArgs struct {
	DisableOneTime bool
	Key            string
	Answer         string
}

// Storage store engine
type Storage interface {
	Store(context.Context, *StoreArgs) error
	Verify(context.Context, *VerifyArgs) bool
}
