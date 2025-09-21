package database

import (
	"context"
	"flux/internal/domain"
	"time"

	generated "flux/internal/adapters/database/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresJobRepository struct {
	db      *pgxpool.Pool
	queries *generated.Queries
}

func NewPostgresJobRepository(db *pgxpool.Pool) domain.JobRepository {
	return &PostgresJobRepository{
		db:      db,
		queries: generated.New(db),
	}
}

func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	params := r.domainJobToCreateParams(job)
	_, err := r.queries.CreateJob(ctx, params)
	return err
}

func (r *PostgresJobRepository) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	dbJob, err := r.queries.GetJob(ctx, pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.dbJobToDomain(dbJob), nil
}

func (r *PostgresJobRepository) UpdateJob(ctx context.Context, job *domain.Job) error {
	job.UpdatedAt = time.Now()
	params := r.domainJobToUpdateParams(job)
	_, err := r.queries.UpdateJob(ctx, params)
	return err
}

func (r *PostgresJobRepository) DeleteJob(ctx context.Context, id string) error {
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, "DELETE FROM jobs WHERE id = $1", pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PostgresJobRepository) ListJobs(ctx context.Context, queue string, status domain.JobStatus, limit int) ([]*domain.Job, error) {
	var jobs []generated.Job
	var err error

	if queue != "" && status != "" {
		jobs, err = r.queries.ListJobs(ctx, generated.ListJobsParams{
			Column1: queue,
			Column2: string(status),
			Limit:   int32(limit),
		})
	} else if queue != "" {
		jobs, err = r.queries.ListJobsByQueue(ctx, generated.ListJobsByQueueParams{
			Queue: queue,
			Limit: int32(limit),
		})
	} else if status != "" {
		jobs, err = r.queries.ListJobsByStatus(ctx, generated.ListJobsByStatusParams{
			Status: string(status),
			Limit:  int32(limit),
		})
	} else {
		jobs, err = r.queries.ListJobs(ctx, generated.ListJobsParams{
			Column1: "",
			Column2: "",
			Limit:   int32(limit),
		})
	}

	if err != nil {
		return nil, err
	}

	domainJobs := make([]*domain.Job, len(jobs))
	for i, job := range jobs {
		domainJobs[i] = r.dbJobToDomain(job)
	}

	return domainJobs, nil
}

func (r *PostgresJobRepository) GetJobsReadyToRun(ctx context.Context, limit int) ([]*domain.Job, error) {
	jobs, err := r.queries.GetJobsReadyToRun(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	domainJobs := make([]*domain.Job, len(jobs))
	for i, job := range jobs {
		domainJobs[i] = r.dbJobToDomain(job)
	}

	return domainJobs, nil
}

func (r *PostgresJobRepository) GetScheduledJobs(ctx context.Context, limit int) ([]*domain.Job, error) {
	jobs, err := r.queries.GetScheduledJobs(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	domainJobs := make([]*domain.Job, len(jobs))
	for i, job := range jobs {
		domainJobs[i] = r.dbJobToDomain(job)
	}

	return domainJobs, nil
}

func (r *PostgresJobRepository) GetTimedOutJobs(ctx context.Context, limit int) ([]*domain.Job, error) {
	jobs, err := r.queries.GetTimedOutJobs(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	domainJobs := make([]*domain.Job, len(jobs))
	for i, job := range jobs {
		domainJobs[i] = r.dbJobToDomain(job)
	}

	return domainJobs, nil
}

func (r *PostgresJobRepository) UpdateJobStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	params := generated.UpdateJobStatusParams{
		ID:     pgtype.UUID{Bytes: jobUUID, Valid: true},
		Status: string(status),
	}

	_, err = r.queries.UpdateJobStatus(ctx, params)
	return err
}

func (r *PostgresJobRepository) domainJobToCreateParams(job *domain.Job) generated.CreateJobParams {
	jobUUID, _ := uuid.Parse(job.ID)

	return generated.CreateJobParams{
		ID:                   pgtype.UUID{Bytes: jobUUID, Valid: true},
		Type:                 job.Type,
		Queue:                job.Queue,
		Payload:              job.Payload,
		Priority:             int32(job.Priority),
		Status:               string(job.Status),
		ScheduleAt:           pgtype.Timestamptz{Time: job.ScheduleAt, Valid: true},
		TimeoutSec:           int32(job.TimeoutSec),
		Attempts:             int32(job.Attempts),
		MaxRetries:           int32(job.MaxRetries),
		BackoffPolicy:        job.BackoffPolicy,
		VisibilityTimeoutSec: int32(job.VisibilityTimeoutSec),
		IdempotencyKey:       r.stringToPgText(job.IdempotencyKey),
		LastError:            r.stringToPgText(job.LastError),
		CreatedAt:            pgtype.Timestamptz{Time: job.CreatedAt, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: job.UpdatedAt, Valid: true},
	}
}

func (r *PostgresJobRepository) domainJobToUpdateParams(job *domain.Job) generated.UpdateJobParams {
	jobUUID, _ := uuid.Parse(job.ID)

	return generated.UpdateJobParams{
		ID:                   pgtype.UUID{Bytes: jobUUID, Valid: true},
		Type:                 job.Type,
		Queue:                job.Queue,
		Payload:              job.Payload,
		Priority:             int32(job.Priority),
		Status:               string(job.Status),
		ScheduleAt:           pgtype.Timestamptz{Time: job.ScheduleAt, Valid: true},
		TimeoutSec:           int32(job.TimeoutSec),
		Attempts:             int32(job.Attempts),
		MaxRetries:           int32(job.MaxRetries),
		BackoffPolicy:        job.BackoffPolicy,
		VisibilityTimeoutSec: int32(job.VisibilityTimeoutSec),
		IdempotencyKey:       r.stringToPgText(job.IdempotencyKey),
		LastError:            r.stringToPgText(job.LastError),
		UpdatedAt:            pgtype.Timestamptz{Time: job.UpdatedAt, Valid: true},
	}
}

func (r *PostgresJobRepository) dbJobToDomain(dbJob generated.Job) *domain.Job {
	return &domain.Job{
		ID:                   r.pgUUIDToString(dbJob.ID),
		Type:                 dbJob.Type,
		Queue:                dbJob.Queue,
		Payload:              dbJob.Payload,
		Priority:             int(dbJob.Priority),
		Status:               domain.JobStatus(dbJob.Status),
		ScheduleAt:           dbJob.ScheduleAt.Time,
		TimeoutSec:           int(dbJob.TimeoutSec),
		Attempts:             int(dbJob.Attempts),
		MaxRetries:           int(dbJob.MaxRetries),
		BackoffPolicy:        dbJob.BackoffPolicy,
		VisibilityTimeoutSec: int(dbJob.VisibilityTimeoutSec),
		IdempotencyKey:       r.pgTextToString(dbJob.IdempotencyKey),
		LastError:            r.pgTextToString(dbJob.LastError),
		CreatedAt:            dbJob.CreatedAt.Time,
		UpdatedAt:            dbJob.UpdatedAt.Time,
	}
}

func (r *PostgresJobRepository) stringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func (r *PostgresJobRepository) pgTextToString(pt pgtype.Text) *string {
	if !pt.Valid {
		return nil
	}
	return &pt.String
}

func (r *PostgresJobRepository) pgUUIDToString(pu pgtype.UUID) string {
	if !pu.Valid {
		return ""
	}
	return uuid.UUID(pu.Bytes).String()
}

