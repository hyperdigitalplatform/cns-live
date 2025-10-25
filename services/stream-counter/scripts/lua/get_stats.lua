-- get_stats.lua
-- Get current stream statistics for all sources
--
-- KEYS: none
-- ARGV[1]: sources (comma-separated, e.g., "DUBAI_POLICE,METRO,BUS,OTHER")
--
-- RETURNS: JSON-like string with stats

local sources_str = ARGV[1] or "DUBAI_POLICE,METRO,BUS,OTHER"
local sources = {}

-- Parse comma-separated sources
for source in string.gmatch(sources_str, '([^,]+)') do
    table.insert(sources, source)
end

local stats = {}

for _, source in ipairs(sources) do
    local limit_key = "stream:limit:" .. source
    local count_key = "stream:count:" .. source

    local limit = tonumber(redis.call('GET', limit_key) or 0)
    local current = tonumber(redis.call('GET', count_key) or 0)

    -- Calculate percentage
    local percentage = 0
    if limit > 0 then
        percentage = math.floor((current / limit) * 100)
    end

    table.insert(stats, {source, current, limit, percentage})
end

return stats
