package domain

import (
	"context"
	"time"
)

type QueueMessage struct {
	ID       string
	Queue    string
	Payload  []byte
	Priority int
}

type QueueBroker interface {
	Enqueue(ctx context.Context, queue string, message *QueueMessage) error
	Dequeue(ctx context.Context, queues []string, timeout time.Duration) (*QueueMessage, error)
	Ack(ctx context.Context, message *QueueMessage) error
	Nack(ctx context.Context, message *QueueMessage) error
	Close() error
}