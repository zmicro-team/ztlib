package limit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type anotherPeriodLimitDriver struct{}

func (u anotherPeriodLimitDriver) Take(context.Context, string) (PeriodLimitState, error) {
	return PeriodLimitStsUnknown, ErrUnsupportedDriver
}
func (u anotherPeriodLimitDriver) SetQuotaFull(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u anotherPeriodLimitDriver) Del(context.Context, string) error {
	return ErrUnsupportedDriver
}
func (u anotherPeriodLimitDriver) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	return nil, ErrUnsupportedDriver
}

func TestPeriodManager(t *testing.T) {
	var unsupported = "unsupported"
	var another = "another_driver"
	var anotherDriver = new(anotherPeriodLimitDriver)

	m := NewPeriodLimitManagerWithDriver(map[string]PeriodLimitDriver{
		unsupported: unsupportedPeriodLimitKindDriver,
	})
	err := m.Register(another, anotherDriver)
	require.Nil(t, err)

	err = m.Register(another, anotherDriver)
	require.ErrorIs(t, err, ErrDuplicateDriver)

	d := m.Acquire(another)
	require.Equal(t, d, anotherDriver)

	d = m.Acquire("not found")
	require.Equal(t, d, unsupportedPeriodLimitKindDriver)
}
