local buys_key = KEYS[1]
local sells_key = KEYS[2]
local order_id = ARGV[1]
local side = ARGV[2]
local price = tonumber(ARGV[3])
local remaining_qty = tonumber(ARGV[4])

local matches = {}

local function process_counter(counter_id, counter_price, zset_key)
    local counter_hash = redis.call('HGETALL', 'order:' .. counter_id)
    local counter = {}
    for j=1, #counter_hash, 2 do
        counter[counter_hash[j]] = counter_hash[j+1]
    end

    local counter_remaining = tonumber(counter.quantity or 0) - tonumber(counter.filled_qty or 0)
    if counter_remaining <= 0 then
        redis.call('ZREM', zset_key, counter_id)
        return 0
    end

    local match_qty = math.min(remaining_qty, counter_remaining)

    redis.call('HINCRBYFLOAT', 'order:' .. order_id, 'filled_qty', match_qty)
    redis.call('HINCRBYFLOAT', 'order:' .. counter_id, 'filled_qty', match_qty)

    counter_remaining = counter_remaining - match_qty

    if counter_remaining <= 0 then
        redis.call('ZREM', zset_key, counter_id)
        redis.call('HSET', 'order:' .. counter_id, 'status', 'FILLED')
    end

    table.insert(matches, {tonumber(counter_id), counter_price, match_qty})

    remaining_qty = remaining_qty - match_qty
    return match_qty
end

if side == 'BUY' then
    local sells = redis.call('ZRANGE', sells_key, 0, -1, 'WITHSCORES')
    for i=1, #sells, 2 do
        local counter_id = tonumber(sells[i])
        local counter_price = tonumber(sells[i+1])
        if counter_price > price then break end

        local matched = process_counter(counter_id, counter_price, sells_key)
        if matched > 0 and remaining_qty <= 0 then
            redis.call('ZREM', buys_key, order_id)
            redis.call('HSET', 'order:' .. order_id, 'status', 'FILLED')
            break
        end
    end
else
    local buys = redis.call('ZREVRANGE', buys_key, 0, -1, 'WITHSCORES')
    for i=1, #buys, 2 do
        local counter_id = tonumber(buys[i])
        local counter_price = tonumber(buys[i+1]) * -1
        if counter_price < price then break end

        local matched = process_counter(counter_id, counter_price, buys_key)
        if matched > 0 and remaining_qty <= 0 then
            redis.call('ZREM', sells_key, order_id)
            redis.call('HSET', 'order:' .. order_id, 'status', 'FILLED')
            break
        end
    end
end

-- Return JSON manual
if #matches == 0 then return "[]" end

local json = "["
for i, m in ipairs(matches) do
    if i > 1 then json = json .. "," end
    json = json .. "{\"counter_id\":" .. m[1] .. ",\"price\":" .. m[2] .. ",\"quantity\":" .. m[3] .. "}"
end
return json .. "]"