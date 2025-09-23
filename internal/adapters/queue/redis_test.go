package queue

import (
	"context"
	"flux/internal/domain"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (testcontainers.Container, *redis.Client) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
		DB:   0,
	})

	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	return redisContainer, client
}

func TestRedisQueueBroker_EnqueueDequeue(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	message := &domain.QueueMessage{
		ID:       "test-job-1",
		Queue:    "test-queue",
		Payload:  []byte(`{"type":"test","data":"hello"}`),
		Priority: 0,
	}

	err := broker.Enqueue(ctx, "test-queue", message)
	require.NoError(t, err)

	dequeuedMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, dequeuedMessage)

	assert.Equal(t, message.ID, dequeuedMessage.ID)
	assert.Equal(t, message.Queue, dequeuedMessage.Queue)
	assert.Equal(t, message.Payload, dequeuedMessage.Payload)
	assert.Equal(t, message.Priority, dequeuedMessage.Priority)
}

func TestRedisQueueBroker_PriorityQueue(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	lowPriorityMessage := &domain.QueueMessage{
		ID:       "low-priority",
		Queue:    "test-queue",
		Payload:  []byte(`{"priority":"low"}`),
		Priority: 1,
	}

	highPriorityMessage := &domain.QueueMessage{
		ID:       "high-priority",
		Queue:    "test-queue",
		Payload:  []byte(`{"priority":"high"}`),
		Priority: 10,
	}

	err := broker.Enqueue(ctx, "test-queue", lowPriorityMessage)
	require.NoError(t, err)

	err = broker.Enqueue(ctx, "test-queue", highPriorityMessage)
	require.NoError(t, err)

	firstMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, firstMessage)

	assert.Equal(t, "high-priority", firstMessage.ID)
	assert.Equal(t, 10, firstMessage.Priority)

	secondMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, secondMessage)

	assert.Equal(t, "low-priority", secondMessage.ID)
	assert.Equal(t, 1, secondMessage.Priority)
}

func TestRedisQueueBroker_AckNack(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	message := &domain.QueueMessage{
		ID:       "test-job-ack",
		Queue:    "test-queue",
		Payload:  []byte(`{"type":"test"}`),
		Priority: 0,
	}

	err := broker.Enqueue(ctx, "test-queue", message)
	require.NoError(t, err)

	dequeuedMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, dequeuedMessage)

	processingKey := "processing:test-queue:test-job-ack"
	exists := client.Exists(ctx, processingKey).Val()
	assert.Equal(t, int64(1), exists)

	err = broker.Ack(ctx, dequeuedMessage)
	require.NoError(t, err)

	exists = client.Exists(ctx, processingKey).Val()
	assert.Equal(t, int64(0), exists)
}

func TestRedisQueueBroker_Nack(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	message := &domain.QueueMessage{
		ID:       "test-job-nack",
		Queue:    "test-queue",
		Payload:  []byte(`{"type":"test"}`),
		Priority: 0,
	}

	err := broker.Enqueue(ctx, "test-queue", message)
	require.NoError(t, err)

	dequeuedMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, dequeuedMessage)

	err = broker.Nack(ctx, dequeuedMessage)
	require.NoError(t, err)

	requeuedMessage, err := broker.Dequeue(ctx, []string{"test-queue"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, requeuedMessage)

	assert.Equal(t, message.ID, requeuedMessage.ID)
	assert.Equal(t, message.Queue, requeuedMessage.Queue)
}

func TestRedisQueueBroker_DequeueTimeout(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	start := time.Now()
	message, err := broker.Dequeue(ctx, []string{"empty-queue"}, 100*time.Millisecond)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Nil(t, message)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

func TestRedisQueueBroker_MultipleQueues(t *testing.T) {
	container, client := setupRedisContainer(t)
	defer container.Terminate(context.Background())

	broker := &RedisQueueBroker{client: client}
	ctx := context.Background()

	message1 := &domain.QueueMessage{
		ID:       "queue1-job",
		Queue:    "queue1",
		Payload:  []byte(`{"queue":"1"}`),
		Priority: 0,
	}

	message2 := &domain.QueueMessage{
		ID:       "queue2-job",
		Queue:    "queue2",
		Payload:  []byte(`{"queue":"2"}`),
		Priority: 0,
	}

	err := broker.Enqueue(ctx, "queue1", message1)
	require.NoError(t, err)

	err = broker.Enqueue(ctx, "queue2", message2)
	require.NoError(t, err)

	dequeuedMessage, err := broker.Dequeue(ctx, []string{"queue1", "queue2"}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, dequeuedMessage)

	assert.Contains(t, []string{"queue1-job", "queue2-job"}, dequeuedMessage.ID)
}

func TestNewRedisQueueBroker(t *testing.T) {
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "testpass")

	broker := NewRedisQueueBroker()
	require.NotNil(t, broker)

	redisClient := broker.(*RedisQueueBroker).client
	assert.Equal(t, "localhost:6379", redisClient.Options().Addr)
	assert.Equal(t, "testpass", redisClient.Options().Password)

	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("REDIS_PASSWORD")
}
