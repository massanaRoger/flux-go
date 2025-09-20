package http

import (
	"encoding/json"
	"flux/internal/app"
	"flux/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService app.JobService
}

type CreateJobRequest struct {
	Type     string          `json:"type" binding:"required"`
	Queue    string          `json:"queue" binding:"required"`
	Payload  json.RawMessage `json:"payload" binding:"required"`
	Priority int             `json:"priority"`
}

func NewJobHandler(jobService app.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := domain.NewJob(req.Type, req.Queue, req.Payload)
	job.Priority = req.Priority

	if err := h.jobService.CreateJob(c, job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job id is required"})
		return
	}

	job, err := h.jobService.GetJob(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job id is required"})
		return
	}

	if err := h.jobService.DeleteJob(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted"})
}

func (h *JobHandler) ListJobs(c *gin.Context) {
	queue := c.Query("queue")

	jobs, err := h.jobService.ListJobs(c, queue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}
