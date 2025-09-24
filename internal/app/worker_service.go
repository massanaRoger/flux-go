package app

import (
	"context"
	"encoding/json"
	"errors"
	"flux/internal/domain"
	"fmt"
	"time"
)

type WorkerService struct {
	broker   domain.QueueBroker
	handlers map[string]domain.JobHandler
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewWorkerService(parent context.Context, queueBroker domain.QueueBroker) *WorkerService {
	ctx, cancel := context.WithCancel(parent)
	return &WorkerService{
		broker:   queueBroker,
		handlers: make(map[string]domain.JobHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *WorkerService) RegisterHandler(jobType string, handler domain.JobHandler) error {
	if jobType == "" {
		return errors.New("job type cannot be empty")
	}
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	s.handlers[jobType] = handler
	return nil
}

func (s *WorkerService) ProcessJobs(queues []string) error {
	if len(queues) == 0 {
		return errors.New("no queues specified for processing")
	}
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
			pollCtx, pollCancel := context.WithTimeout(s.ctx, 30*time.Second)
			message, err := s.broker.Dequeue(pollCtx, queues, 30*time.Second)
			pollCancel()

			if err != nil {
				//TODO: log the dequeue message
				// Only Nack if a message was actually received
				if message != nil {
					_ = s.broker.Nack(s.ctx, message)
				}
				continue
			}
			if message != nil {
				jobType, err := extractJobType(message.Payload)
				if err != nil {
					// Malformed payload; Nack so it can be dead-lettered/retried per broker policy.
					_ = s.broker.Nack(s.ctx, message)
					continue
				}
				handler, ok := s.handlers[jobType]
				if !ok {
					// Unknown job; Nack (or consider routing to a DLQ).
					_ = s.broker.Nack(s.ctx, message)
					continue
				}

				// Optional: per-job timeout/deadline to keep things snappy and shutdown-friendly.
				jobCtx, jobCancel := context.WithTimeout(s.ctx, 2*time.Minute)
				err = handler(jobCtx, message)
				jobCancel()

				if err != nil {
					// Handler failed; decide retry policy here.
					_ = s.broker.Nack(s.ctx, message)
					continue
				}

				_ = s.broker.Ack(s.ctx, message)
			}
		}
	}
}

func (s *WorkerService) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func extractJobType(payload []byte) (string, error) {
	var wrapper struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return "", fmt.Errorf("failed to parse job wrapper: %w", err)
	}
	if wrapper.Type == "" {
		return "", errors.New("job type is required")
	}

	return wrapper.Type, nil
}
