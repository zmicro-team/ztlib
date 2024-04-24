package redis

import (
	"time"

	_ "embed"
)

//go:embed token_limit.lua
var TokenLimitScript string

const (
	TokenLimitTokenFormat     = "{%s}.tokens"
	TokenLimitTimestampFormat = "{%s}.ts"
	TokenLimitPingInterval    = time.Millisecond * 100
)
