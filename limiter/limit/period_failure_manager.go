package limit

import (
	"context"
	"sync"
)

var unsupportedPeriodFailureLimitKindDriver = new(UnsupportedPeriodFailureLimitDriver)

// PeriodFailureLimitState period failure limit state.
type PeriodFailureLimitState int

const (
	// PeriodFailureLimitStsUnknown means not initialized state.
	PeriodFailureLimitStsUnknown PeriodFailureLimitState = iota - 1
	// PeriodFailureLimitStsSuccess means success.
	PeriodFailureLimitStsSuccess
	// PeriodFailureLimitStsInQuota means within the quota.
	PeriodFailureLimitStsInQuota
	// PeriodFailureLimitStsOverQuota means over the quota.
	PeriodFailureLimitStsOverQuota
)

// IsSuccess means success state.
func (p PeriodFailureLimitState) IsSuccess() bool { return p == PeriodFailureLimitStsSuccess }

// IsWithinQuota means within the quota.
func (p PeriodFailureLimitState) IsWithinQuota() bool { return p == PeriodFailureLimitStsInQuota }

// IsOverQuota means passed the quota.
func (p PeriodFailureLimitState) IsOverQuota() bool { return p == PeriodFailureLimitStsOverQuota }

// PeriodFailureLimitDriver driver interface
type PeriodFailureLimitDriver interface {
	// CheckErr requests a permit state.
	// same as Check
	CheckErr(ctx context.Context, key string, err error) (PeriodFailureLimitState, error)
	// Check requests a permit.
	Check(ctx context.Context, key string, success bool) (PeriodFailureLimitState, error)
	// SetQuotaFull set a permit over quota.
	SetQuotaFull(ctx context.Context, key string) error
	// Del delete a permit
	Del(ctx context.Context, key string) error
	// GetRunValue get run value
	// Exist: false if key not exist.
	// Count: current failure count
	// TTL: not set expire time, t = -1.
	GetRunValue(ctx context.Context, key string) (*RunValue, error)
}

// PeriodFailureLimitManager manage limit period failure
type PeriodFailureLimitManager[T comparable] struct {
	mu     sync.RWMutex
	driver map[T]PeriodFailureLimitDriver
}

// NewPeriodFailureLimitManager new a instance
func NewPeriodFailureLimitManager[T comparable]() *PeriodFailureLimitManager[T] {
	return &PeriodFailureLimitManager[T]{
		driver: map[T]PeriodFailureLimitDriver{},
	}
}

// NewPeriodFailureLimitManagerWithDriver new a instance with driver
func NewPeriodFailureLimitManagerWithDriver[T comparable](drivers map[T]PeriodFailureLimitDriver) *PeriodFailureLimitManager[T] {
	p := NewPeriodFailureLimitManager[T]()
	for kind, drive := range drivers {
		p.driver[kind] = drive
	}
	return p
}

// PeriodFailureLimitManager register a PeriodFailureLimitDriver with kind.
func (p *PeriodFailureLimitManager[T]) Register(kind T, d PeriodFailureLimitDriver) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.driver[kind]
	if ok {
		return ErrDuplicateDriver
	}
	p.driver[kind] = d
	return nil
}

// Acquire driver. if driver not exist. it will return UnsupportedPeriodFailureLimitDriver.
func (p *PeriodFailureLimitManager[T]) Acquire(kind T) PeriodFailureLimitDriver {
	p.mu.RLock()
	defer p.mu.RUnlock()
	d, ok := p.driver[kind]
	if ok {
		return d
	}
	return unsupportedPeriodFailureLimitKindDriver
}

// UnsupportedPeriodFailureLimitDriver unsupported limit period failure driver
type UnsupportedPeriodFailureLimitDriver struct{}

func (UnsupportedPeriodFailureLimitDriver) CheckErr(ctx context.Context, key string, err error) (PeriodFailureLimitState, error) {
	return PeriodFailureLimitStsUnknown, ErrUnsupportedDriver
}
func (UnsupportedPeriodFailureLimitDriver) Check(context.Context, string, bool) (PeriodFailureLimitState, error) {
	return PeriodFailureLimitStsUnknown, ErrUnsupportedDriver
}
func (UnsupportedPeriodFailureLimitDriver) SetQuotaFull(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (UnsupportedPeriodFailureLimitDriver) Del(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u UnsupportedPeriodFailureLimitDriver) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	return nil, ErrUnsupportedDriver
}
