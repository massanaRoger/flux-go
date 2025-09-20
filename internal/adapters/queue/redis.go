package queue

import (
	"context"
	"flux/internal/ports"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisQueueBroker struct {
	client *redis.Client
}

func NewRedisQueueBroker() ports.QueueBroker {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &RedisQueueBroker{client: rdb}
}

func (r *RedisQueueBroker) Enqueue(ctx context.Context, queue string, message *ports.QueueMessage) error {
	// TODO: Implement
	return nil
}

func (r *RedisQueueBroker) Dequeue(ctx context.Context, queues []string, timeout time.Duration) (*ports.QueueMessage, error) {
	// TODO: Implement
	return nil, nil
}

func (r *RedisQueueBroker) Ack(ctx context.Context, message *ports.QueueMessage) error {
	// TODO: Implement
	return nil
}

func (r *RedisQueueBroker) Nack(ctx context.Context, message *ports.QueueMessage) error {
	// TODO: Implement
	return nil
}

func (r *RedisQueueBroker) Close() error {
	return r.client.Close()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}