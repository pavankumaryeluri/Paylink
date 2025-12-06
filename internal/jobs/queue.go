package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const WebhookQueueKey = "queue:webhooks"

type WebhookJob struct {
	Provider string          `json:"provider"`
	Payload  json.RawMessage `json:"payload"`
}

type Enqueuer struct {
	Redis *redis.Client
}

func NewEnqueuer(r *redis.Client) *Enqueuer {
	return &Enqueuer{Redis: r}
}

func (e *Enqueuer) EnqueueWebhook(ctx context.Context, provider string, payload []byte) error {
	job := WebhookJob{
		Provider: provider,
		Payload:  payload,
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return e.Redis.LPush(ctx, WebhookQueueKey, data).Err()
}

type Worker struct {
	Redis *redis.Client
}

func NewWorker(r *redis.Client) *Worker {
	return &Worker{Redis: r}
}

func (w *Worker) Process(ctx context.Context) {
	for {
		// BLPOP with 5 second timeout
		res, err := w.Redis.BLPop(ctx, 5*time.Second, WebhookQueueKey).Result()
		if err != nil {
			if err != redis.Nil {
				// Log error (needs logger)
				// fmt.Println("Error popping:", err)
			}
			continue
		}

		// res[0] is key, res[1] is value
		var job WebhookJob
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			fmt.Printf("Failed to unmarshal job: %v\n", err)
			continue
		}

		w.handleJob(ctx, &job)
	}
}

func (w *Worker) handleJob(ctx context.Context, job *WebhookJob) {
	fmt.Printf("Processing webhook for %s: %s\n", job.Provider, string(job.Payload))
	// Here you would call the reconciliation or business logic updates
	// For now, simulate work
	time.Sleep(100 * time.Millisecond)
}
