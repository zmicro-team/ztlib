local key = KEYS[1]

local tb = {}
local cnt = tonumber(redis.call("GET", key))
if cnt == nil then 
    tb[1] = 0
    return tb
end
local ttl =  tonumber(redis.call("TTL", key))
if ttl == nil then 
    tb[1] = 0
    return tb
end
tb[1] = 1
tb[2] = cnt
tb[3] = ttl
return tb

