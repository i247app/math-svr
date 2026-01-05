# Math-AI Server Observability Stack

This directory contains the OpenTelemetry observability stack for the Math-AI Server, providing comprehensive monitoring, tracing, and metrics visualization.

## 🏗️ Architecture

The observability stack consists of three main components:

- **Jaeger**: Distributed tracing backend for visualizing request flows
- **Prometheus**: Time-series database for metrics collection
- **Grafana**: Visualization platform for dashboards and alerts

## 🚀 Quick Start

### 1. Start the Observability Stack

```bash
cd docker/observability
docker-compose -f docker-compose.otel.yml up -d
```

### 2. Configure Math-AI Server

Ensure your `.env` file has the following OpenTelemetry configuration:

```bash
# OpenTelemetry Configuration
OTEL_ENABLE_TRACING=true
OTEL_ENABLE_METRICS=true
OTEL_SERVICE_NAME=math-srv
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=development
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_TRACE_SAMPLE_RATE=1.0
OTEL_PROMETHEUS_PORT=9091
```

### 3. Start Math-AI Server

```bash
# From project root
make run
```

### 4. Access the UIs

| Service    | URL                      | Credentials         |
|------------|--------------------------|---------------------|
| Jaeger     | http://localhost:16686   | None (open access)  |
| Prometheus | http://localhost:9090    | None (open access)  |
| Grafana    | http://localhost:3000    | admin / admin       |

## 📊 Using the Observability Stack

### Jaeger - Distributed Tracing

**Access**: http://localhost:16686

Jaeger shows you the complete journey of each request through your system:

1. **Find Traces**: Select "math-srv" service and click "Find Traces"
2. **View Details**: Click on a trace to see the complete span tree
3. **Analyze Performance**: Identify slow operations and bottlenecks

**What You Can See:**
- Complete HTTP request lifecycle
- Database queries with timing
- AI provider API calls
- Error propagation through the stack
- Distributed context across services

**Example Trace Flow:**
```
HTTP POST /generate-quiz (2.5s)
├─ db.query: SELECT user profile (12ms)
├─ ai.send_message: Google Gemini (2.3s)
│  ├─ Prompt tokens: 450
│  └─ Completion tokens: 1200
└─ db.exec: INSERT quiz (8ms)
```

### Prometheus - Metrics Collection

**Access**: http://localhost:9090

Prometheus collects time-series metrics from your application:

**Useful Queries:**

```promql
# HTTP request rate
rate(http_server_request_count_total[5m])

# 95th percentile HTTP latency
histogram_quantile(0.95, rate(http_server_request_duration_bucket[5m]))

# Database query rate by operation
sum by(db_operation) (rate(db_query_count_total[5m]))

# AI provider request rate
rate(ai_request_count_total[5m])

# AI token usage
rate(ai_token_usage_total[5m])

# Active HTTP requests
http_server_active_requests
```

### Grafana - Visualization

**Access**: http://localhost:3000
**Default Credentials**: admin / admin (change on first login)

A pre-configured dashboard "Math-AI Server Overview" is automatically loaded with:

#### HTTP Metrics
- Request rate by endpoint
- Active requests gauge
- Latency percentiles (p50, p95, p99)
- Status code distribution

#### Database Metrics
- Query rate by operation type
- Query latency percentiles
- Success vs error rates

#### AI Provider Metrics
- Request rate by provider
- Latency by provider and model
- Token usage (prompt vs completion)

#### Creating Custom Dashboards

1. Click "+" → "Dashboard"
2. Add panel → Select "Prometheus" datasource
3. Write PromQL query
4. Configure visualization
5. Save dashboard

## 📈 Available Metrics

### HTTP Metrics
| Metric Name | Type | Description |
|-------------|------|-------------|
| `http_server_request_count` | Counter | Total HTTP requests by method, route, status |
| `http_server_request_duration` | Histogram | Request duration in milliseconds |
| `http_server_active_requests` | Gauge | Current active requests |
| `http_server_response_size` | Counter | Response size in bytes |

### Database Metrics
| Metric Name | Type | Description |
|-------------|------|-------------|
| `db_query_count` | Counter | Total DB queries by operation and status |
| `db_query_duration` | Histogram | Query duration in milliseconds |
| `db_active_connections` | Gauge | Active database connections |

### AI Provider Metrics
| Metric Name | Type | Description |
|-------------|------|-------------|
| `ai_request_count` | Counter | AI requests by provider, operation, status |
| `ai_request_duration` | Histogram | AI request duration in milliseconds |
| `ai_token_usage` | Counter | Tokens used (prompt and completion) |

### External Service Metrics
| Metric Name | Type | Description |
|-------------|------|-------------|
| `s3_operation_count` | Counter | S3 operations by type and status |
| `s3_operation_duration` | Histogram | S3 operation duration |
| `sms_sent_count` | Counter | SMS messages sent |
| `email_sent_count` | Counter | Emails sent |

## 🔍 Troubleshooting

### No Metrics in Prometheus

1. **Check Prometheus targets**: http://localhost:9090/targets
   - Should show `math-srv` target as "UP"
   - If "DOWN", check if app is exposing metrics on port 9091

2. **Verify app is running**: `curl http://localhost:9091/metrics`
   - Should return Prometheus-format metrics

3. **Check Prometheus logs**: `docker logs math-srv-prometheus`

### No Traces in Jaeger

1. **Verify OTLP endpoint**: Check app can reach `localhost:4318`
   - Test: `curl http://localhost:4318`

2. **Check Jaeger logs**: `docker logs math-srv-jaeger`

3. **Verify tracing is enabled**: `.env` has `OTEL_ENABLE_TRACING=true`

4. **Check sample rate**: If `OTEL_TRACE_SAMPLE_RATE=0.1`, only 10% of requests are traced

### Grafana Dashboard Not Loading

1. **Check datasources**: Grafana → Configuration → Data Sources
   - Prometheus and Jaeger should be configured

2. **Refresh dashboards**: Grafana → Dashboards → Manage → Reload

3. **Check Grafana logs**: `docker logs math-srv-grafana`

## 🔧 Configuration

### Adjusting Trace Sampling

For **production**, reduce sampling to minimize overhead:

```bash
OTEL_TRACE_SAMPLE_RATE=0.1  # Sample 10% of requests
```

For **development**, use 100% sampling:

```bash
OTEL_TRACE_SAMPLE_RATE=1.0  # Sample all requests
```

### Prometheus Retention

Default retention is 30 days. To change:

Edit `docker-compose.otel.yml`:
```yaml
command:
  - '--storage.tsdb.retention.time=90d'  # Change to 90 days
```

### Grafana Plugins

Additional plugins can be installed via environment variable:

```yaml
environment:
  - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-clock-panel
```

## 🛑 Stopping the Stack

```bash
# Stop services but keep data
docker-compose -f docker-compose.otel.yml stop

# Stop and remove containers (keeps volumes/data)
docker-compose -f docker-compose.otel.yml down

# Stop and remove everything including data
docker-compose -f docker-compose.otel.yml down -v
```

## 📊 Performance Impact

The observability stack has minimal performance impact:

- **Tracing overhead**: ~0.5-1ms per request (at 100% sampling)
- **Metrics overhead**: <0.1ms (negligible)
- **Recommended production sampling**: 10% (reduces overhead to ~0.05-0.1ms)

## 🔗 Useful Links

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger UI Guide](https://www.jaegertracing.io/docs/latest/frontend-ui/)
- [Prometheus Query Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Tutorials](https://grafana.com/tutorials/)

## 📝 Next Steps

1. **Explore Traces**: Generate some quiz requests and view them in Jaeger
2. **Monitor Metrics**: Check real-time metrics in Grafana dashboard
3. **Set Up Alerts**: Configure Grafana alerts for high latency or errors
4. **Custom Dashboards**: Create dashboards for your specific use cases

## 🆘 Support

For issues or questions:
- Check service logs: `docker-compose -f docker-compose.otel.yml logs <service-name>`
- Verify network connectivity between services
- Ensure ports are not already in use (16686, 4318, 9090, 3000)
