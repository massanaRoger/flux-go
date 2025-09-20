package app

import (
	"context"
	"flux/internal/domain"
)

type JobService interface {
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, queue string) ([]*domain.Job, error)
	ProcessJob(ctx context.Context, jobID string) error
}

type jobService struct {
	jobRepo domain.JobRepository
}

func NewJobService(jobRepo domain.JobRepository) JobService {
	return &jobService{
		jobRepo: jobRepo,
	}
}

func (s *jobService) CreateJob(ctx context.Context, job *domain.Job) error {
	return s.jobRepo.CreateJob(ctx, job)
}

func (s *jobService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	return s.jobRepo.GetJob(ctx, id)
}

func (s *jobService) DeleteJob(ctx context.Context, id string) error {
	return s.jobRepo.DeleteJob(ctx, id)
}

func (s *jobService) ListJobs(ctx context.Context, queue string) ([]*domain.Job, error) {
	const defaultLimit = 100
	return s.jobRepo.ListJobs(ctx, queue, "", defaultLimit)
}

func (s *jobService) ProcessJob(ctx context.Context, jobID string) error {
	job, err := s.jobRepo.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return domain.ErrJobNotFound
	}

	job.Status = domain.JobStatusRunning
	job.Attempts++

	err = s.jobRepo.UpdateJob(ctx, job)
	if err != nil {
		return err
	}

	return nil
}