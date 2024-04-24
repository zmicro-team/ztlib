local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local now = tonumber(ARGV[2]) -- 当前时间, 单位秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}" 

if redis.call("EXISTS", codeKey) == 0 then
    return 1  -- 未发送短信验证码 或 验证码已过期
end

local errCnt = tonumber(redis.call('HGET', codeKey, "err"))
local codeMaxErrorQuota = tonumber(redis.call('HGET', codeKey, "quota"))
local currentCode = redis.call('HGET', codeKey, "code")
if errCnt >= codeMaxErrorQuota then
    return 2  -- 验证码错误次数超过限制
end
if currentCode == code then
    redis.call("DEL", codeKey) -- 删除 code key
    return 0 -- 成功
else
    redis.call('HINCRBY', codeKey, "err", 1)
    return 3 -- 验证码错误
end