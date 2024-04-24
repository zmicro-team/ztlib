package limit_verified

import (
	"context"
	"time"
)

// StoreArgs store arguments
type StoreArgs struct {
	KeyPrefix                string        // 存验证码key的前缀
	Kind                     string        // CodeParam.Kind
	Target                   string        // CodeParam.Target
	KeyExpires               time.Duration // 存验证码key的过期时间
	MaxSendPerDay            int           // 限制一天最大发送次数(全局)
	Code                     string        // CodeParam.Code
	CodeMaxSendPerDay        int           // 验证码限制一天最大发送次数(验证码全局)
	CodeMaxErrorQuota        int           // 验证码最大验证失败次数
	CodeAvailWindowSecond    int           // 验证码有效窗口时间
	CodeResendIntervalSecond int           // 验证码重发间隔时间,
	NowSecond                string        // 当前时间, 秒
}

type RollbackArgs struct {
	KeyPrefix string // 存验证码key的前缀
	Kind      string // CodeParam.Kind
	Target    string // CodeParam.Target
	Code      string // CodeParam.Code
	NowSecond string // 当前时间, 秒
}

type VerifyArgs struct {
	KeyPrefix string // 存验证码key的前缀
	Kind      string // CodeParam.Kind
	Target    string // CodeParam.Target
	Code      string // CodeParam.Code
	NowSecond string // 当前时间, 秒
}

type IncrArgs struct {
	KeyPrefix     string        // 存验证码key的前缀
	Target        string        // CodeParam.Target
	KeyExpires    time.Duration // 存验证码key的过期时间
	MaxSendPerDay int           // 限制一天最大发送次数(全局)
}

type DecrArgs struct {
	KeyPrefix string // 存验证码key的前缀
	Target    string // CodeParam.Target
}

type Storage interface {
	// Store code
	Store(context.Context, *StoreArgs) error
	// Rollback when store success but send code failed.
	Rollback(context.Context, *RollbackArgs) error
	// Verify code
	Verify(context.Context, *VerifyArgs) error
	// Incr send cnt.
	Incr(context.Context, *IncrArgs) error
	// Decr send cnt.
	Decr(context.Context, *DecrArgs) error
}
