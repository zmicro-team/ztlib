package limit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type anotherPeriodFailureLimitDriver struct{}

func (anotherPeriodFailureLimitDriver) CheckErr(ctx context.Context, key string, err error) (PeriodFailureLimitState, error) {
	return PeriodFailureLimitStsUnknown, ErrUnsupportedDriver
}
func (anotherPeriodFailureLimitDriver) Check(context.Context, string, bool) (PeriodFailureLimitState, error) {
	return PeriodFailureLimitStsUnknown, ErrUnsupportedDriver
}
func (anotherPeriodFailureLimitDriver) SetQuotaFull(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (anotherPeriodFailureLimitDriver) Del(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u anotherPeriodFailureLimitDriver) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	return nil, ErrUnsupportedDriver
}

func TestPeriodFailureManager(t *testing.T) {
	var unsupported = "unsupported"
	var another = "another_driver"
	var anotherDriver = new(anotherPeriodFailureLimitDriver)

	m := NewPeriodFailureLimitManagerWithDriver(map[string]PeriodFailureLimitDriver{
		unsupported: unsupportedPeriodFailureLimitKindDriver,
	})
	err := m.Register(another, anotherDriver)
	require.Nil(t, err)

	err = m.Register(another, anotherDriver)
	require.ErrorIs(t, err, ErrDuplicateDriver)

	d := m.Acquire(another)
	require.Equal(t, d, anotherDriver)

	d = m.Acquire("not found")
	require.Equal(t, d, unsupportedPeriodFailureLimitKindDriver)
}
