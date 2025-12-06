# Error Handling, Retry Logic, and Fallback Behavior

This document describes how PayLink handles errors, retries, and fallback scenarios.

## Table of Contents

1. [Error Categories](#error-categories)
2. [Retry Logic](#retry-logic)
3. [Fallback Behavior](#fallback-behavior)
4. [Webhook Processing](#webhook-processing)
5. [Reconciliation](#reconciliation)
6. [Dead Letter Queue](#dead-letter-queue)

---

## Error Categories

PayLink classifies errors into the following categories:

### Retriable Errors

| Error Type | Description | Action |
|------------|-------------|--------|
| Network Timeout | Provider API timeout | Retry with exponential backoff |
| 5xx Response | Provider server error | Retry up to max attempts |
| Connection Refused | Provider unavailable | Retry with delay |
| Rate Limited | Too many requests | Retry after rate limit window |

### Non-Retriable Errors

| Error Type | Description | Action |
|------------|-------------|--------|
| Invalid Request | Malformed request data | Return 400 immediately |
| Authentication Failed | Invalid credentials | Return 401 |
| Invalid Signature | Webhook signature mismatch | Return 401 |
| Resource Not Found | Transaction/Invoice not found | Return 404 |

---

## Retry Logic

### Configuration

```
MaxRetries: 3
RetryDelay: 5 seconds (initial)
MaxRetryDelay: 60 seconds
BackoffMultiplier: 2.0
```

### Retry Flow

```
Request
   |
   v
+------------------+
|  Execute Call    |
+------------------+
   |
   v
+------------------+
|  Check Response  |
+------------------+
   |
   +-- Success --> Return Result
   |
   +-- Retriable Error
   |      |
   |      v
   |   +------------------+
   |   | Retries < Max?   |
   |   +------------------+
   |      |
   |      +-- Yes --> Wait (delay * backoff^attempt) --> Retry
   |      |
   |      +-- No --> Move to DLQ
   |
   +-- Non-Retriable --> Return Error Immediately
```

### Example: Webhook Processing with Retry

```go
func (w *Worker) handleJobWithRetry(ctx context.Context, job *WebhookJob) {
    err := w.handleJob(ctx, job)
    if err != nil {
        if job.Retries < MaxRetries {
            job.Retries++
            time.Sleep(RetryDelay * time.Duration(math.Pow(2, float64(job.Retries-1))))
            w.Redis.LPush(ctx, WebhookQueueKey, job)
        } else {
            // Move to Dead Letter Queue
            w.moveToDeadLetterQueue(job)
        }
    }
}
```

---

## Fallback Behavior

### Provider Fallback

If a primary provider fails, PayLink can optionally fallback to a secondary provider:

```
Primary Provider (Midtrans)
   |
   +-- Success --> Return Result
   |
   +-- Failure after retries
          |
          v
   Secondary Provider (Xendit)
          |
          +-- Success --> Return Result
          |
          +-- Failure --> Return Error to Client
```

**Note**: Provider fallback is not enabled by default. It must be explicitly configured per merchant.

### Crypto Fallback

PayLink includes both CGO-based and pure-Go crypto implementations:

| Environment | Implementation |
|-------------|----------------|
| Production with C++ lib | CGO wrapper (faster) |
| Development without C++ | Pure Go (portable) |

The selection is automatic based on build tags.

---

## Webhook Processing

### Flow

```
1. Provider sends webhook
2. PayLink receives at /v1/webhook/{provider}
3. Verify signature (provider-specific)
4. If valid: Enqueue to Redis for async processing
5. Return 200 OK immediately (within 5 seconds)
6. Worker picks up job
7. Process webhook (update transaction status)
8. On failure: Retry up to 3 times
9. After max retries: Move to DLQ
```

### Idempotency

Webhooks are processed idempotently using the event ID:

```
1. Extract event_id from webhook payload
2. Check if event_id exists in processed_events table
3. If exists: Return 200 (already processed)
4. If not: Process and store event_id
```

### Signature Verification

| Provider | Algorithm | Header/Field |
|----------|-----------|--------------|
| Midtrans | SHA512(order_id + status_code + gross_amount + ServerKey) | Body: `signature_key` |
| Xendit | Token comparison | Header: `x-callback-token` |

---

## Reconciliation

### Purpose

Reconciliation ensures local transaction status matches provider status.

### Schedule

```
- Frequency: Every 15 minutes (configurable)
- Scope: Transactions in PENDING status older than 30 minutes
```

### Flow

```
1. Query DB for pending transactions
2. For each transaction:
   a. Call provider API to get status
   b. Compare with local status
   c. If different: Update local status
   d. Log discrepancy
3. Generate reconciliation report
```

### Example Output

```
Reconciliation Report - 2024-12-06 10:00
-----------------------------------------
Provider: midtrans
Transactions checked: 150
Status changes: 12
  - PENDING -> PAID: 10
  - PENDING -> EXPIRED: 2
Discrepancies: 0
```

---

## Dead Letter Queue

### Purpose

Stores jobs that failed after all retry attempts for manual investigation.

### Structure

```sql
CREATE TABLE dead_letter_queue (
    id UUID PRIMARY KEY,
    queue_name TEXT NOT NULL,
    payload JSONB NOT NULL,
    error_message TEXT,
    retry_count INT,
    created_at TIMESTAMP,
    processed_at TIMESTAMP
);
```

### Operations

| Action | Description |
|--------|-------------|
| View | List DLQ entries with filters |
| Retry | Manually re-process a DLQ entry |
| Delete | Remove entry after investigation |
| Export | Export for external analysis |

---

## Error Response Format

All API errors follow this format:

```json
{
    "error": "Error message description",
    "code": "ERROR_CODE",
    "details": {
        "field": "Additional context"
    }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | Request validation failed |
| INVALID_SIGNATURE | 401 | Webhook signature invalid |
| UNAUTHORIZED | 401 | Missing or invalid API key |
| PROVIDER_NOT_FOUND | 404 | Unknown payment provider |
| TRANSACTION_NOT_FOUND | 404 | Transaction does not exist |
| PROVIDER_ERROR | 502 | Provider API returned error |
| SERVICE_UNAVAILABLE | 503 | System temporarily unavailable |

---

## Monitoring Error Rates

Errors are tracked via the `/metrics` endpoint:

```
paylink_requests_total{status="failed"} 42
paylink_webhooks_total{status="failed"} 7
```

### Alerting Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Error rate | > 1% | > 5% |
| Webhook failures | > 10/min | > 50/min |
| DLQ size | > 100 | > 500 |

---

*Document maintained by PayLink Engineering Team*
