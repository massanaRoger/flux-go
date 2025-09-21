# Flux

A production-style distributed job queue & workflow orchestrator in Go. Think "a tiny Celery/Sidekiq + Argo Workflows," but approachable and well-documented.

## Overview

Flux is a complete job processing system where developers can submit jobs via HTTP/gRPC, and a fleet of Go workers execute them with retries, priorities, and scheduling. Built to showcase Go's strengths: concurrency, networking, observability, performance, and clean tooling.

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ API Server  │    │  Scheduler  │    │   Workers   │
│ (REST/gRPC) │    │   (Cron)    │    │ (Goroutine  │
│             │    │             │    │   Pools)    │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
    ┌─────────────────────────────────────────────┐
    │              Message Broker                 │
    │         (Redis Streams + PostgreSQL)        │
    └─────────────────────────────────────────────┘
```

### Components

- **API Server**: REST + gRPC endpoints to enqueue jobs, query status, define workflows
- **Scheduler**: Picks ready jobs (immediate/cron/delayed), manages retries & backoff
- **Worker Agent**: Pulls jobs, executes them in goroutine pools, handles graceful shutdown
- **Broker & Store**: Redis Streams for queuing, PostgreSQL for job metadata and workflow DAGs
- **Web Dashboard**: Minimal UI to view queues, workers, and job runs
- **Observability**: Prometheus metrics, health/liveness endpoints

## Features

### Core Features (MVP)
- ✅ Enqueue jobs with JSON payload, priority, delay, timeout, max retries
- ✅ Exponential backoff + dead-letter queue after max retries
- ✅ Idempotency keys (avoid duplicates)
- ✅ Worker pools with configurable concurrency per queue
- ✅ At-least-once processing with visibility timeouts
- ✅ Cron schedules (`*/5 * * * *`) and one-shot delayed jobs
- ✅ Metrics: queue depth, processed/sec, success/failure rates, p95 latency

### Stretch Features
- 🔄 Workflow DAGs: steps with dependencies, pass outputs → inputs
- 🔄 Rate limiting per queue or job type
- 🔄 S3/MinIO artifact store for large results
- 🔄 JWT auth + API keys
- 🔄 Pluggable broker: Redis now, Kafka/Rabbit later

## Quick Start

### Prerequisites
- Go 1.23+
- Docker & Docker Compose
- Make

### 1. Start Infrastructure

```bash
# Start PostgreSQL and Redis
make db-up

# Run database migrations
make db-migrate
```

### 2. Start Services

```bash
# Terminal 1: API Server (port 8080)
make run-api

# Terminal 2: Scheduler
make run-scheduler

# Terminal 3: Worker
make run-worker

# Terminal 4: Dashboard (port 8081)
make run-dashboard
```

### 3. Submit a Job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": {"to": "user@example.com", "subject": "Hello"},
    "queue": "default",
    "priority": 5
  }'
```

### 4. View Dashboard

Open http://localhost:8081 to monitor queues, workers, and job execution.

## Development

### Building

```bash
# Generate SQL code and build all binaries
make build
```

### Testing

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate test coverage report
make test-coverage
```

### Database Operations

```bash
# Reset database (drop + recreate + migrate)
make db-reset

# Create new migration
make db-create-migration name=add_workflow_table

# Rollback last migration
make db-rollback
```

### Code Generation

```bash
# Generate SQL queries (using sqlc)
make sqlc-generate
```

## API Reference

### REST Endpoints

```
POST   /api/v1/jobs           # Enqueue job
GET    /api/v1/jobs/{id}      # Get job status
GET    /api/v1/jobs           # List jobs (with filters)
DELETE /api/v1/jobs/{id}      # Cancel job
POST   /api/v1/workflows      # Create workflow
GET    /metrics               # Prometheus metrics
GET    /health                # Health check
```

### Job Payload

```json
{
  "type": "email",                    // Job type
  "payload": {"key": "value"},        // Job data (JSON)
  "queue": "default",                 // Queue name
  "priority": 5,                      // Priority (1-10)
  "delay": "30s",                     // Delay before execution
  "timeout": "5m",                    // Max execution time
  "max_retries": 3,                   // Retry attempts
  "idempotency_key": "unique-id",     // Prevent duplicates
  "cron": "0 */6 * * *"              // Cron schedule (optional)
}
```

## Configuration

Environment variables:

```bash
# Database
DATABASE_URL=postgres://flux:flux123@localhost:5432/flux?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# API Server
API_PORT=8080
GRPC_PORT=9090

# Worker
WORKER_CONCURRENCY=10
WORKER_QUEUES=default,email,reports

# Scheduler
SCHEDULER_INTERVAL=1s
```

## Monitoring

### Metrics

Available at `/metrics` endpoint (Prometheus format):

- `flux_jobs_total` - Total jobs processed
- `flux_jobs_duration_seconds` - Job execution duration
- `flux_queue_depth` - Current queue depth
- `flux_workers_active` - Active worker count
- `flux_jobs_failed_total` - Failed job count

### Health Checks

- `GET /health` - Overall system health
- `GET /health/db` - Database connectivity
- `GET /health/redis` - Redis connectivity

## Project Structure

```
.
├── cmd/                    # Application entrypoints
│   ├── api-server/        # REST/gRPC API server
│   ├── scheduler/         # Job scheduler
│   ├── worker/           # Job worker
│   └── dashboard/        # Web dashboard
├── internal/             # Private application code
│   ├── domain/          # Core business logic
│   ├── app/            # Application services
│   └── adapters/       # Infrastructure adapters
│       ├── http/       # HTTP handlers
│       ├── database/   # PostgreSQL adapter
│       └── queue/      # Redis adapter
├── db/                 # Database migrations
├── deployments/        # Kubernetes/Docker configs
└── scripts/           # Build and deployment scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes following the existing patterns
4. Run tests: `make test`
5. Commit: `git commit -m 'Add amazing feature'`
6. Push: `git push origin feature/amazing-feature`
7. Open a Pull Request

## License

[MIT License](LICENSE)

## FAQ

**Q: How does Flux compare to Sidekiq/Celery?**
A: Flux is lighter weight but provides similar core functionality. It's designed to be embedded in Go applications rather than run as a separate service.

**Q: Can I use Flux in production?**
A: Flux is designed with production patterns but should be thoroughly tested in your environment first.

**Q: How do I scale workers?**
A: Run multiple worker processes/containers. They'll automatically coordinate through Redis streams.

**Q: What happens if a worker crashes?**
A: Jobs have visibility timeouts. If a worker doesn't heartbeat, the scheduler will make the job available for retry.