package tidings

import (
	"context"
	"errors"

	"github.com/zmicro-team/ztlib/limiter/limit_verified"
)

var (
	ErrSendFailed         = errors.New("tidings: send failed.")
	ErrNotSupportProvider = errors.New("tidings: not support provider.")
	ErrDuplicateProvider  = errors.New("tidings: duplicate provider.")
)

type CodeParamOption = limit_verified.CodeParamOption

var (
	WithMaxErrorQuota        = limit_verified.WithMaxErrorQuota
	WithAvailWindowSecond    = limit_verified.WithAvailWindowSecond
	WithResendIntervalSecond = limit_verified.WithResendIntervalSecond
)

// CodeParam 验证码参数
type CodeParam struct {
	Kind         string // O, 类别
	Target       string // M, 目标
	Code         string // M, 验证码
	TemplateCode string // M, 模版代码
}

// NoticeParam 通知参数
type NoticeParam struct {
	Target        string // M, 目标
	Template      string // 模板
	TemplateParam string // 模板参数
	Data          any    // 数据
}

type Provider interface {
	SendCode(context.Context, *CodeParam, ...CodeParamOption) error
	VerifyCode(context.Context, *CodeParam) error
	Send(context.Context, *NoticeParam) error
}

type Tidings struct {
	provider map[string]Provider
}

// New 实例化
func New() *Tidings {
	return &Tidings{
		provider: make(map[string]Provider),
	}
}

// RegisterProvider 注册 provider
func (t *Tidings) RegisterProvider(name string, p Provider) error {
	_, ok := t.provider[name]
	if ok {
		return ErrDuplicateProvider
	}
	t.provider[name] = p
	return nil
}

// SendCode 发送验证码
// opts: support WithAvailWindowSecond, WithResendIntervalSecond, WithMaxErrorQuota
func (t *Tidings) SendCode(ctx context.Context, name string, c *CodeParam, opts ...CodeParamOption) error {
	d, ok := t.provider[name]
	if !ok {
		return ErrNotSupportProvider
	}
	return d.SendCode(ctx, c, opts...)
}

// VerifyCode 验证验证码
// opts: only support WithMaxErrorQuota
func (t *Tidings) VerifyCode(ctx context.Context, name string, c *CodeParam) error {
	d, ok := t.provider[name]
	if !ok {
		return ErrNotSupportProvider
	}
	return d.VerifyCode(ctx, c)
}

// Send 通知
func (t *Tidings) Send(ctx context.Context, name string, n *NoticeParam) error {
	d, ok := t.provider[name]
	if !ok {
		return ErrNotSupportProvider
	}
	return d.Send(ctx, n)
}
