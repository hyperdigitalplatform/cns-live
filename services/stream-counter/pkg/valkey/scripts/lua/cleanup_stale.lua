-- cleanup_stale.lua
-- Cleanup stale reservations that have expired but not been properly released
-- This is a maintenance script run periodically
--
-- KEYS: none
-- ARGV[1]: max_age_seconds (cleanup reservations older than this, default 3600)
--
-- RETURNS: {cleaned_count, sources_affected}

local max_age = tonumber(ARGV[1]) or 3600
local current_time = redis.call('TIME')[1]
local cleaned_count = 0
local sources_affected = {}

-- Scan for all reservation keys
local cursor = "0"
local reservations = {}

repeat
    local result = redis.call('SCAN', cursor, 'MATCH', 'stream:reservation:*', 'COUNT', 100)
    cursor = result[1]
    local keys = result[2]

    for _, key in ipairs(keys) do
        -- Check if key still exists (might have expired)
        if redis.call('EXISTS', key) == 1 then
            -- Get created_at timestamp
            local created_at = tonumber(redis.call('HGET', key, 'created_at') or 0)
            local age = current_time - created_at

            -- Check if reservation is stale
            if age > max_age then
                -- Get source before deleting
                local source = redis.call('HGET', key, 'source')

                if source then
                    -- Decrement counter
                    local count_key = "stream:count:" .. source
                    local new_count = redis.call('DECR', count_key)

                    -- Ensure count doesn't go negative
                    if new_count < 0 then
                        redis.call('SET', count_key, 0)
                    end

                    -- Track affected sources
                    if not sources_affected[source] then
                        sources_affected[source] = 0
                    end
                    sources_affected[source] = sources_affected[source] + 1

                    cleaned_count = cleaned_count + 1
                end

                -- Delete reservation
                redis.call('DEL', key)

                -- Delete associated heartbeat
                local reservation_id = string.match(key, 'stream:reservation:(.+)')
                if reservation_id then
                    local heartbeat_key = "stream:heartbeat:" .. reservation_id
                    redis.call('DEL', heartbeat_key)
                end
            end
        end
    end
until cursor == "0"

-- Convert sources_affected to array
local sources_list = {}
for source, count in pairs(sources_affected) do
    table.insert(sources_list, source .. ":" .. count)
end

return {cleaned_count, table.concat(sources_list, ",")}
