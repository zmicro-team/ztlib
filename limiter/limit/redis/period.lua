local key = KEYS[1]
local quota = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCRBY", key, 1)
if current == 1 then
    redis.call("EXPIRE", key, window)
end
if current < quota then
    return 0 -- allow
elseif current == quota then
    return 1 -- hit quota
else
    return 2 -- over quata
end