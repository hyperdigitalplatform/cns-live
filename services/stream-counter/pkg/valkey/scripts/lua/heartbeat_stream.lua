-- heartbeat_stream.lua
-- Update heartbeat timestamp and extend reservation TTL
--
-- KEYS: none
-- ARGV[1]: reservation_id (UUID)
-- ARGV[2]: ttl_extension (seconds, default 60)
--
-- RETURNS: {success (0|1), remaining_ttl, message}

local reservation_id = ARGV[1]
local ttl_extension = tonumber(ARGV[2]) or 60

-- Validate input
if not reservation_id then
    return {-1, 0, "Invalid reservation_id"}
end

local reservation_key = "stream:reservation:" .. reservation_id
local heartbeat_key = "stream:heartbeat:" .. reservation_id

-- Check if reservation exists
if redis.call('EXISTS', reservation_key) == 0 then
    return {0, 0, "Reservation not found"}
end

-- Update heartbeat timestamp
redis.call('SET', heartbeat_key, redis.call('TIME')[1])
redis.call('EXPIRE', heartbeat_key, 30)  -- Heartbeat expires in 30 seconds

-- Extend reservation TTL
local result = redis.call('EXPIRE', reservation_key, ttl_extension)

if result == 0 then
    return {0, 0, "Failed to extend reservation TTL"}
end

-- Get remaining TTL
local remaining_ttl = redis.call('TTL', reservation_key)

-- Return success
return {1, remaining_ttl, "Heartbeat updated"}
