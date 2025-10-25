-- Quota Validator Plugin Schema
-- Validates stream quota before proxying requests

local typedefs = require "kong.db.schema.typedefs"

return {
  name = "quota-validator",
  fields = {
    {
      config = {
        type = "record",
        fields = {
          {
            stream_counter_url = {
              type = "string",
              required = true,
              default = "http://stream-counter:8087"
            }
          },
          {
            validate_before_proxy = {
              type = "boolean",
              required = true,
              default = true
            }
          },
          {
            cache_ttl = {
              type = "number",
              required = true,
              default = 5
            }
          },
          {
            timeout = {
              type = "number",
              required = true,
              default = 1000  -- 1 second
            }
          },
          {
            enabled_routes = {
              type = "array",
              elements = { type = "string" },
              default = { "/api/v1/stream/reserve" }
            }
          }
        }
      }
    }
  }
}
