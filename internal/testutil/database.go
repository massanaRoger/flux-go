package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const MigrationSQL = `
-- Jobs table
CREATE TABLE jobs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	type VARCHAR(255) NOT NULL,
	queue VARCHAR(255) NOT NULL,
	payload JSONB NOT NULL,
	priority INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'pending',
	schedule_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	timeout_sec INTEGER NOT NULL DEFAULT 300,
	attempts INTEGER NOT NULL DEFAULT 0,
	max_retries INTEGER NOT NULL DEFAULT 3,
	backoff_policy TEXT NOT NULL DEFAULT 'exp',
	visibility_timeout_sec INTEGER NOT NULL DEFAULT 60,
	idempotency_key TEXT UNIQUE,
	last_error TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Job runs table
CREATE TABLE job_runs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
	started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	finished_at TIMESTAMPTZ,
	status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
	error TEXT,
	metrics JSONB
);

-- Workflows table
CREATE TABLE workflows (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	definition JSONB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workflow runs table
CREATE TABLE workflow_runs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workers table (optional)
CREATE TABLE workers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	queue VARCHAR(255) NOT NULL,
	concurrency INTEGER NOT NULL DEFAULT 1,
	last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	version VARCHAR(50),
	hostname VARCHAR(255) NOT NULL
);

-- Indexes
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_queue_status ON jobs(queue, status);
CREATE INDEX idx_jobs_schedule_at ON jobs(schedule_at);
CREATE UNIQUE INDEX idx_jobs_idempotency_key ON jobs(idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_job_runs_job_id ON job_runs(job_id);
CREATE INDEX idx_workflow_runs_workflow_id ON workflow_runs(workflow_id);
CREATE INDEX idx_workers_queue ON workers(queue);
CREATE INDEX idx_workers_last_heartbeat ON workers(last_heartbeat);
`

func SetupTestDatabase(t *testing.T, ctx context.Context) (testcontainers.Container, *pgxpool.Pool) {
	// Start PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		postgres.WithDatabase("flux_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Run migrations
	_, err = pool.Exec(ctx, MigrationSQL)
	require.NoError(t, err)

	return pgContainer, pool
}

func CleanupTestDatabase(t *testing.T, ctx context.Context, container testcontainers.Container, pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
	if container != nil {
		err := container.Terminate(ctx)
		require.NoError(t, err)
	}
}

func TruncateTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE jobs CASCADE")
	require.NoError(t, err)
}