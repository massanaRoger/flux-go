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
	// Enqueue adds a message to the specified queue for processing
	Enqueue(ctx context.Context, queue string, message *QueueMessage) error
	// Dequeue retrieves the next available message from any of the specified queues.
	// Blocks until a message is available or the timeout is reached.
	// Returns nil, nil if timeout expires with no messages available.
	Dequeue(ctx context.Context, queues []string, timeout time.Duration) (*QueueMessage, error)
	// Ack acknowledges successful processing of a message, removing it from the queue
	Ack(ctx context.Context, message *QueueMessage) error
	// Nack negatively acknowledges a message, returning it to the queue for retry
	Nack(ctx context.Context, message *QueueMessage) error
	// Close gracefully shuts down the queue broker connection
	Close() error
}