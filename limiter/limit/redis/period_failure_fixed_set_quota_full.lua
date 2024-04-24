local key = KEYS[1]
local quota = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = tonumber(redis.call("GET", key))
if current == nil then 
	redis.call("SETEX", key, window, quota)
elseif current < quota then 
	redis.call("SET", key, quota, "KEEPTTL")
end
return 0