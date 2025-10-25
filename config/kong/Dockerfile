# Kong API Gateway with Custom Plugins
# For RTA CCTV System

FROM kong:3.4

# Set user to root for setup
USER root

# Install dependencies (Kong 3.4 is Ubuntu-based)
RUN apt-get update && apt-get install -y \
    curl \
    jq \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Create plugin directory
RUN mkdir -p /etc/kong/plugins/quota-validator

# Copy custom plugin files
COPY plugins/quota-validator/ /etc/kong/plugins/quota-validator/

# Copy Kong configuration
COPY kong.conf /etc/kong/kong.conf
COPY kong.yml /etc/kong/kong.yml

# Set permissions
RUN chown -R kong:kong /etc/kong

# Switch back to kong user
USER kong

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
  CMD kong health || exit 1

# Expose ports
# 8000: Proxy HTTP
# 8443: Proxy HTTPS
# 8001: Admin API HTTP
# 8444: Admin API HTTPS
# 8100: Status API
EXPOSE 8000 8443 8001 8444 8100

# Start Kong
CMD ["kong", "docker-start"]
