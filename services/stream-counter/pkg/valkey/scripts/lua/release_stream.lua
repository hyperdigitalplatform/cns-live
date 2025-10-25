-- release_stream.lua
-- Atomically release stream reservation and decrement counter
--
-- KEYS: none
-- ARGV[1]: reservation_id (UUID)
--
-- RETURNS: {success (0|1), new_count, source}

local reservation_id = ARGV[1]

-- Validate input
if not reservation_id then
    return {-1, 0, "Invalid reservation_id"}
end

local reservation_key = "stream:reservation:" .. reservation_id

-- Check if reservation exists
if redis.call('EXISTS', reservation_key) == 0 then
    return {0, 0, "Reservation not found"}
end

-- Get source from reservation
local source = redis.call('HGET', reservation_key, 'source')

if not source then
    return {0, 0, "Invalid reservation: missing source"}
end

-- Decrement counter
local count_key = "stream:count:" .. source
local new_count = redis.call('DECR', count_key)

-- Ensure count doesn't go negative
if new_count < 0 then
    redis.call('SET', count_key, 0)
    new_count = 0
end

-- Delete reservation
redis.call('DEL', reservation_key)

-- Delete heartbeat
local heartbeat_key = "stream:heartbeat:" .. reservation_id
redis.call('DEL', heartbeat_key)

-- Log release (for monitoring)
local camera_id = redis.call('HGET', reservation_key, 'camera_id') or 'unknown'
local user_id = redis.call('HGET', reservation_key, 'user_id') or 'unknown'
local log_key = "stream:log:release"
redis.call('LPUSH', log_key,
    string.format('%s|%s|%s|%s', redis.call('TIME')[1], source, camera_id, user_id)
)
redis.call('LTRIM', log_key, 0, 999)  -- Keep last 1000 entries

-- Return success with new count
return {1, new_count, source}
