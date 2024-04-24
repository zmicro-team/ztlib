package limit

import (
	"context"
)

type PeriodFailureStorage interface {
	Check(ctx context.Context, key string, quota, expireSec int, success bool) (int64, error)
	SetQuotaFull(ctx context.Context, key string, quota, expireSec int) error
	Del(ctx context.Context, key string) error
	GetRunValue(ctx context.Context, key string) ([]int64, error)
}

type PeriodStorage interface {
	Take(ctx context.Context, key string, quota, expireSec int) (int64, error)
	SetQuotaFull(ctx context.Context, key string, quota, expireSec int) error
	Del(ctx context.Context, key string) error
	GetRunValue(ctx context.Context, key string) ([]int64, error)
}
