package database

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"flux/internal/domain"
	"flux/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type JobRepositoryIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	repo      *PostgresJobRepository
	ctx       context.Context
}

func (suite *JobRepositoryIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.container, suite.pool = testutil.SetupTestDatabase(suite.T(), suite.ctx)
	suite.repo = NewPostgresJobRepository(suite.pool).(*PostgresJobRepository)
}

func (suite *JobRepositoryIntegrationTestSuite) TearDownSuite() {
	testutil.CleanupTestDatabase(suite.T(), suite.ctx, suite.container, suite.pool)
}

func (suite *JobRepositoryIntegrationTestSuite) SetupTest() {
	testutil.TruncateTables(suite.T(), suite.ctx, suite.pool)
}


func (suite *JobRepositoryIntegrationTestSuite) createTestJob() *domain.Job {
	payload := json.RawMessage(`{"message": "test payload"}`)
	idempotencyKey := fmt.Sprintf("test-key-%d", time.Now().UnixNano())

	return &domain.Job{
		Type:                 "test-job",
		Queue:                "default",
		Payload:              payload,
		Priority:             0,
		Status:               domain.JobStatusPending,
		ScheduleAt:           time.Now(),
		TimeoutSec:           300,
		Attempts:             0,
		MaxRetries:           3,
		BackoffPolicy:        "exp",
		VisibilityTimeoutSec: 60,
		IdempotencyKey:       &idempotencyKey,
	}
}

func (suite *JobRepositoryIntegrationTestSuite) TestCreateJob() {
	job := suite.createTestJob()

	err := suite.repo.CreateJob(suite.ctx, job)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), job.ID)
	assert.NotZero(suite.T(), job.CreatedAt)
	assert.NotZero(suite.T(), job.UpdatedAt)
}

func (suite *JobRepositoryIntegrationTestSuite) TestCreateJobWithExistingID() {
	job := suite.createTestJob()
	job.ID = uuid.New().String()
	originalID := job.ID

	err := suite.repo.CreateJob(suite.ctx, job)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), originalID, job.ID)
}

func (suite *JobRepositoryIntegrationTestSuite) TestCreateJobWithDuplicateIdempotencyKey() {
	job1 := suite.createTestJob()
	job2 := suite.createTestJob()
	job2.IdempotencyKey = job1.IdempotencyKey

	err1 := suite.repo.CreateJob(suite.ctx, job1)
	err2 := suite.repo.CreateJob(suite.ctx, job2)

	assert.NoError(suite.T(), err1)
	assert.Error(suite.T(), err2) // Should fail due to unique constraint
}

func (suite *JobRepositoryIntegrationTestSuite) TestGetJob() {
	job := suite.createTestJob()
	err := suite.repo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	retrievedJob, err := suite.repo.GetJob(suite.ctx, job.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedJob)
	assert.Equal(suite.T(), job.ID, retrievedJob.ID)
	assert.Equal(suite.T(), job.Type, retrievedJob.Type)
	assert.Equal(suite.T(), job.Queue, retrievedJob.Queue)
	assert.Equal(suite.T(), job.Status, retrievedJob.Status)
	assert.Equal(suite.T(), job.Priority, retrievedJob.Priority)
	assert.JSONEq(suite.T(), string(job.Payload), string(retrievedJob.Payload))
}

func (suite *JobRepositoryIntegrationTestSuite) TestGetJobNotFound() {
	nonExistentID := uuid.New().String()

	retrievedJob, err := suite.repo.GetJob(suite.ctx, nonExistentID)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrievedJob)
}

func (suite *JobRepositoryIntegrationTestSuite) TestGetJobInvalidUUID() {
	invalidID := "invalid-uuid"

	retrievedJob, err := suite.repo.GetJob(suite.ctx, invalidID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedJob)
}

func (suite *JobRepositoryIntegrationTestSuite) TestUpdateJob() {
	job := suite.createTestJob()
	err := suite.repo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	// Update job properties
	job.Status = domain.JobStatusRunning
	job.Attempts = 1
	job.Priority = 5
	errorMsg := "test error"
	job.LastError = &errorMsg

	err = suite.repo.UpdateJob(suite.ctx, job)

	assert.NoError(suite.T(), err)

	// Verify the update
	retrievedJob, err := suite.repo.GetJob(suite.ctx, job.ID)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), domain.JobStatusRunning, retrievedJob.Status)
	assert.Equal(suite.T(), 1, retrievedJob.Attempts)
	assert.Equal(suite.T(), 5, retrievedJob.Priority)
	assert.Equal(suite.T(), "test error", *retrievedJob.LastError)
	assert.True(suite.T(), retrievedJob.UpdatedAt.After(retrievedJob.CreatedAt))
}

func (suite *JobRepositoryIntegrationTestSuite) TestUpdateJobNotFound() {
	job := suite.createTestJob()
	job.ID = uuid.New().String()

	err := suite.repo.UpdateJob(suite.ctx, job)

	assert.Error(suite.T(), err)
}

func (suite *JobRepositoryIntegrationTestSuite) TestDeleteJob() {
	job := suite.createTestJob()
	err := suite.repo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	err = suite.repo.DeleteJob(suite.ctx, job.ID)

	assert.NoError(suite.T(), err)

	// Verify job is deleted
	retrievedJob, err := suite.repo.GetJob(suite.ctx, job.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrievedJob)
}

func (suite *JobRepositoryIntegrationTestSuite) TestDeleteJobNotFound() {
	nonExistentID := uuid.New().String()

	err := suite.repo.DeleteJob(suite.ctx, nonExistentID)

	assert.Error(suite.T(), err)
}

func (suite *JobRepositoryIntegrationTestSuite) TestListJobs() {
	// Create test jobs
	job1 := suite.createTestJob()
	job1.Queue = "queue1"
	job1.Status = domain.JobStatusPending

	job2 := suite.createTestJob()
	job2.Queue = "queue2"
	job2.Status = domain.JobStatusRunning

	job3 := suite.createTestJob()
	job3.Queue = "queue1"
	job3.Status = domain.JobStatusPending

	for _, job := range []*domain.Job{job1, job2, job3} {
		err := suite.repo.CreateJob(suite.ctx, job)
		require.NoError(suite.T(), err)
	}

	// Test: List all jobs
	allJobs, err := suite.repo.ListJobs(suite.ctx, "", "", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), allJobs, 3)

	// Test: List jobs by queue
	queue1Jobs, err := suite.repo.ListJobs(suite.ctx, "queue1", "", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), queue1Jobs, 2)
	for _, job := range queue1Jobs {
		assert.Equal(suite.T(), "queue1", job.Queue)
	}

	// Test: List jobs by status
	pendingJobs, err := suite.repo.ListJobs(suite.ctx, "", domain.JobStatusPending, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), pendingJobs, 2)
	for _, job := range pendingJobs {
		assert.Equal(suite.T(), domain.JobStatusPending, job.Status)
	}

	// Test: List jobs by queue and status
	queue1PendingJobs, err := suite.repo.ListJobs(suite.ctx, "queue1", domain.JobStatusPending, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), queue1PendingJobs, 2)
	for _, job := range queue1PendingJobs {
		assert.Equal(suite.T(), "queue1", job.Queue)
		assert.Equal(suite.T(), domain.JobStatusPending, job.Status)
	}

	// Test: Limit results
	limitedJobs, err := suite.repo.ListJobs(suite.ctx, "", "", 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), limitedJobs, 2)
}

func (suite *JobRepositoryIntegrationTestSuite) TestGetJobsReadyToRun() {
	now := time.Now()

	// Create jobs with different statuses and schedule times
	readyJob1 := suite.createTestJob()
	readyJob1.Status = domain.JobStatusPending
	readyJob1.ScheduleAt = now.Add(-1 * time.Hour) // Past time
	readyJob1.Priority = 10

	readyJob2 := suite.createTestJob()
	readyJob2.Status = domain.JobStatusScheduled
	readyJob2.ScheduleAt = now.Add(-30 * time.Minute) // Past time
	readyJob2.Priority = 5

	futureJob := suite.createTestJob()
	futureJob.Status = domain.JobStatusPending
	futureJob.ScheduleAt = now.Add(1 * time.Hour) // Future time

	runningJob := suite.createTestJob()
	runningJob.Status = domain.JobStatusRunning
	runningJob.ScheduleAt = now.Add(-1 * time.Hour)

	for _, job := range []*domain.Job{readyJob1, readyJob2, futureJob, runningJob} {
		err := suite.repo.CreateJob(suite.ctx, job)
		require.NoError(suite.T(), err)
	}

	readyJobs, err := suite.repo.GetJobsReadyToRun(suite.ctx, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), readyJobs, 2)

	// Should be ordered by priority DESC, then schedule_at ASC
	assert.Equal(suite.T(), readyJob1.ID, readyJobs[0].ID) // Higher priority
	assert.Equal(suite.T(), readyJob2.ID, readyJobs[1].ID)
}

func (suite *JobRepositoryIntegrationTestSuite) TestGetJobsReadyToRunWithLimit() {
	now := time.Now()

	// Create multiple ready jobs
	for i := 0; i < 5; i++ {
		job := suite.createTestJob()
		job.Status = domain.JobStatusPending
		job.ScheduleAt = now.Add(-1 * time.Hour)
		err := suite.repo.CreateJob(suite.ctx, job)
		require.NoError(suite.T(), err)
	}

	readyJobs, err := suite.repo.GetJobsReadyToRun(suite.ctx, 3)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), readyJobs, 3)
}

func (suite *JobRepositoryIntegrationTestSuite) TestJobWithNilFields() {
	job := suite.createTestJob()
	job.IdempotencyKey = nil
	job.LastError = nil

	err := suite.repo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	retrievedJob, err := suite.repo.GetJob(suite.ctx, job.ID)
	require.NoError(suite.T(), err)

	assert.Nil(suite.T(), retrievedJob.IdempotencyKey)
	assert.Nil(suite.T(), retrievedJob.LastError)
}

func TestJobRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JobRepositoryIntegrationTestSuite))
}