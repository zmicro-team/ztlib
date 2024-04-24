local keyPrefix = KEYS[1] -- keyPrefix
local target = KEYS[2] -- target

local globalKey = keyPrefix .. target

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", -1)
if sendCnt < 0 then
    redis.call("DEL", globalKey)
end
return 0