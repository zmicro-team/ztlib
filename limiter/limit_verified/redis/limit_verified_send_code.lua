local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local maxSendPerDay = tonumber(ARGV[2]) -- 限制一天最大发送次数
local codeMaxSendPerDay = tonumber(ARGV[3]) -- 限制一天最大发送次数
local codeMaxErrorQuota = tonumber(ARGV[4]) -- 验证码最大验证失败次数
local codeAvailWindowSecond = tonumber(ARGV[5]) -- 验证码有效窗口时间, 单位: 秒
local codeResendIntervalSecond = tonumber(ARGV[6]) -- 验证码重发间隔时间
local now = tonumber(ARGV[7]) -- 当前时间, 单位秒
local expires = tonumber(ARGV[8]) -- global key 过期时间, 单位: 秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}"

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", 1)
local codeCnt = redis.call("HINCRBY", globalKey, "codeCnt", 1)
if sendCnt == 1 then
    redis.call("EXPIRE", globalKey, expires)
end
if sendCnt > maxSendPerDay or codeCnt > codeMaxSendPerDay then
	redis.call("HINCRBY", globalKey, "sendCnt", -1)
	redis.call("HINCRBY", globalKey, "codeCnt", -1)
    return 1 -- 超过每天发送限制次数
end

if (redis.call("EXISTS", codeKey) == 1) then
    local lastedAt = tonumber(redis.call("HGET", codeKey, "lasted"))
    if lastedAt + codeResendIntervalSecond > now then
		redis.call("HINCRBY", globalKey, "sendCnt", -1)
		redis.call("HINCRBY", globalKey, "codeCnt", -1)
        return 2 -- 发送过于频繁, 即还在重发限制窗口
    end
end

redis.call("HMSET", codeKey, "code", code, "quota", codeMaxErrorQuota, "err", 0, "lasted", now)
redis.call("EXPIRE", codeKey, codeAvailWindowSecond)

return 0 -- 成功