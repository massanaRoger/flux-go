# Claude Code Configuration for flux-go

## Project Overview
This is a Go-based job scheduling and workflow management system using hexagonal architecture with PostgreSQL and Redis.

## Code Formatting and Quality Commands

### Format and Lint
```bash
# Format all Go code
go fmt ./...

# Static analysis and vet checks
go vet ./...

# Generate SQLC database code
make sqlc-generate
```

### Testing
```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Generate test coverage report
make test-coverage
```

### Build and Run
```bash
# Build all components
make build

# Run individual services
make run-api         # API server
make run-scheduler   # Job scheduler
make run-worker      # Job worker
make run-dashboard   # Dashboard
```

### Database Operations
```bash
# Start database
make db-up

# Run migrations
make db-migrate

# Reset database
make db-reset

# Create new migration
make db-create-migration name=migration_name
```

## Project Structure

```
flux-go/
├── cmd/                    # Application entry points
│   ├── api-server/        # REST API server
│   ├── scheduler/         # Job scheduler service
│   ├── worker/           # Job worker service
│   └── dashboard/        # Web dashboard
├── internal/
│   ├── domain/           # Business logic and entities
│   ├── app/              # Application services
│   └── adapters/         # External interfaces
│       ├── database/     # PostgreSQL adapter
│       ├── http/         # HTTP handlers
│       └── queue/        # Redis queue adapter
├── db/
│   ├── migrations/       # Database schema migrations
│   └── queries/          # SQLC queries
└── Makefile             # Build and development commands
```

## Code Style Guidelines

### Go Standards
- Follow standard Go conventions (gofmt, go vet)
- Use hexagonal architecture patterns
- Domain entities in `internal/domain/`
- Application services in `internal/app/`
- External adapters in `internal/adapters/`

### Database
- Use SQLC for type-safe SQL queries
- Store queries in `db/queries/`
- Migrations in `db/migrations/`
- Generated code in `internal/adapters/database/generated/`

### Testing
- Unit tests: `*_test.go` files alongside source
- Integration tests for database and HTTP adapters
- Use testcontainers for integration testing
- Maintain test coverage with `make test-coverage`

### Dependencies
- PostgreSQL database
- Redis for job queuing
- Gin for HTTP routing
- SQLC for database code generation
- Testcontainers for testing

## Development Workflow
1. Make code changes
2. Run `go fmt ./...` to format
3. Run `go vet ./...` for static analysis
4. Run `make test` to ensure tests pass
5. Use `make sqlc-generate` if database queries changed