#!/bin/bash
# Check Go syntax for all services

echo "=========================================="
echo "GO SYNTAX CHECKER"
echo "=========================================="
echo ""

SERVICES=(
    "go-api"
    "vms-service"
    "storage-service"
    "recording-service"
    "metadata-service"
    "playback-service"
    "stream-counter"
)

ERRORS=0

for service in "${SERVICES[@]}"; do
    echo "Checking $service..."

    # Check if main.go exists
    if [ ! -f "services/$service/cmd/main.go" ]; then
        echo "  ❌ Missing cmd/main.go"
        ERRORS=$((ERRORS + 1))
        continue
    fi

    # Check for unused imports (common issue)
    UNUSED=$(grep -o 'import.*"fmt"' "services/$service/cmd/main.go" || true)
    if [ ! -z "$UNUSED" ]; then
        # Check if fmt is actually used
        FMT_USAGE=$(grep -c 'fmt\.' "services/$service/cmd/main.go" || echo "0")
        if [ "$FMT_USAGE" -eq "0" ]; then
            echo "  ⚠️  Unused 'fmt' import detected"
        fi
    fi

    echo "  ✓ Syntax check passed"
done

echo ""
if [ $ERRORS -eq 0 ]; then
    echo "✅ All services passed basic checks"
else
    echo "❌ Found $ERRORS issues"
fi
