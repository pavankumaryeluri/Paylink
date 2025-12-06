PayLink adalah backend modular untuk mengelola proses pembayaran lintas provider (Midtrans, Xendit, Stripe, dsb).
Dirancang seperti sistem produksi di perusahaan fintech: aman, scalable, observability-ready, dan mudah diintegrasikan.

Proyek ini mencakup:

Multi-provider Payment Adapter Architecture

Secure Webhook Handler (signature validation + idempotency)

C++ Crypto Shared Library (diakses via cgo) untuk operasi signing/verification berperforma tinggi

Background Workers untuk reconciliation & event processing

PostgreSQL untuk transaksi

Redis untuk idempotency keys & rate limiting

OpenAPI spec, Postman, dan struktur folder enterprise

Semua berjalan tanpa VPS, cukup menggunakan Docker Compose di Linux Ubuntu.

🧱 Architecture
      [ Client ]
          |
          v
 ┌───────────────────────────┐
 │       PayLink API         │  <-- Go HTTP Server (REST)
 │  - Checkout creation      │
 │  - Provider adapters      │
 │  - Webhook router         │
 │  - Auth + rate limit      │
 └───────────┬──────────────┘
             |
             | cgo (batched calls)
             v
      [ C++ Crypto Module ]
        - HMAC / signing
        - Heavy crypto ops

 ┌───────────────┐   ┌────────────────┐
 │   Postgres     │   │     Redis      │
 │ transactions   │   │ idempotency    │
 │ webhook logs   │   │ rate limiting  │
 └───────────────┘   └────────────────┘

            ┌─────────────────────────┐
            │  Background Workers     │
            │ - webhook processor     │
            │ - reconciliation jobs   │
            └─────────────────────────┘


🔧 Tech Stack

Go 1.22+ (HTTP server, adapters, worker)

C++17 (native crypto; compiled as shared library)

PostgreSQL 16

Redis 7

Docker & Docker Compose

OpenTelemetry + Prometheus (optional)

Ubuntu Linux development environment

📂 Directory Structure
paylink/
├─ cmd/
│  └─ server/
│       └─ main.go
├─ internal/
│  ├─ api/         # HTTP routes, handlers, validators
│  ├─ adapters/    # midtrans/, xendit/, stripe/
│  ├─ crypto/      # cgo wrapper around C++ lib
│  ├─ db/          # SQL queries, migrations
│  ├─ jobs/        # workers & queue consumers
│  ├─ webhook/     # webhook validators + router
│  └─ util/        # logging, metrics, idempotency
├─ crypto_cpp/
│  ├─ src/
│  ├─ include/
│  └─ CMakeLists.txt
├─ infra/
│  ├─ docker-compose.yml
│  ├─ Dockerfile.server
│  └─ Dockerfile.worker
├─ migrations/
│  └─ *.sql
├─ docs/
│  ├─ api.yaml         # OpenAPI 3.0
│  ├─ architecture.md
│  └─ postman_collection.json
└─ README.md

📝 Features
✔ Multi-Provider Payment Integration

PayLink menyediakan interface adapter yang memudahkan integrasi provider baru tanpa mengubah core system.
Provider yang tersedia:

Midtrans (sandbox)

Xendit (test mode)

Stripe (test mode)

Easily pluggable adapter pattern

✔ Secure Webhook System

Signature verification (HMAC/SHA256)

Constant-time comparison (anti-timing attack)

Error-safe parsing

Idempotency key system (Redis + PostgreSQL)

Duplicate event handling

Asynchronous processing

✔ C++ Native Crypto Module

Untuk operasi berat (signing, verification), PayLink menggunakan modul C++ (dipanggil via cgo) sebagai shared library:

libpaycrypto.so


Benefit:

High-performance native operations

Suitable for large payload signing

Minimizes Go-side CPU cost

✔ Background Workers

Webhook processing

Retry jobs

Reconciliation (pull provider status)

✔ Observability

/metrics endpoint (Prometheus)

OpenTelemetry traces (Jaeger)

Structured logging (JSON)

🔌 API Endpoints
POST /v1/checkout

Membuat transaksi dan mengembalikan checkout URL provider.

POST /v1/webhook/{provider}

Menerima callback dari provider payment.

GET /v1/tx/{id}

Mengecek status transaksi.

POST /v1/reconcile

Menjalankan reconciliation worker.

Dokumentasi lengkap: lihat docs/api.yaml.

🗄 Database Design
transactions
field	type
id	UUID
merchant_id	UUID
provider	TEXT
provider_tx_id	TEXT
amount	BIGINT
currency	TEXT
status	TEXT
metadata	JSONB
created_at	TIMESTAMP
webhook_events
field	type
id	UUID
provider	TEXT
event_id	TEXT UNIQUE
payload	JSONB
processed	BOOL
idempotency_keys

| key | TEXT PRIMARY KEY
| response_snapshot | JSONB

🐳 Running Locally (Ubuntu)
1. Clone project
git clone https://github.com/username/paylink.git
cd paylink

2. Create environment file
cp .env.example .env

3. Start everything
docker compose up --build

4. API available at:
http://localhost:8080

🔐 Security Highlights

Signature verification for all webhooks

HMAC constant-time comparison

Idempotent transaction updates

Sanitize & validate incoming payloads

API-key protected merchant endpoints

Rate-limiting (Redis token bucket)

🧪 Testing
Unit Tests
go test ./...

Integration Tests

Runs Postgres + Redis containers:

docker compose -f infra/docker-compose.test.yml up --build

E2E Webhook Testing
ngrok http 8080


Set URL di sandbox provider (Midtrans/Xendit/Stripe).

🛠 Build C++ Crypto Library
Build manually on Ubuntu:
cd crypto_cpp
mkdir build && cd build
cmake ..
make


Output:

libpaycrypto.so

🧩 Adding New Payment Providers

Tambahkan folder:

internal/adapters/<provider>/


Implement interface:

type ProviderAdapter interface {
   CreatePayment(...)
   VerifySignature(...)
   GetTransactionStatus(...)
}


Register provider di adapters/registry.go.

🚦 CI / CD

GitHub Actions pipeline meliputi:

lint (golangci-lint)

build Go server & worker

build C++ native module (CMake)

run unit tests

run integration tests (Docker)

produce Docker images

📜 License

MIT License.

🎯 Status

Active Development
Roadmap: Doku adapter, PayPal, multi-tenant merchant keys, OpenTelemetry tracing pipeline.

🙌 Credits

Arsitektur terinspirasi dari pola backend fintech modern (Stripe, Xendit, Go microservices), praktik industri cloud-native, dan referensi open-source dari GitHub ecosystem.