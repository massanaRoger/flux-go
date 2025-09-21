package app

import (
	"context"
	"flux/internal/domain"
	"log"
	"time"
)

type SchedulerService struct {
	jobRepo     domain.JobRepository
	queueBroker domain.QueueBroker
}

func NewSchedulerService(jobRepo domain.JobRepository, queueBroker domain.QueueBroker) *SchedulerService {
	return &SchedulerService{
		jobRepo:     jobRepo,
		queueBroker: queueBroker,
	}
}

func (s *SchedulerService) SchedulePendingJobs(ctx context.Context) error {
	const limit = 100
	jobs, err := s.jobRepo.GetJobsReadyToRun(ctx, limit)
	if err != nil {
		return err
	}

	now := time.Now()
	scheduled := 0

	for _, job := range jobs {
		job.Status = domain.JobStatusPending
		job.UpdatedAt = now

		if err := s.jobRepo.UpdateJob(ctx, job); err != nil {
			log.Printf("Failed to update job %s: %v", job.ID, err)
			continue
		}

		message := &domain.QueueMessage{
			ID:       job.ID,
			Queue:    job.Queue,
			Payload:  job.Payload,
			Priority: job.Priority,
		}

		if err := s.queueBroker.Enqueue(ctx, job.Queue, message); err != nil {
			log.Printf("Failed to enqueue job %s: %v", job.ID, err)
			job.Status = domain.JobStatusScheduled
			s.jobRepo.UpdateJob(ctx, job)
			continue
		}

		scheduled++
	}

	if scheduled > 0 {
		log.Printf("Scheduled %d jobs for execution", scheduled)
	}

	return nil
}

func (s *SchedulerService) ReclaimTimedOutJobs(ctx context.Context) error {
	const limit = 100
	jobs, err := s.jobRepo.GetTimedOutJobs(ctx, limit)
	if err != nil {
		return err
	}

	reclaimed := 0
	now := time.Now()

	for _, job := range jobs {
		if job.Attempts >= job.MaxRetries {
			job.Status = domain.JobStatusFailed
			job.UpdatedAt = now

			if err := s.jobRepo.UpdateJob(ctx, job); err != nil {
				log.Printf("Failed to mark job %s as failed: %v", job.ID, err)
			}
			continue
		}

		backoffDelay := s.calculateBackoffDelay(job.Attempts, job.BackoffPolicy)

		job.Status = domain.JobStatusRetrying
		job.ScheduleAt = now.Add(backoffDelay)
		job.UpdatedAt = now

		if err := s.jobRepo.UpdateJob(ctx, job); err != nil {
			log.Printf("Failed to schedule job %s for retry: %v", job.ID, err)
			continue
		}

		reclaimed++
	}

	if reclaimed > 0 {
		log.Printf("Reclaimed %d timed-out jobs", reclaimed)
	}

	return nil
}

func (s *SchedulerService) ProcessCronJobs(ctx context.Context) error {
	log.Println("Cron job processing not yet implemented")
	return nil
}

func (s *SchedulerService) calculateBackoffDelay(attempts int, policy string) time.Duration {
	switch policy {
	case "linear":
		return time.Duration(attempts) * time.Minute
	case "exp", "exponential":
		delay := time.Duration(1<<uint(attempts)) * time.Minute
		if delay > time.Hour {
			delay = time.Hour
		}
		return delay
	default:
		return time.Minute
	}
}
