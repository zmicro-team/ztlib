local key = KEYS[1] -- key
local quota = tonumber(ARGV[1]) -- 限制次数
local window = tonumber(ARGV[2]) -- 限制时间
local success = tonumber(ARGV[3]) -- 是否成功

if success == 1 then
    local current = tonumber(redis.call("GET", key))
    if current == nil then
        return 0 -- 成功
    end
    if current < quota then -- 未超出失败最大次数限制范围, 成功, 并清除限制
        redis.call("DEL", key)
        return 0 -- 成功
    end
    return 2 -- 超过失败最大次数限制
end

local current = redis.call("INCRBY", key, 1)
if current == 1 then 
    redis.call("EXPIRE", key, window)
end 
if current <= quota then
    return 1 -- 还在限制范围, 只提示错误
end
return 2 -- 超过失败最大次数限制