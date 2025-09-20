package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/internal/adapters/database"
	"flux/internal/app"
	"flux/internal/domain"
	"flux/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	router    *gin.Engine
	ctx       context.Context
}

func (suite *HTTPIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	gin.SetMode(gin.TestMode)

	suite.container, suite.pool = testutil.SetupTestDatabase(suite.T(), suite.ctx)

	// Set up application
	jobRepo := database.NewPostgresJobRepository(suite.pool)
	jobService := app.NewJobService(jobRepo)
	jobHandler := NewJobHandler(jobService)

	// Set up router
	suite.router = gin.New()
	v1 := suite.router.Group("/api/v1")
	{
		v1.POST("/jobs", jobHandler.CreateJob)
		v1.GET("/jobs/:id", jobHandler.GetJob)
		v1.GET("/jobs", jobHandler.ListJobs)
		v1.DELETE("/jobs/:id", jobHandler.DeleteJob)
	}
}

func (suite *HTTPIntegrationTestSuite) TearDownSuite() {
	testutil.CleanupTestDatabase(suite.T(), suite.ctx, suite.container, suite.pool)
}

func (suite *HTTPIntegrationTestSuite) SetupTest() {
	testutil.TruncateTables(suite.T(), suite.ctx, suite.pool)
}

func (suite *HTTPIntegrationTestSuite) TestCreateJob() {
	payload := json.RawMessage(`{"message": "test job payload"}`)
	createReq := CreateJobRequest{
		Type:     "email",
		Queue:    "default",
		Payload:  payload,
		Priority: 5,
	}

	body, err := json.Marshal(createReq)
	require.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(body))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusCreated, recorder.Code)

	var response domain.Job
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), response.ID)
	assert.Equal(suite.T(), "email", response.Type)
	assert.Equal(suite.T(), "default", response.Queue)
	assert.Equal(suite.T(), 5, response.Priority)
	assert.Equal(suite.T(), domain.JobStatusPending, response.Status)
	assert.JSONEq(suite.T(), string(payload), string(response.Payload))
}

func (suite *HTTPIntegrationTestSuite) TestCreateJobInvalidPayload() {
	body := []byte(`{"invalid": "json"}`)

	req, err := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(body))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

func (suite *HTTPIntegrationTestSuite) TestGetJob() {
	// First create a job directly in database
	payload := json.RawMessage(`{"message": "get job test"}`)
	job := domain.NewJob("test-job", "default", payload)
	job.Priority = 3

	jobRepo := database.NewPostgresJobRepository(suite.pool)
	err := jobRepo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	// Now get it via HTTP
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%s", job.ID), nil)
	require.NoError(suite.T(), err)

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response domain.Job
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), job.ID, response.ID)
	assert.Equal(suite.T(), "test-job", response.Type)
	assert.Equal(suite.T(), "default", response.Queue)
	assert.Equal(suite.T(), 3, response.Priority)
}

func (suite *HTTPIntegrationTestSuite) TestGetJobNotFound() {
	// Use a valid UUID format that doesn't exist
	nonExistentID := "123e4567-e89b-12d3-a456-426614174000"
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%s", nonExistentID), nil)
	require.NoError(suite.T(), err)

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
}

func (suite *HTTPIntegrationTestSuite) TestListJobs() {
	// Create test jobs directly in database
	jobRepo := database.NewPostgresJobRepository(suite.pool)

	job1 := domain.NewJob("email", "default", json.RawMessage(`{"to": "user1@example.com"}`))
	job2 := domain.NewJob("sms", "notifications", json.RawMessage(`{"phone": "+1234567890"}`))

	err := jobRepo.CreateJob(suite.ctx, job1)
	require.NoError(suite.T(), err)
	err = jobRepo.CreateJob(suite.ctx, job2)
	require.NoError(suite.T(), err)

	// Test list all jobs
	req, err := http.NewRequest("GET", "/api/v1/jobs", nil)
	require.NoError(suite.T(), err)

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response map[string][]domain.Job
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	jobs := response["jobs"]
	assert.Len(suite.T(), jobs, 2)

	// Test list jobs by queue
	req, err = http.NewRequest("GET", "/api/v1/jobs?queue=default", nil)
	require.NoError(suite.T(), err)

	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	jobs = response["jobs"]
	assert.Len(suite.T(), jobs, 1)
	assert.Equal(suite.T(), "default", jobs[0].Queue)
}

func (suite *HTTPIntegrationTestSuite) TestDeleteJob() {
	// First create a job directly in database
	payload := json.RawMessage(`{"message": "delete job test"}`)
	job := domain.NewJob("test-job", "default", payload)

	jobRepo := database.NewPostgresJobRepository(suite.pool)
	err := jobRepo.CreateJob(suite.ctx, job)
	require.NoError(suite.T(), err)

	// Delete it via HTTP
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/%s", job.ID), nil)
	require.NoError(suite.T(), err)

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	// Verify it's deleted
	deletedJob, err := jobRepo.GetJob(suite.ctx, job.ID)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), deletedJob)
}

func TestHTTPIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPIntegrationTestSuite))
}