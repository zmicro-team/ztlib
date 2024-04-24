package redis

import (
	_ "embed"
)

//go:embed period_failure_fixed.lua
var PeriodFailureLimitFixedScript string

//go:embed period_failure_fixed_set_quota_full.lua
var PeriodFailureLimitFixedSetQuotaFullScript string

//go:embed  period_failure_run_value.lua
var PeriodFailureLimitRunValueScript string

const (
	// inner lua code
	// InnerPeriodFailureLimitCodeSuccess means success.
	InnerPeriodFailureLimitCodeSuccess = 0
	// InnerPeriodFailureLimitCodeInQuota means within the quota.
	InnerPeriodFailureLimitCodeInQuota = 1
	// InnerPeriodFailureLimitCodeOverQuota means passed the quota.
	InnerPeriodFailureLimitCodeOverQuota = 2
)
