#!/bin/bash

# Port conflict checker for TinyApp Gateway testing
echo "🔍 Checking for port conflicts..."

PORTS=(5001 8005 9000 9090)
CONFLICTS=()

for port in "${PORTS[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "⚠️  Port $port is already in use:"
        lsof -Pi :$port -sTCP:LISTEN
        CONFLICTS+=($port)
    else
        echo "✅ Port $port is available"
    fi
done

if [ ${#CONFLICTS[@]} -gt 0 ]; then
    echo ""
    echo "❌ Port conflicts detected on: ${CONFLICTS[*]}"
    echo ""
    echo "💡 Solutions:"
    echo "1. Stop the services using these ports"
    echo "2. Kill specific processes: sudo kill -9 \$(lsof -t -i:PORT)"
    echo "3. On macOS, disable AirPlay Receiver if port 5000/5001 conflicts:"
    echo "   System Preferences > Sharing > AirPlay Receiver (turn off)"
    echo ""
    exit 1
else
    echo ""
    echo "🎉 All required ports are available!"
    echo "You can now run the test scripts:"
    echo "  ./test-simple.sh"
    echo "  ./test-oidc.sh"
fi