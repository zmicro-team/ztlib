package redis

import (
	_ "embed"
)

//go:embed period.lua
var PeriodLimitScript string

//go:embed  period_set_quota_full.lua
var PeriodLimitSetQuotaFullScript string

//go:embed  period_run_value.lua
var PeriodLimitRunValueScript string
