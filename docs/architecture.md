# PayLink Architecture Documentation

**Version:** 1.0.0  
**Last Updated:** December 6, 2024  
**Author:** Wahyu Ardiansyah

---

## Table of Contents
1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Component Diagram](#component-diagram)
4. [Directory Structure](#directory-structure)
5. [Core Components](#core-components)
6. [Data Flow](#data-flow)
7. [Database Schema](#database-schema)
8. [API Design](#api-design)
9. [Security Architecture](#security-architecture)
10. [Deployment Architecture](#deployment-architecture)

---

## Overview

PayLink is a **multi-provider payment gateway backend** designed for production-grade reliability. It provides a unified API for creating payment checkouts across multiple payment providers (Midtrans, Xendit, Stripe) while handling webhooks, idempotency, and reconciliation.

### Design Principles
- **Provider Agnostic**: Single API abstracts multiple payment providers
- **Idempotent**: Safe retry mechanisms for all state-changing operations
- **Observable**: Structured logging and metrics-ready architecture
- **Portable**: Minimal dependencies, pure Go implementation
- **Secure**: Constant-time signature verification, no secrets in code

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CLIENTS                                    │
│         (Mobile Apps, Web Apps, Backend Services)                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        PAYLINK API SERVER                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │   /v1/      │  │  Middleware │  │   Handlers  │                 │
│  │  checkout   │◄─┤  - Auth     │◄─┤  - Checkout │                 │
│  │  webhook    │  │  - Logging  │  │  - Webhook  │                 │
│  │  tx/{id}    │  │  - Recovery │  │  - Status   │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘                 │
│                           │                                         │
│                           ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    ADAPTER LAYER                             │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │   │
│  │  │ Midtrans │  │  Xendit  │  │  Stripe  │  │  (Future)│    │   │
│  │  │ Adapter  │  │ Adapter  │  │ Adapter  │  │ Adapters │    │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
         │                    │                         │
         ▼                    ▼                         ▼
┌──────────────┐    ┌──────────────┐    ┌─────────────────────────┐
│  PostgreSQL  │    │    Redis     │    │   Payment Providers     │
│  - Txns      │    │  - Queue     │    │  - Midtrans Sandbox     │
│  - Webhooks  │    │  - Idempot.  │    │  - Xendit Test          │
│  - Merchants │    │  - Rate Lim  │    │  - Stripe Test          │
└──────────────┘    └──────────────┘    └─────────────────────────┘
         ▲
         │
┌──────────────────────────────────────────────────────────────────┐
│                       WORKER PROCESS                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Webhook   │  │   Reconci-  │  │    Retry    │              │
│  │  Processor  │  │   liation   │  │   Handler   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└──────────────────────────────────────────────────────────────────┘
```

---

## Component Diagram

```
paylink/
├── cmd/                    # Application Entry Points
│   ├── server/main.go      # HTTP API Server
│   └── worker/main.go      # Background Job Worker
│
├── internal/               # Private Application Code
│   ├── api/                # HTTP Handlers & Routes
│   │   ├── handler.go      # Request handlers
│   │   └── handler_test.go # Unit tests
│   │
│   ├── adapters/           # Payment Provider Adapters
│   │   ├── interface.go    # ProviderAdapter interface
│   │   ├── registry.go     # Adapter factory
│   │   ├── midtrans/       # Midtrans implementation
│   │   └── xendit/         # Xendit implementation
│   │
│   ├── config/             # Configuration Management
│   │   └── config.go       # Env-based config loader
│   │
│   ├── crypto/             # Cryptographic Operations
│   │   ├── crypto.go       # HMAC-SHA256 implementation
│   │   └── crypto_test.go  # Crypto unit tests
│   │
│   ├── db/                 # Database Layer
│   │   └── db.go           # PostgreSQL + Redis connections
│   │
│   ├── jobs/               # Background Job Processing
│   │   └── queue.go        # Redis-backed job queue
│   │
│   ├── models/             # Data Models
│   │   └── transaction.go  # Transaction, Webhook, Merchant
│   │
│   └── util/               # Shared Utilities
│       └── logger.go       # Structured JSON logger
│
├── crypto_cpp/             # C++ Crypto Module (Optional)
│   ├── paycrypto.h         # C ABI header
│   ├── paycrypto.cpp       # HMAC implementation
│   └── CMakeLists.txt      # Build configuration
│
├── infra/                  # Infrastructure
│   ├── docker-compose.yml  # Service orchestration
│   ├── Dockerfile.server   # Server container
│   └── Dockerfile.worker   # Worker container
│
├── migrations/             # Database Migrations
│   └── 001_initial_schema.sql
│
└── docs/                   # Documentation
    ├── architecture.md     # This file
    └── api.yaml            # OpenAPI 3.0 spec
```

---

## Core Components

### 1. API Handler (`internal/api/handler.go`)

The HTTP handler manages all incoming requests using Go's standard `net/http` package.

```go
type Handler struct {
    Config   *config.Config
    DB       *db.DB
    Enqueuer *jobs.Enqueuer
}

// Routes returns the HTTP router
func (h *Handler) Routes() http.Handler
```

**Endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| POST | /v1/checkout | Create payment checkout |
| POST | /v1/webhook/{provider} | Receive provider webhooks |
| GET | /v1/tx/{id} | Get transaction status |
| GET | /health | Health check |

### 2. Provider Adapter (`internal/adapters/interface.go`)

Abstract interface for payment provider implementations:

```go
type ProviderAdapter interface {
    CreatePayment(ctx context.Context, tx *Transaction) (providerTxID, checkoutURL string, err error)
    VerifySignature(r *http.Request, body []byte) (eventID string, valid bool, err error)
    GetTransactionStatus(ctx context.Context, providerTxID string) (status string, err error)
}
```

**Implementations:**
- `midtrans.Adapter` - Midtrans Snap API
- `xendit.Adapter` - Xendit Invoice API
- (Future) `stripe.Adapter` - Stripe Checkout

### 3. Job Queue (`internal/jobs/queue.go`)

Redis-backed asynchronous job processing:

```go
type Enqueuer struct {
    Redis *db.RedisClient
}

func (e *Enqueuer) EnqueueWebhook(ctx context.Context, provider string, payload []byte) error

type Worker struct {
    Redis *db.RedisClient
}

func (w *Worker) Process(ctx context.Context)
```

### 4. Crypto Module (`internal/crypto/crypto.go`)

Secure cryptographic operations:

```go
// ComputeHMAC computes HMAC-SHA256
func ComputeHMAC(key, data string) string

// VerifyHMAC verifies signature with constant-time comparison
func VerifyHMAC(key, data, signature string) bool
```

---

## Data Flow

### Checkout Flow

```
1. Client → POST /v1/checkout
2. Handler validates request
3. Handler selects adapter based on provider_preference
4. Adapter.CreatePayment() calls provider API
5. Transaction saved to PostgreSQL
6. Response with checkout_url returned to client
```

### Webhook Flow

```
1. Provider → POST /v1/webhook/{provider}
2. Handler reads body
3. Adapter.VerifySignature() validates HMAC
4. If valid, job enqueued to Redis
5. Return 200 OK immediately (fast response)
6. Worker picks up job asynchronously
7. Worker updates transaction status in DB
```

---

## Database Schema

### Tables

```sql
-- Merchants table
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    api_key_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    provider TEXT NOT NULL,
    provider_tx_id TEXT,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Webhook events table
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    event_id TEXT,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_event_id UNIQUE (provider, event_id)
);

-- Idempotency keys table
CREATE TABLE idempotency_keys (
    key TEXT PRIMARY KEY,
    response JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Indexes

```sql
CREATE UNIQUE INDEX idx_tx_provider_id ON transactions(provider, provider_tx_id);
```

---

## API Design

See `docs/api.yaml` for full OpenAPI 3.0 specification.

### Request/Response Examples

**POST /v1/checkout**
```json
// Request
{
    "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 150000,
    "currency": "IDR",
    "order_id": "ORDER-123",
    "provider_preference": "midtrans"
}

// Response
{
    "checkout_url": "https://app.sandbox.midtrans.com/snap/v3/redirection/...",
    "provider_tx_id": "snap_ORDER-123_150000"
}
```

---

## Security Architecture

### Authentication
- Merchant API key in `X-Api-Key` header
- API keys stored as bcrypt hashes in database

### Webhook Verification
- Each provider has specific signature verification
- **Midtrans**: SHA512(order_id + status_code + gross_amount + ServerKey)
- **Xendit**: `x-callback-token` header verification
- All comparisons use constant-time algorithms

### Secrets Management
- No hardcoded secrets in code
- Environment variables via `.env` file
- In production: use secret managers (Vault, AWS Secrets Manager)

---

## Deployment Architecture

### Docker Compose (Development)

```yaml
services:
  paylink:      # API Server (port 8080)
  worker:       # Background Worker
  db:           # PostgreSQL 16
  redis:        # Redis 7
```

### Production Recommendations

```
                    ┌─────────────┐
                    │   Load      │
                    │  Balancer   │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ PayLink  │      │ PayLink  │      │ PayLink  │
   │ Server 1 │      │ Server 2 │      │ Server 3 │
   └──────────┘      └──────────┘      └──────────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ Worker 1 │      │ Worker 2 │      │ Worker 3 │
   └──────────┘      └──────────┘      └──────────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
        ┌──────────┐              ┌──────────┐
        │ Postgres │              │  Redis   │
        │ Primary  │              │ Cluster  │
        └──────────┘              └──────────┘
```

---

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go 1.22 | Concurrency, performance, simple deployment |
| HTTP Framework | `net/http` | Zero dependencies, production-ready |
| Database | PostgreSQL | ACID, JSONB support, reliability |
| Queue | Redis | Speed, simplicity, pub/sub capability |
| Crypto | Pure Go | Portability, no CGO complexity |
| Container | Docker | Reproducible builds, easy deployment |

---

## Future Considerations

1. **Additional Providers**: Stripe, GoPay, OVO, DANA
2. **Rate Limiting**: Token bucket algorithm with Redis
3. **Metrics**: Prometheus `/metrics` endpoint
4. **Tracing**: OpenTelemetry integration
5. **Multi-tenancy**: Merchant isolation and quotas

---

*Document maintained by PayLink Engineering Team*
