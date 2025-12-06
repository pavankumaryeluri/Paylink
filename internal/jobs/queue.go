package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/util"
)

const (
	WebhookQueueKey = "queue:webhooks"
	MaxRetries      = 3
	RetryDelay      = 5 * time.Second
)

// WebhookJob represents a webhook processing job
type WebhookJob struct {
	Provider  string          `json:"provider"`
	Payload   json.RawMessage `json:"payload"`
	Retries   int             `json:"retries"`
	CreatedAt time.Time       `json:"created_at"`
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
		Provider:  provider,
		Payload:   payload,
		Retries:   0,
		CreatedAt: time.Now(),
	}
	data, err := json.Marshal(job)
	if err != nil {
		util.Logger.Error("Failed to marshal webhook job", "error", err)
		return err
	}

	err = e.Redis.LPush(ctx, WebhookQueueKey, data)
	if err != nil {
		util.Logger.Error("Failed to enqueue webhook", "error", err)
		return err
	}

	util.Logger.Info("Webhook job enqueued",
		"provider", provider,
		"queue", WebhookQueueKey)

	return nil
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
	util.Logger.Info("Worker started processing jobs")

	for {
		select {
		case <-ctx.Done():
			util.Logger.Info("Worker shutting down gracefully")
			return
		default:
		}

		res, err := w.Redis.BLPop(ctx, 5*time.Second, WebhookQueueKey)
		if err != nil || len(res) < 2 {
			continue
		}

		var job WebhookJob
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			util.Logger.Error("Failed to unmarshal job", "error", err)
			continue
		}

		w.handleJobWithRetry(ctx, &job)
	}
}

func (w *Worker) handleJobWithRetry(ctx context.Context, job *WebhookJob) {
	err := w.handleJob(ctx, job)
	if err != nil {
		if job.Retries < MaxRetries {
			job.Retries++
			util.Logger.Warn("Job failed, scheduling retry",
				"provider", job.Provider,
				"retries", job.Retries,
				"max_retries", MaxRetries,
				"error", err)

			// Re-enqueue with delay
			time.Sleep(RetryDelay)
			data, _ := json.Marshal(job)
			w.Redis.LPush(ctx, WebhookQueueKey, data)
		} else {
			util.Logger.Error("Job failed after max retries, moving to DLQ",
				"provider", job.Provider,
				"retries", job.Retries,
				"error", err)
			// TODO: Move to Dead Letter Queue
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job *WebhookJob) error {
	// Check context before processing
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled")
	default:
	}

	util.Logger.Info("Processing webhook job",
		"provider", job.Provider,
		"retries", job.Retries,
		"age_seconds", time.Since(job.CreatedAt).Seconds())

	// Simulate processing
	// In production: update transaction status, trigger notifications, etc.
	time.Sleep(100 * time.Millisecond)

	util.Logger.Info("Webhook job completed",
		"provider", job.Provider)

	return nil
}

// ReconciliationJob represents a reconciliation task
type ReconciliationJob struct {
	Provider       string    `json:"provider"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	TransactionIDs []string  `json:"transaction_ids,omitempty"`
}

// RunReconciliation performs reconciliation for a time range
func (w *Worker) RunReconciliation(ctx context.Context, job *ReconciliationJob) error {
	util.Logger.Info("Starting reconciliation",
		"provider", job.Provider,
		"start", job.StartTime,
		"end", job.EndTime)

	// In production:
	// 1. Fetch transactions from DB in PENDING state
	// 2. Query provider API for actual status
	// 3. Update local DB if status changed
	// 4. Generate discrepancy report

	util.Logger.Info("Reconciliation completed",
		"provider", job.Provider)

	return nil
}
