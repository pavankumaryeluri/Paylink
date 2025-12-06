package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vibeswithkk/paylink/internal/db"
)

const WebhookQueueKey = "queue:webhooks"

// WebhookJob represents a webhook processing job
type WebhookJob struct {
	Provider string          `json:"provider"`
	Payload  json.RawMessage `json:"payload"`
}

// Enqueuer handles job enqueueing
type Enqueuer struct {
	Redis *db.RedisClient
}

// NewEnqueuer creates a new job enqueuer
func NewEnqueuer(r *db.RedisClient) *Enqueuer {
	return &Enqueuer{Redis: r}
}

// EnqueueWebhook adds a webhook job to the queue
func (e *Enqueuer) EnqueueWebhook(ctx context.Context, provider string, payload []byte) error {
	job := WebhookJob{
		Provider: provider,
		Payload:  payload,
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return e.Redis.LPush(ctx, WebhookQueueKey, data)
}

// Worker processes jobs from the queue
type Worker struct {
	Redis *db.RedisClient
}

// NewWorker creates a new job worker
func NewWorker(r *db.RedisClient) *Worker {
	return &Worker{Redis: r}
}

// Process continuously processes jobs from the queue
func (w *Worker) Process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Worker shutting down...")
			return
		default:
		}

		res, err := w.Redis.BLPop(ctx, 5*time.Second, WebhookQueueKey)
		if err != nil || len(res) < 2 {
			continue
		}

		var job WebhookJob
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			fmt.Printf("Failed to unmarshal job: %v\n", err)
			continue
		}

		w.handleJob(ctx, &job)
	}
}

func (w *Worker) handleJob(ctx context.Context, job *WebhookJob) {
	// Check context before processing
	select {
	case <-ctx.Done():
		fmt.Println("Job cancelled")
		return
	default:
	}

	fmt.Printf("Processing webhook for %s: %s\n", job.Provider, string(job.Payload))
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
}
