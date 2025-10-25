-- Quota Validator Plugin Handler
-- Integrates with Stream Counter Service for quota enforcement

local http = require "resty.http"
local cjson = require "cjson"

local QuotaValidatorHandler = {
  VERSION = "1.0.0",
  PRIORITY = 1000  -- Execute before most plugins
}

-- Cache for quota statistics
local quota_cache = {}
local cache_expire_time = {}

-- Helper function to check quota availability
local function check_quota(conf, source)
  kong.log.debug("Checking quota for source: ", source)

  -- Check cache first
  local now = ngx.now()
  if quota_cache[source] and cache_expire_time[source] and cache_expire_time[source] > now then
    kong.log.debug("Using cached quota data for source: ", source)
    return quota_cache[source]
  end

  -- Query Stream Counter Service
  local httpc = http.new()
  httpc:set_timeout(conf.timeout)

  local stats_url = conf.stream_counter_url .. "/api/v1/stream/stats"
  kong.log.debug("Fetching stats from: ", stats_url)

  local res, err = httpc:request_uri(stats_url, {
    method = "GET",
    headers = {
      ["Content-Type"] = "application/json",
    }
  })

  if not res then
    kong.log.err("Failed to fetch quota stats: ", err)
    return nil, "Failed to connect to Stream Counter Service"
  end

  if res.status ~= 200 then
    kong.log.err("Stream Counter returned error: ", res.status, " ", res.body)
    return nil, "Stream Counter Service error"
  end

  local stats = cjson.decode(res.body)
  if not stats or not stats.stats then
    kong.log.err("Invalid stats response from Stream Counter")
    return nil, "Invalid response from Stream Counter"
  end

  -- Find stats for requested source
  local source_stats = nil
  for _, stat in ipairs(stats.stats) do
    if stat.source == source then
      source_stats = stat
      break
    end
  end

  if not source_stats then
    kong.log.warn("Source not found in stats: ", source)
    return nil, "Invalid source"
  end

  -- Cache the result
  quota_cache[source] = source_stats
  cache_expire_time[source] = now + conf.cache_ttl

  return source_stats
end

-- Helper function to create bilingual error response
local function quota_exceeded_response(source, stats)
  local messages = {
    DUBAI_POLICE = {
      en = "Camera limit reached for Dubai Police",
      ar = "تم الوصول إلى حد الكاميرات لشرطة دبي"
    },
    METRO = {
      en = "Camera limit reached for Metro",
      ar = "تم الوصول إلى حد الكاميرات للمترو"
    },
    BUS = {
      en = "Camera limit reached for Bus",
      ar = "تم الوصول إلى حد الكاميرات للحافلات"
    },
    OTHER = {
      en = "Camera limit reached for Other Agencies",
      ar = "تم الوصول إلى حد الكاميرات للجهات الأخرى"
    }
  }

  local msg = messages[source] or messages.OTHER

  return {
    error = {
      code = "RATE_LIMIT_EXCEEDED",
      message_en = msg.en,
      message_ar = msg.ar,
      source = source,
      current = stats.current,
      limit = stats.limit,
      available = stats.available,
      percentage = stats.percentage,
      retry_after = 30
    }
  }
end

-- Access phase: Validate quota before proxying
function QuotaValidatorHandler:access(conf)
  -- Only validate specific routes
  local route_path = kong.request.get_path()
  local should_validate = false

  for _, enabled_route in ipairs(conf.enabled_routes) do
    if string.find(route_path, enabled_route, 1, true) then
      should_validate = true
      break
    end
  end

  if not should_validate then
    kong.log.debug("Route not configured for quota validation: ", route_path)
    return
  end

  -- Only validate for reserve requests
  if route_path ~= "/api/v1/stream/reserve" then
    return
  end

  -- Parse request body to get source
  local body, err = kong.request.get_body()
  if err then
    kong.log.err("Failed to parse request body: ", err)
    return kong.response.exit(400, {
      error = {
        code = "INVALID_REQUEST",
        message_en = "Invalid request body",
        message_ar = "طلب غير صالح"
      }
    })
  end

  if not body or not body.source then
    kong.log.warn("Missing source in request body")
    return kong.response.exit(400, {
      error = {
        code = "MISSING_SOURCE",
        message_en = "Source is required",
        message_ar = "المصدر مطلوب"
      }
    })
  end

  local source = body.source

  -- Check quota
  local stats, err = check_quota(conf, source)
  if err then
    kong.log.err("Quota check failed: ", err)
    -- Fail open: allow request if quota service is unavailable
    kong.log.warn("Failing open due to quota service error")
    return
  end

  -- Check if quota exceeded
  if stats.available <= 0 then
    kong.log.warn("Quota exceeded for source: ", source, " (", stats.current, "/", stats.limit, ")")

    -- Add rate limit headers
    kong.response.set_header("X-RateLimit-Limit", stats.limit)
    kong.response.set_header("X-RateLimit-Remaining", 0)
    kong.response.set_header("X-RateLimit-Reset", 30)
    kong.response.set_header("Retry-After", 30)

    return kong.response.exit(429, quota_exceeded_response(source, stats))
  end

  -- Quota available, add headers and proceed
  kong.response.set_header("X-RateLimit-Limit", stats.limit)
  kong.response.set_header("X-RateLimit-Remaining", stats.available)
  kong.response.set_header("X-Quota-Percentage", stats.percentage)

  kong.log.info("Quota check passed for source: ", source, " (", stats.available, " available)")
end

-- Response phase: Add quota headers to response
function QuotaValidatorHandler:header_filter(conf)
  -- Add custom headers
  kong.response.set_header("X-Kong-Plugin", "quota-validator")
end

return QuotaValidatorHandler
