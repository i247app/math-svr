#!/bin/bash

# Quick start script for Math-AI observability stack
set -e

echo "🚀 Starting Math-AI Observability Stack..."
echo ""

# Start the observability stack
docker-compose -f docker-compose.otel.yml up -d

echo ""
echo "✅ Observability stack started successfully!"
echo ""
echo "📊 Access the following services:"
echo "   Jaeger (Tracing):    http://localhost:16686"
echo "   Prometheus (Metrics): http://localhost:9090"
echo "   Grafana (Dashboards): http://localhost:3000"
echo ""
echo "🔑 Grafana credentials: admin / admin"
echo ""
echo "📝 Next steps:"
echo "   1. Ensure your .env has OTEL_ENABLE_TRACING=true and OTEL_ENABLE_METRICS=true"
echo "   2. Start math-srv with: make run"
echo "   3. Generate some traffic to see traces and metrics"
echo ""
echo "🛑 To stop the stack: ./stop.sh"
echo ""
