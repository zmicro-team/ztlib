local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local now = tonumber(ARGV[2]) -- 当前时间, 单位秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}"

redis.call("HINCRBY", globalKey, "sendCnt", -1)
redis.call("HINCRBY", globalKey, "codeCnt", -1)
if (redis.call("EXISTS", codeKey) == 1) then
	local currentCode = redis.call("HGET", codeKey, "code")
	local lastedAt = tonumber(redis.call("HGET", codeKey, "lasted"))
	if currentCode == code and lastedAt == now then 
        redis.call("DEL", codeKey) -- 的确是你发送的, 删除 code key
	end
end