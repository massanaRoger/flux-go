-- name: CreateJob :one
INSERT INTO jobs (
    id, type, queue, payload, priority, status, schedule_at,
    timeout_sec, attempts, max_retries, backoff_policy,
    visibility_timeout_sec, idempotency_key, last_error,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: UpdateJob :one
UPDATE jobs
SET type = $2, queue = $3, payload = $4, priority = $5, status = $6,
    schedule_at = $7, timeout_sec = $8, attempts = $9, max_retries = $10,
    backoff_policy = $11, visibility_timeout_sec = $12,
    idempotency_key = $13, last_error = $14, updated_at = $15
WHERE id = $1
RETURNING *;

-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;

-- name: ListJobs :many
SELECT * FROM jobs
WHERE ($1::text = '' OR queue = $1)
  AND ($2::text = '' OR status = $2)
ORDER BY created_at DESC
LIMIT $3;

-- name: ListJobsByQueue :many
SELECT * FROM jobs
WHERE queue = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListJobsByStatus :many
SELECT * FROM jobs
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetJobsReadyToRun :many
SELECT * FROM jobs
WHERE status IN ('pending', 'scheduled')
  AND schedule_at <= NOW()
ORDER BY priority DESC, schedule_at ASC
LIMIT $1;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateJobAttempts :one
UPDATE jobs
SET attempts = $2, last_error = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetJobByIdempotencyKey :one
SELECT * FROM jobs WHERE idempotency_key = $1;

-- name: GetScheduledJobs :many
SELECT * FROM jobs
WHERE status = 'scheduled'
  AND schedule_at <= NOW()
ORDER BY priority DESC, schedule_at ASC
LIMIT $1;

-- name: GetTimedOutJobs :many
SELECT * FROM jobs
WHERE status = 'running'
  AND updated_at < NOW() - INTERVAL '1 second' * visibility_timeout_sec
ORDER BY updated_at ASC
LIMIT $1;