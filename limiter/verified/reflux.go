package verified

import (
	"context"
	"time"
)

// RefluxProvider the reflux provider
type RefluxProvider interface {
	Name() string
	GenerateUniqueId() string
}

// Reflux verified reflux limiter
type Reflux[P RefluxProvider, S Storage] struct {
	p              P             // CaptchaProvider generate captcha
	store          S             // store client
	disableOneTime bool          // 禁用一次性验证
	keyPrefix      string        // store 存验证码key的前缀, 默认 verified:reflux:
	keyExpires     time.Duration // store 存验证码key的过期时间, 默认: 3 分种
	maxErrQuota    int           // store 验证码验证最大错误次数限制, 默认: 1
}

// NewVerifiedReflux new reflux instance.
func NewVerifiedReflux[P RefluxProvider, S Storage](p P, s S, opts ...Option) *Reflux[P, S] {
	v := &Reflux[P, S]{
		p,
		s,
		false,
		"verified:reflux:",
		time.Minute * 3,
		1,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Name the provider name
func (v *Reflux[P, S]) Name() string { return v.p.Name() }

// Generate generate uniqueId. use GenerateOption overwrite default key expires
func (v *Reflux[P, S]) Generate(ctx context.Context, kind, key string, opts ...GenerateOption) (string, error) {
	genOpt := generateOption{
		keyExpires:  v.keyExpires,
		maxErrQuota: v.maxErrQuota,
	}
	for _, f := range opts {
		f(&genOpt)
	}
	answer := v.p.GenerateUniqueId()
	err := v.store.Store(ctx, &StoreArgs{
		v.disableOneTime,
		v.keyPrefix + kind + ":" + key,
		genOpt.keyExpires,
		genOpt.maxErrQuota,
		answer,
	})
	if err != nil {
		return "", err
	}
	return answer, nil
}

// Verify the answer.
func (v *Reflux[P, S]) Verify(ctx context.Context, kind, key, answer string) bool {
	return v.store.Verify(ctx,
		&VerifyArgs{
			v.disableOneTime,
			v.keyPrefix + kind + ":" + key,
			answer,
		},
	)
}

func (v *Reflux[P, S]) setKeyPrefix(k string)         { v.keyPrefix = k }
func (v *Reflux[P, S]) setKeyExpires(t time.Duration) { v.keyExpires = t }
func (v *Reflux[P, S]) setMaxErrQuota(quota int) {
	v.disableOneTime = true
	v.maxErrQuota = quota
}
