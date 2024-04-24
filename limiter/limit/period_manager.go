package limit

import (
	"context"
	"sync"
)

var unsupportedPeriodLimitKindDriver = new(UnsupportedPeriodLimitDriver)

// PeriodLimitState period limit state.
type PeriodLimitState int

const (
	// PeriodLimitStsUnknown means not initialized state.
	PeriodLimitStsUnknown PeriodLimitState = iota - 1
	// PeriodLimitStsAllowed means allowed.
	PeriodLimitStsAllowed
	// PeriodLimitStsHitQuota means hit the quota.
	PeriodLimitStsHitQuota
	// PeriodLimitStsOverQuota means passed the quota.
	PeriodLimitStsOverQuota
)

// IsAllowed means allowed state.
func (p PeriodLimitState) IsAllowed() bool { return p == PeriodLimitStsAllowed }

// IsHitQuota means this request exactly hit the quota.
func (p PeriodLimitState) IsHitQuota() bool { return p == PeriodLimitStsHitQuota }

// IsOverQuota means passed the quota.
func (p PeriodLimitState) IsOverQuota() bool { return p == PeriodLimitStsOverQuota }

// PeriodLimitDriver driver interface
type PeriodLimitDriver interface {
	// Take requests a permit with context, it returns the permit state.
	Take(ctx context.Context, key string) (PeriodLimitState, error)
	// SetQuotaFull set a permit over quota.
	SetQuotaFull(ctx context.Context, key string) error
	// Del delete a permit
	Del(ctx context.Context, key string) error
	// GetRunValue get run value
	// Exist: false if key not exist.
	// Count: current count
	// TTL: not set expire time, t = -1.
	GetRunValue(ctx context.Context, key string) (*RunValue, error)
}

// PeriodLimitManager manage limit period
type PeriodLimitManager[T comparable] struct {
	mu     sync.RWMutex
	driver map[T]PeriodLimitDriver
}

// NewPeriodLimitManager new a instance
func NewPeriodLimitManager[T comparable]() *PeriodLimitManager[T] {
	return &PeriodLimitManager[T]{
		driver: map[T]PeriodLimitDriver{},
	}
}

// NewPeriodLimitManagerWithDriver new a instance with driver
func NewPeriodLimitManagerWithDriver[T comparable](drivers map[T]PeriodLimitDriver) *PeriodLimitManager[T] {
	p := NewPeriodLimitManager[T]()
	for kind, drive := range drivers {
		p.driver[kind] = drive
	}
	return p
}

// Register register a PeriodLimitDriver with kind
func (p *PeriodLimitManager[T]) Register(kind T, d PeriodLimitDriver) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.driver[kind]
	if ok {
		return ErrDuplicateDriver
	}
	p.driver[kind] = d
	return nil
}

// Acquire driver. if driver not exist. it will return UnsupportedPeriodLimitDriver.
func (p *PeriodLimitManager[T]) Acquire(kind T) PeriodLimitDriver {
	p.mu.RLock()
	defer p.mu.RUnlock()
	d, ok := p.driver[kind]
	if ok {
		return d
	}
	return unsupportedPeriodLimitKindDriver
}

// UnsupportedPeriodLimitDriver unsupported limit period driver
type UnsupportedPeriodLimitDriver struct{}

func (u UnsupportedPeriodLimitDriver) Take(context.Context, string) (PeriodLimitState, error) {
	return PeriodLimitStsUnknown, ErrUnsupportedDriver
}
func (u UnsupportedPeriodLimitDriver) SetQuotaFull(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u UnsupportedPeriodLimitDriver) Del(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u UnsupportedPeriodLimitDriver) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	return nil, ErrUnsupportedDriver
}
