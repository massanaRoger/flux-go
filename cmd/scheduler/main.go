package main

import (
	"context"
	"flux/internal/adapters/database"
	"flux/internal/adapters/queue"
	"flux/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Flux Scheduler starting...")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbURL := getEnv("DATABASE_URL", "postgres://flux:flux123@localhost:5432/flux?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	jobRepo := database.NewPostgresJobRepository(db)
	queueBroker := queue.NewRedisQueueBrokerWithClient(redisClient)

	schedulerService := app.NewSchedulerService(jobRepo, queueBroker)
	schedulerRunner := app.NewSchedulerRunner(schedulerService)

	log.Println("Scheduler started successfully")
	if err := schedulerRunner.Start(ctx); err != nil {
		log.Printf("Scheduler error: %v", err)
	}

	log.Println("Scheduler stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
