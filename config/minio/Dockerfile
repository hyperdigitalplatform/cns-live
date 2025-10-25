# MinIO Init Container
# Initializes buckets and policies for RTA CCTV System

FROM minio/mc:latest

# Copy initialization script
COPY init-buckets.sh /init-buckets.sh

# Make executable
RUN chmod +x /init-buckets.sh

# Run script
ENTRYPOINT ["/init-buckets.sh"]
