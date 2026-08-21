package notify

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Topic     string            `json:"topic"`
	Payload   map[string]string `json:"payload"`
	CreatedAt time.Time         `json:"created_at"`
}

type MemoryNotifier struct {
	mu     sync.Mutex
	events []Event
}

func (n *MemoryNotifier) Notify(ctx context.Context, topic string, payload map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	clone := make(map[string]string, len(payload))
	for key, value := range payload {
		clone[key] = value
	}
	n.events = append(n.events, Event{Topic: topic, Payload: clone, CreatedAt: time.Now().UTC()})
	return nil
}

func (n *MemoryNotifier) Events() []Event {
	n.mu.Lock()
	defer n.mu.Unlock()
	result := make([]Event, len(n.events))
	copy(result, n.events)
	return result
}
