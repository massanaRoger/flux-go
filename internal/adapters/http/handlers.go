package http

import (
	"flux/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService ports.JobService
}

func NewJobHandler(jobService ports.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusCreated, gin.H{"message": "Job created"})
}

func (h *JobHandler) GetJob(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Get job"})
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Job deleted"})
}

func (h *JobHandler) ListJobs(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"jobs": []interface{}{}})
}