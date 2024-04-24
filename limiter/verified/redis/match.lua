local key = KEYS[1]    -- key
local value = ARGV[1] -- 答案

if redis.call("EXISTS", key) == 0 then
	return 1   -- 键不存在, 验证失败
end

local wantValue = redis.call("HGET", key, "value")
if wantValue == value then
	redis.call("DEL", key)
	return 0  -- 成功
else 
	local quota = tonumber(redis.call("HGET", key, "quota"))
	local errCnt = redis.call("HINCRBY", key, "err", 1)
	if errCnt >= quota then 
		redis.call("DEL", key)
	end
		return 1   -- 值不相等, 验证失败
end