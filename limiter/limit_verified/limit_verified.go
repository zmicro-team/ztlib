package limit_verified

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

const DefaultKind = "default"

// error defined for verified
var (
	// ErrUnknownCode is an error that represents unknown status code.
	ErrUnknownCode           = errors.New("limit: unknown status code")
	ErrMaxSendPerDay         = errors.New("limit: reach the maximum send times")
	ErrResendTooFrequently   = errors.New("limit: resend too frequently")
	ErrCodeRequiredOrExpired = errors.New("limit: code is required or expired")
	ErrCodeMaxErrorQuota     = errors.New("limit: over the maximum error quota")
	ErrCodeVerification      = errors.New("limit: code verified failed")
)

// LimitVerifiedProvider the provider
type LimitVerifiedProvider interface {
	Name() string
	SendCode(CodeParam) error
}

// LimitVerified limit verified code
type LimitVerified[P LimitVerifiedProvider, S Storage] struct {
	p             LimitVerifiedProvider // LimitVerifiedProvider send code
	store         Storage               // store client
	keyPrefix     string                // store 存验证码key的前缀, 默认 limit:verified:
	keyExpires    time.Duration         // store 存验证码key的过期时间, 默认: 24 小时
	maxSendPerDay int                   // 限制一天最大发送次数(全局), 默认: 10
	// 以下只针对验证码进行限制
	codeMaxSendPerDay        int // 验证码限制一天最大发送次数(验证码全局), 默认: 10, codeMaxSendPerDay <= maxSendPerDay
	codeMaxErrorQuota        int // 验证码最大验证失败次数, 默认: 3
	codeAvailWindowSecond    int // 验证码有效窗口时间, 默认180, 单位: 秒
	codeResendIntervalSecond int // 验证码重发间隔时间, 默认60, 单位: 秒
}

// NewLimitVerified  new a limit verified
func NewLimitVerified[P LimitVerifiedProvider, S Storage](p P, store S, opts ...Option) *LimitVerified[P, S] {
	v := &LimitVerified[P, S]{
		p,
		store,
		"limit:verified:",
		time.Hour * 24,
		10,
		10,
		3,
		180,
		60,
	}
	for _, opt := range opts {
		opt(v)
	}
	if v.codeMaxSendPerDay > v.maxSendPerDay {
		v.codeMaxSendPerDay = v.maxSendPerDay
	}
	return v
}

// Name the provider name
func (v *LimitVerified[P, S]) Name() string { return v.p.Name() }

// SendCode send code and store.
func (v *LimitVerified[P, S]) SendCode(ctx context.Context, c CodeParam, opts ...CodeParamOption) error {
	v.takeCodeParamOption(&c, opts...)

	nowSecond := strconv.FormatInt(time.Now().Unix(), 10)
	err := v.store.Store(ctx, &StoreArgs{
		KeyPrefix:                v.keyPrefix,
		Kind:                     c.Kind,
		Target:                   c.Target,
		KeyExpires:               v.keyExpires,
		MaxSendPerDay:            v.maxSendPerDay,
		Code:                     c.Code,
		CodeMaxSendPerDay:        v.codeMaxSendPerDay,
		CodeMaxErrorQuota:        c.codeMaxErrorQuota,
		CodeAvailWindowSecond:    c.codeAvailWindowSecond,
		CodeResendIntervalSecond: c.codeResendIntervalSecond,
		NowSecond:                nowSecond,
	})
	if err != nil {
		return err
	}
	// 发送失败, 回滚发送次数
	defer func() {
		if err != nil && !errors.Is(err, ErrMaxSendPerDay) {
			_ = v.store.Rollback(context.Background(), &RollbackArgs{
				KeyPrefix: v.keyPrefix,
				Kind:      c.Kind,
				Target:    c.Target,
				Code:      c.Code,
				NowSecond: nowSecond,
			})
		}
	}()
	err = v.p.SendCode(c)
	return err
}

// VerifyCode verify code from cache.
func (v *LimitVerified[P, S]) VerifyCode(ctx context.Context, c CodeParam) error {
	v.takeCodeParamOption(&c)
	return v.store.Verify(ctx, &VerifyArgs{
		KeyPrefix: v.keyPrefix,
		Kind:      c.Kind,
		Target:    c.Target,
		Code:      c.Code,
		NowSecond: strconv.FormatInt(time.Now().Unix(), 10),
	})
}

// Incr send cnt.
func (v *LimitVerified[P, S]) Incr(ctx context.Context, target string) error {
	return v.store.Incr(ctx, &IncrArgs{
		KeyPrefix:     v.keyPrefix,
		Target:        target,
		KeyExpires:    v.keyExpires,
		MaxSendPerDay: v.maxSendPerDay,
	})
}

// Decr send cnt.
func (v *LimitVerified[P, S]) Decr(ctx context.Context, target string) error {
	return v.store.Decr(ctx, &DecrArgs{
		KeyPrefix: v.keyPrefix,
		Target:    target,
	})
}

func (v *LimitVerified[P, S]) setKeyPrefix(k string) {
	if k != "" {
		if !strings.HasSuffix(k, ":") {
			k += ":"
		}
		v.keyPrefix = k
	}
}
func (v *LimitVerified[P, S]) setKeyExpires(expires time.Duration) {
	v.keyExpires = expires
}
func (v *LimitVerified[P, S]) setMaxSendPerDay(cnt int) {
	v.maxSendPerDay = cnt
}
func (v *LimitVerified[P, S]) setCodeMaxSendPerDay(cnt int) {
	v.codeMaxSendPerDay = cnt
}
func (v *LimitVerified[P, S]) setCodeMaxErrorQuota(cnt int) {
	v.codeMaxErrorQuota = cnt
}
func (v *LimitVerified[P, S]) setCodeAvailWindowSecond(sec int) {
	v.codeAvailWindowSecond = sec
}
func (v *LimitVerified[P, S]) setCodeResendIntervalSecond(sec int) {
	v.codeResendIntervalSecond = sec
}
