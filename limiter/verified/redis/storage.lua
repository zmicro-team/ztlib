local key = KEYS[1]    -- key
local value = ARGV[1] -- 值
local quota = tonumber(ARGV[2]) -- 最大错误限制次数
local expires = tonumber(ARGV[3]) -- 过期时间

redis.call("HMSET", key, "value", value, "quota", quota, "err", 0)
redis.call("EXPIRE", key, expires)
return 0    -- 成功