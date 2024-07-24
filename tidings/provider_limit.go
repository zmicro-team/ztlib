package tidings

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/zmicro-team/ztlib/limiter/limit_verified"
	limitVerifiedRedisV9 "github.com/zmicro-team/ztlib/limiter/limit_verified/redis/v9"
)

var _ Provider = (*Limit)(nil)

type LimitDriver interface {
	limit_verified.LimitVerifiedProvider
	Send(ctx context.Context, c *NoticeParam) error
}

type Limit struct {
	limit  *limit_verified.LimitVerified[LimitDriver, *limitVerifiedRedisV9.RedisStore]
	driver LimitDriver
}

type LimitDependency struct {
	Redisc  *redis.Client
	Driver  LimitDriver
	Options []limit_verified.Option
}

func NewLimit(d *LimitDependency) *Limit {
	return &Limit{
		driver: d.Driver,
		limit: limit_verified.NewLimitVerified(
			d.Driver,
			limitVerifiedRedisV9.NewRedisStore(d.Redisc),
			d.Options...,
		),
	}
}

func (s *Limit) SendCode(ctx context.Context, c *CodeParam, opts ...CodeParamOption) error {
	return s.limit.SendCode(
		ctx,
		limit_verified.CodeParam{
			Kind:   c.Kind,
			Target: c.Target,
			Code:   c.Code,
		},
		opts...,
	)
}

func (s *Limit) VerifyCode(ctx context.Context, c *CodeParam) error {
	return s.limit.VerifyCode(
		ctx,
		limit_verified.CodeParam{
			Kind:   c.Kind,
			Target: c.Target,
			Code:   c.Code,
		})
}

func (s *Limit) Send(ctx context.Context, c *NoticeParam) error {
	err := s.limit.Incr(ctx, c.Target)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && errors.Is(err, limit_verified.ErrMaxSendPerDay) {
			_ = s.limit.Decr(ctx, c.Target)
		}
	}()
	err = s.driver.Send(ctx, c)
	return err
}
