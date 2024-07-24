package limit_verified

import (
	"time"
)

// OptionSetter option setter
type OptionSetter interface {
	setKeyPrefix(k string)
	setKeyExpires(expires time.Duration)
	setMaxSendPerDay(cnt int)
	setCodeMaxSendPerDay(cnt int)
	setCodeMaxErrorQuota(cnt int)
	setCodeAvailWindowSecond(sec int)
	setCodeResendIntervalSecond(sec int)
}

// Option LimitVerified 选项
type Option func(OptionSetter)

// WithKeyPrefix redis存验证码key的前缀, 默认 limit:verified:
func WithKeyPrefix(k string) Option {
	return func(v OptionSetter) {
		v.setKeyPrefix(k)
	}
}

// WithKeyExpires redis存验证码key的过期时间, 默认 24小时
func WithKeyExpires(expires time.Duration) Option {
	return func(v OptionSetter) {
		v.setKeyExpires(expires)
	}
}

// WithMaxSendPerDay 限制一天最大发送次数(全局), 默认: 10
func WithMaxSendPerDay(cnt int) Option {
	return func(v OptionSetter) {
		v.setMaxSendPerDay(cnt)
	}
}

// WithMaxSendPerDay 验证码限制一天最大发送次数(验证码全局), 默认: 10
func WithCodeMaxSendPerDay(cnt int) Option {
	return func(v OptionSetter) {
		v.setCodeMaxSendPerDay(cnt)
	}
}

// WithCodeMaxErrorQuota 验证码最大验证失败次数, 默认: 3
func WithCodeMaxErrorQuota(cnt int) Option {
	return func(v OptionSetter) {
		v.setCodeMaxErrorQuota(cnt)
	}
}

// WithCodeAvailWindowSecond 验证码有效窗口时间, 默认180, 单位: 秒
func WithCodeAvailWindowSecond(sec int) Option {
	return func(v OptionSetter) {
		v.setCodeAvailWindowSecond(sec)
	}
}

// WithCodeResendIntervalSecond 验证码重发间隔时间, 默认60, 单位: 秒
func WithCodeResendIntervalSecond(sec int) Option {
	return func(v OptionSetter) {
		v.setCodeResendIntervalSecond(sec)
	}
}

type CodeParam struct {
	Kind                     string // optional, 默认为: DefaultKind
	Target                   string // required
	Code                     string // required
	TemplateCode             string // required
	codeMaxErrorQuota        int    // 验证码最大验证失败次数, 默认: 3
	codeAvailWindowSecond    int    // 验证码有效窗口时间, 默认180, 单位: 秒
	codeResendIntervalSecond int    // 验证码重发间隔时间, 默认60, 单位: 秒
}

// CodeParamOption LimitVerified code 选项
type CodeParamOption func(*CodeParam)

// WithMaxErrorQuota 验证码最大验证失败次数, 覆盖默认值
func WithMaxErrorQuota(cnt int) CodeParamOption {
	return func(v *CodeParam) {
		v.codeMaxErrorQuota = cnt
	}
}

// WithAvailWindowSecond 验证码有效窗口时间, 覆盖默认值, 单位: 秒
func WithAvailWindowSecond(sec int) CodeParamOption {
	return func(v *CodeParam) {
		v.codeAvailWindowSecond = sec
	}
}

// WithResendIntervalSecond 重发验证码间隔时间, 覆盖默认值, 单位: 秒
func WithResendIntervalSecond(sec int) CodeParamOption {
	return func(v *CodeParam) {
		v.codeResendIntervalSecond = sec
	}
}

func (v *LimitVerified[P, S]) takeCodeParamOption(c *CodeParam, opts ...CodeParamOption) {
	if c.Kind == "" {
		c.Kind = DefaultKind
	}
	c.codeMaxErrorQuota = v.codeMaxErrorQuota
	c.codeAvailWindowSecond = v.codeAvailWindowSecond
	c.codeResendIntervalSecond = v.codeResendIntervalSecond
	for _, opt := range opts {
		opt(c)
	}
}
