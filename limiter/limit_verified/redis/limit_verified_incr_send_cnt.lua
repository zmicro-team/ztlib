local keyPrefix = KEYS[1] -- keyPrefix
local target = KEYS[2] -- target
local maxSendPerDay = tonumber(ARGV[1]) -- 限制一天最大发送次数
local expires = tonumber(ARGV[2]) -- global key 过期时间, 单位: 秒

local globalKey = keyPrefix .. target

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", 1)
if sendCnt == 1 then
    redis.call("EXPIRE", globalKey, expires)
end
if sendCnt > maxSendPerDay then
	redis.call("HINCRBY", globalKey, "sendCnt", -1)
    return 1 -- 超过每天发送限制次数
end
return 0 -- 成功