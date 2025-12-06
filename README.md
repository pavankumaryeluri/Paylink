# PayLink

**Multi-Provider Payment Gateway Backend**

PayLink is a production-grade payment orchestration service that provides a unified API for creating payment checkouts across multiple payment providers. Built with Go for high performance and reliability.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

PayLink abstracts the complexity of integrating multiple payment providers into a single, consistent API. Whether you are processing payments through Midtrans, Xendit, or Stripe, PayLink handles the provider-specific logic, webhook verification, and transaction management.

### Key Benefits

| Benefit | Description |
|---------|-------------|
| Provider Agnostic | Single API interface for all payment providers |
| Webhook Management | Automatic signature verification and idempotent processing |
| Horizontal Scaling | Stateless design supports multiple instances |
| Observable | Structured logging and metrics-ready architecture |
| Secure | Constant-time signature verification, no secrets in code |

---

## Features

### Core Capabilities

- **Unified Checkout API** - Create payment sessions with any supported provider
- **Webhook Processing** - Secure, idempotent webhook handling with async job processing
- **Transaction Management** - Query transaction status across providers
- **Background Workers** - Asynchronous job processing for webhooks and reconciliation

### Supported Providers

| Provider | Status | Documentation |
|----------|--------|---------------|
| Midtrans | Supported | [Midtrans Docs](https://docs.midtrans.com/) |
| Xendit | Supported | [Xendit Docs](https://developers.xendit.co/) |
| Stripe | Planned | [Stripe Docs](https://stripe.com/docs) |

---

## Architecture

```
                                    Client Request
                                          |
                                          v
+-------------------------------------------------------------------------+
|                            PayLink API Server                           |
|                                                                         |
|   +-------------+    +-------------+    +-------------+                 |
|   |   Routes    | -> | Middleware  | -> |  Handlers   |                 |
|   +-------------+    +-------------+    +-------------+                 |
|                              |                                          |
|                              v                                          |
|   +-------------------------------------------------------------+       |
|   |                     Adapter Layer                           |       |
|   |   +-----------+   +-----------+   +-----------+             |       |
|   |   | Midtrans  |   |  Xendit   |   |  Stripe   |             |       |
|   |   +-----------+   +-----------+   +-----------+             |       |
|   +-------------------------------------------------------------+       |
+-------------------------------------------------------------------------+
         |                    |                         |
         v                    v                         v
   +-----------+        +-----------+        +-------------------+
   | PostgreSQL|        |   Redis   |        | Payment Providers |
   +-----------+        +-----------+        +-------------------+
```

### Components

| Component | Purpose |
|-----------|---------|
| API Server | HTTP request handling and routing |
| Adapters | Provider-specific payment logic |
| Worker | Async webhook and reconciliation processing |
| PostgreSQL | Transaction and webhook storage |
| Redis | Job queue and idempotency keys |

For detailed architecture information, see [docs/architecture.md](docs/architecture.md).

---

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Git

### Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/vibeswithkk/Paylink.git
cd Paylink
```

2. **Configure environment**

```bash
cp .env.example .env
# Edit .env with your provider credentials
```

3. **Start services**

```bash
docker compose -f infra/docker-compose.yml up --build
```

4. **Verify installation**

```bash
curl http://localhost:8080/health
# Expected: {"status":"ok"}
```

---

## Configuration

PayLink is configured via environment variables. Copy `.env.example` to `.env` and update the values.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `paylink` |
| `DB_PASSWORD` | PostgreSQL password | `secret` |
| `DB_NAME` | PostgreSQL database | `paylink` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `MIDTRANS_SERVER_KEY` | Midtrans server key | - |
| `XENDIT_API_KEY` | Xendit API key | - |
| `STRIPE_SECRET_KEY` | Stripe secret key | - |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | - |

---

## API Reference

### Endpoints

#### Create Checkout

Creates a new payment checkout session.

```
POST /v1/checkout
Content-Type: application/json
```

**Request Body**

```json
{
    "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 150000,
    "currency": "IDR",
    "order_id": "ORDER-123",
    "provider_preference": "midtrans"
}
```

**Response**

```json
{
    "checkout_url": "https://app.sandbox.midtrans.com/snap/v3/redirection/...",
    "provider_tx_id": "snap_ORDER-123_150000"
}
```

#### Receive Webhook

Receives and processes webhooks from payment providers.

```
POST /v1/webhook/{provider}
```

#### Get Transaction

Retrieves transaction status.

```
GET /v1/tx/{id}
```

#### Health Check

Returns service health status.

```
GET /health
```

For complete API documentation, see [docs/api.yaml](docs/api.yaml).

---

## Development

### Project Structure

```
paylink/
├── cmd/
│   ├── server/          # API server entry point
│   └── worker/          # Background worker entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── adapters/        # Provider implementations
│   ├── config/          # Configuration
│   ├── crypto/          # Cryptographic utilities
│   ├── db/              # Database layer
│   ├── jobs/            # Job queue
│   ├── models/          # Data models
│   └── util/            # Utilities
├── crypto_cpp/          # C++ crypto module (optional)
├── infra/               # Docker and infrastructure
├── migrations/          # Database migrations
└── docs/                # Documentation
```

### Building from Source

```bash
# Build API server
go build -o bin/server ./cmd/server

# Build worker
go build -o bin/worker ./cmd/worker
```

### Adding a New Provider

1. Create adapter directory: `internal/adapters/{provider}/`
2. Implement `ProviderAdapter` interface
3. Register provider in `internal/adapters/registry.go`

---

## Testing

### Run Unit Tests

```bash
# Using Docker (recommended)
docker build -f infra/Dockerfile.test -t paylink-test .
docker run --rm paylink-test

# Using Go directly
go test -v ./...
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require running services:

```bash
# Start services
docker compose -f infra/docker-compose.yml up -d db redis

# Run integration tests
go test -v -tags=integration ./...

# Cleanup
docker compose -f infra/docker-compose.yml down
```

### Webhook Testing with Sandbox

1. Start PayLink locally
2. Expose local server via ngrok: `ngrok http 8080`
3. Configure webhook URL in provider sandbox dashboard
4. Trigger test payment in sandbox
5. Verify webhook received in PayLink logs

---

## Operations

### Metrics

PayLink exposes Prometheus-compatible metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

**Available Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `paylink_uptime_seconds` | Gauge | Time since server start |
| `paylink_requests_total` | Counter | Total requests by status |
| `paylink_checkouts_total` | Counter | Total checkouts created |
| `paylink_checkouts_by_provider` | Counter | Checkouts by provider |
| `paylink_webhooks_total` | Counter | Webhooks by status |
| `paylink_latency_avg_ms` | Gauge | Average request latency |

### Prometheus Configuration

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'paylink'
    static_configs:
      - targets: ['paylink:8080']
```

### Logging

PayLink uses structured JSON logging:

```json
{
  "time": "2024-12-06T10:00:00Z",
  "level": "INFO",
  "msg": "Checkout created",
  "provider": "midtrans",
  "order_id": "ORDER-123",
  "amount": 150000
}
```

**Log Levels:**
- `INFO` - Normal operations
- `WARN` - Recoverable issues (invalid signatures, validation failures)
- `ERROR` - Failures requiring attention

### Health Checks

```bash
# Liveness probe
curl http://localhost:8080/health

# Expected response
{"status":"ok"}
```

### Error Handling

See [docs/error_handling.md](docs/error_handling.md) for:
- Retry logic configuration
- Fallback behavior
- Dead letter queue handling

### Security

See [docs/security.md](docs/security.md) for:
- Webhook signature verification
- Input validation
- OWASP compliance

---

## Deployment

### Docker Compose (Development)

```bash
docker compose -f infra/docker-compose.yml up --build
```

### Docker Compose (Production)

```bash
docker compose -f infra/docker-compose.yml up -d
```

### Kubernetes

Helm charts and Kubernetes manifests are planned for future releases.

---

## Security

### Webhook Verification

All incoming webhooks are verified using provider-specific signature algorithms:

| Provider | Algorithm |
|----------|-----------|
| Midtrans | SHA512(order_id + status_code + gross_amount + ServerKey) |
| Xendit | x-callback-token header verification |
| Stripe | HMAC-SHA256 with webhook secret |

### Best Practices

- Store credentials in environment variables or secret managers
- Use HTTPS for all external communications
- Enable rate limiting in production
- Rotate API keys periodically
- Review audit logs regularly

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Support

For issues and feature requests, please open a GitHub issue.

---

*PayLink is maintained by the development team.*