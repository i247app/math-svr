#!/bin/bash

# Stop script for Math-AI observability stack
set -e

echo "🛑 Stopping Math-AI Observability Stack..."
echo ""

# Stop the observability stack
docker-compose -f docker-compose.otel.yml down

echo ""
echo "✅ Observability stack stopped successfully!"
echo ""
echo "📝 Data is preserved in Docker volumes."
echo "   To remove all data: docker-compose -f docker-compose.otel.yml down -v"
echo ""
