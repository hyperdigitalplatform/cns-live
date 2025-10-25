-- reserve_stream.lua
-- Atomically check limit and reserve stream slot
--
-- KEYS: none
-- ARGV[1]: source (e.g., "DUBAI_POLICE")
-- ARGV[2]: reservation_id (UUID)
-- ARGV[3]: camera_id (UUID)
-- ARGV[4]: user_id
-- ARGV[5]: ttl (seconds)
--
-- RETURNS: {success (0|1), current_count, limit}

local source = ARGV[1]
local reservation_id = ARGV[2]
local camera_id = ARGV[3]
local user_id = ARGV[4]
local ttl = tonumber(ARGV[5])

-- Validate inputs
if not source or not reservation_id or not camera_id or not user_id or not ttl then
    return {-1, 0, 0, "Invalid arguments"}
end

-- Key patterns
local limit_key = "stream:limit:" .. source
local count_key = "stream:count:" .. source
local reservation_key = "stream:reservation:" .. reservation_id

-- Get current limit and count
local limit = tonumber(redis.call('GET', limit_key) or 0)
local current = tonumber(redis.call('GET', count_key) or 0)

-- Check if limit is reached
if current >= limit then
    return {0, current, limit}  -- Reject: limit reached
end

-- Atomically increment counter
local new_count = redis.call('INCR', count_key)

-- Double-check after increment (race condition safety)
if new_count > limit then
    -- Rollback: decrement counter
    redis.call('DECR', count_key)
    return {0, limit, limit}  -- Reject: limit reached during increment
end

-- Create reservation with metadata
redis.call('HSET', reservation_key,
    'camera_id', camera_id,
    'source', source,
    'user_id', user_id,
    'created_at', redis.call('TIME')[1],
    'expires_at', redis.call('TIME')[1] + ttl
)

-- Set TTL on reservation
redis.call('EXPIRE', reservation_key, ttl)

-- Create heartbeat key
local heartbeat_key = "stream:heartbeat:" .. reservation_id
redis.call('SET', heartbeat_key, redis.call('TIME')[1])
redis.call('EXPIRE', heartbeat_key, 30)  -- 30 second heartbeat

-- Log reservation (for monitoring)
local log_key = "stream:log:reserve"
redis.call('LPUSH', log_key,
    string.format('%s|%s|%s|%s', redis.call('TIME')[1], source, camera_id, user_id)
)
redis.call('LTRIM', log_key, 0, 999)  -- Keep last 1000 entries

-- Return success with new count
return {1, new_count, limit}
