package queue

import (
	"context"
	"encoding/json"
	"flux/internal/domain"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisQueueBroker struct {
	client *redis.Client
}

func NewRedisQueueBroker() domain.QueueBroker {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &RedisQueueBroker{client: rdb}
}

func NewRedisQueueBrokerWithClient(client *redis.Client) domain.QueueBroker {
	return &RedisQueueBroker{client: client}
}

func (r *RedisQueueBroker) Enqueue(ctx context.Context, queue string, message *domain.QueueMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", queue)

	if message.Priority > 0 {
		priorityQueueKey := fmt.Sprintf("queue:%s:priority", queue)
		return r.client.ZAdd(ctx, priorityQueueKey, &redis.Z{
			Score:  float64(-message.Priority),
			Member: string(data),
		}).Err()
	}

	return r.client.LPush(ctx, queueKey, string(data)).Err()
}

func (r *RedisQueueBroker) Dequeue(ctx context.Context, queues []string, timeout time.Duration) (*domain.QueueMessage, error) {
	queueKeys := make([]string, 0, len(queues)*2)

	for _, queue := range queues {
		priorityQueueKey := fmt.Sprintf("queue:%s:priority", queue)
		queueKey := fmt.Sprintf("queue:%s", queue)
		queueKeys = append(queueKeys, priorityQueueKey, queueKey)
	}

	for _, queue := range queues {
		priorityQueueKey := fmt.Sprintf("queue:%s:priority", queue)
		result := r.client.ZPopMin(ctx, priorityQueueKey, 1)
		if result.Err() == nil && len(result.Val()) > 0 {
			data := result.Val()[0].Member.(string)
			var message domain.QueueMessage
			if err := json.Unmarshal([]byte(data), &message); err != nil {
				return nil, fmt.Errorf("failed to unmarshal priority message: %w", err)
			}

			processingKey := fmt.Sprintf("processing:%s:%s", queue, message.ID)
			r.client.Set(ctx, processingKey, data, 5*time.Minute)
			return &message, nil
		}
	}

	keys := make([]string, len(queues))
	for i, queue := range queues {
		keys[i] = fmt.Sprintf("queue:%s", queue)
	}

	result := r.client.BRPop(ctx, timeout, keys...)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil
		}
		return nil, result.Err()
	}

	if len(result.Val()) < 2 {
		return nil, nil
	}

	data := result.Val()[1]
	var message domain.QueueMessage
	if err := json.Unmarshal([]byte(data), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	processingKey := fmt.Sprintf("processing:%s:%s", message.Queue, message.ID)
	r.client.Set(ctx, processingKey, data, 5*time.Minute)

	return &message, nil
}

func (r *RedisQueueBroker) Ack(ctx context.Context, message *domain.QueueMessage) error {
	processingKey := fmt.Sprintf("processing:%s:%s", message.Queue, message.ID)
	return r.client.Del(ctx, processingKey).Err()
}

func (r *RedisQueueBroker) Nack(ctx context.Context, message *domain.QueueMessage) error {
	processingKey := fmt.Sprintf("processing:%s:%s", message.Queue, message.ID)

	data, err := r.client.Get(ctx, processingKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get processing message: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", message.Queue)
	if message.Priority > 0 {
		priorityQueueKey := fmt.Sprintf("queue:%s:priority", message.Queue)
		err = r.client.ZAdd(ctx, priorityQueueKey, &redis.Z{
			Score:  float64(-message.Priority),
			Member: data,
		}).Err()
	} else {
		err = r.client.LPush(ctx, queueKey, data).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to requeue message: %w", err)
	}

	return r.client.Del(ctx, processingKey).Err()
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
