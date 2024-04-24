package redis

import _ "embed"

type LimitVerifiedState int

const (
	// inner lua send/verify code statue value
	InnerLimitVerifiedSuccess = 0
	// inner lua send code value
	InnerLimitVerifiedOfSendCodeReachMaxSendPerDay  = 1
	InnerLimitVerifiedOfSendCodeResendTooFrequently = 2
	// inner lua verify code value
	InnerLimitVerifiedOfVerifyCodeRequiredOrExpired   = 1
	InnerLimitVerifiedOfVerifyCodeReachMaxError       = 2
	InnerLimitVerifiedOfVerifyCodeVerificationFailure = 3
)

//go:embed limit_verified_send_code.lua
var LimitVerifiedSendCodeScript string

//go:embed limit_verified_rollback_send_cnt_and_code_cnt.lua
var LimitVerifiedRollbackSendCntAndCodeCntScript string

//go:embed limit_verified_verify_code.lua
var LimitVerifiedVerifyCodeScript string

//go:embed limit_verified_incr_send_cnt.lua
var LimitVerifiedIncrSendCntScript string

//go:embed limit_verified_decr_send_cnt.lua
var LimitVerifiedDecrSendCntScript string
