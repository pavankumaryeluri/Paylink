# How to Build an Enterprise Payment Gateway (The PayLink Way)

## Introduction
This tutorial dissects **PayLink**, a production-ready payment orchestration backend. You will learn how to architect a system that handles money safely, using Go, C++, and Docker.

## Prerequisites
- Linux (Ubuntu 20.04+)
- Go 1.22+
- Docker & Docker Compose
- Basic C++ knowledge

## Chapter 1: The Architecture
We don't just dump code; we design systems.
1. **The Core (Go):** Handles HTTP requests. It's fast, concurrent, and safe.
2. **The Muscle (C++):** Handles crypto. When you need to verify 10,000 webhook signatures/sec, Go's GC can get in the way. We use a C++ shared library.
3. **The Memory (Redis):** Handles idempotency. Replaying a $100 charge request shouldn't charge the user $200. Redis keys with expiration prevent this.

## Chapter 2: The C++ Interop (cgo)
Look at `crypto_cpp/paycrypto.cpp`. We expose `ComputeHMAC_SHA256` as an `extern "C"` function.
In Go (`internal/crypto/wrapper.go`), we use:
```go
// #cgo LDFLAGS: -lpaycrypto
import "C"
```
This allows us to call C++ code as if it were a Go function. *Warning:* Don't do this for everything. Use it for batch operations or heavy compute to offset the Context Switch cost.

## Chapter 3: The Adapter Pattern
We define a single interface:
```go
type ProviderAdapter interface {
    CreatePayment(...)
}
```
Midtrans, Xendit, and Stripe implementations satisfy this. This means `api/handler.go` doesn't care which provider is used. It just calls `.CreatePayment()`. This is **Dependency Inversion**, key for maintainable software.

## Chapter 4: Running It
1. **Setup:**
   ```bash
   make build-cpp # Compiles the .so library
   docker compose up --build
   ```
2. **Test:**
   ```bash
   curl -X POST http://localhost:8080/v1/checkout -d '{"amount": 10000, "provider_preference": "midtrans"}'
   ```

## Chapter 5: Best Practices Used
- **Idempotency:** We check if an Event ID has been processed in `HandleWebhook`.
- **Graceful Shutdown:** The worker listens for `SIGTERM` to finish current jobs before quitting.
- **Structural Logging:** JSON logs for Datadog/Splunk integration.
- **Clean Architecture:** `internal/` prevents external imports, keeping the core logic protected.

## Conclusion
You now possess the blueprint for a fintech-grade backend. Study the `internal/adapters` to see how we abstract external APIs.
