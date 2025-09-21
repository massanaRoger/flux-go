package app

import (
	"context"
	"log"
	"time"
)

type SchedulerRunner struct {
	service *SchedulerService
}

func NewSchedulerRunner(service *SchedulerService) *SchedulerRunner {
	return &SchedulerRunner{
		service: service,
	}
}

func (r *SchedulerRunner) Start(ctx context.Context) error {
	log.Println("Starting scheduler...")

	fastTicker := time.NewTicker(2 * time.Second)
	slowTicker := time.NewTicker(60 * time.Second)

	defer fastTicker.Stop()
	defer slowTicker.Stop()

	go r.fastTick(ctx, fastTicker)
	go r.slowTick(ctx, slowTicker)

	<-ctx.Done()
	log.Println("Scheduler shutting down...")
	return nil
}

func (r *SchedulerRunner) fastTick(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.service.SchedulePendingJobs(ctx); err != nil {
				log.Printf("Error scheduling pending jobs: %v", err)
			}

			if err := r.service.ReclaimTimedOutJobs(ctx); err != nil {
				log.Printf("Error reclaiming timed-out jobs: %v", err)
			}
		}
	}
}

func (r *SchedulerRunner) slowTick(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.service.ProcessCronJobs(ctx); err != nil {
				log.Printf("Error processing cron jobs: %v", err)
			}
		}
	}
}