package redis

import _ "embed"

//go:embed storage.lua
var StorageScript string

//go:embed match.lua
var MatchScript string
