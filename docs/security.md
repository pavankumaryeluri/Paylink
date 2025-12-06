# Security Documentation

This document describes the security measures implemented in PayLink.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Webhook Security](#webhook-security)
4. [Input Validation](#input-validation)
5. [Cryptographic Operations](#cryptographic-operations)
6. [Secrets Management](#secrets-management)
7. [OWASP Compliance](#owasp-compliance)

---

## Overview

PayLink implements defense-in-depth security with multiple layers of protection:

| Layer | Protection |
|-------|------------|
| Transport | HTTPS/TLS encryption |
| Authentication | API key validation |
| Authorization | Merchant-scoped access |
| Input | Validation and sanitization |
| Crypto | Constant-time operations |
| Logging | Sensitive data masking |

---

## Authentication

### API Key Authentication

All API requests (except /health and /metrics) require authentication via API key:

```
Header: X-Api-Key: <merchant_api_key>
```

### Key Storage

- API keys are stored as bcrypt hashes in the database
- Original keys are never stored
- Keys are validated using constant-time comparison

### Key Rotation

Recommended key rotation schedule:
- Production: Every 90 days
- Sandbox: As needed

---

## Webhook Security

### Signature Verification

Each provider uses specific signature verification:

#### Midtrans

```
Algorithm: SHA512
Input: order_id + status_code + gross_amount + ServerKey
Location: JSON body field "signature_key"
```

**Implementation:**
```go
raw := orderID + statusCode + grossAmount + serverKey
expected := sha512.Sum512([]byte(raw))
valid := constantTimeCompare(hex.EncodeToString(expected[:]), signatureKey)
```

#### Xendit

```
Algorithm: Token comparison
Location: Header "x-callback-token"
```

**Implementation:**
```go
callbackToken := r.Header.Get("x-callback-token")
valid := hmac.Equal([]byte(callbackToken), []byte(configuredToken))
```

### Timing Attack Prevention

All signature comparisons use constant-time algorithms:

```go
func constantTimeCompare(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    var result byte
    for i := 0; i < len(a); i++ {
        result |= a[i] ^ b[i]
    }
    return result == 0
}
```

---

## Input Validation

### Request Validation

All incoming requests are validated:

| Field | Validation |
|-------|------------|
| merchant_id | UUID format |
| amount | Positive integer, max 999999999 |
| currency | ISO 4217, 3 characters |
| order_id | Alphanumeric, max 50 chars |
| provider | Enum: midtrans, xendit |

### Sanitization Rules

| Input | Sanitization |
|-------|--------------|
| Order ID | Regex: `^[a-zA-Z0-9_-]+$` |
| External ID | Regex: `^[a-zA-Z0-9_-]+$` |
| Amount | Parsed as int64, validated range |

### JSON Validation

All JSON payloads are validated before parsing:

```go
if !json.Valid(body) {
    return fmt.Errorf("invalid JSON payload")
}
```

---

## Cryptographic Operations

### Why Pure Go?

PayLink uses Go's standard library for cryptographic operations:

| Reason | Explanation |
|--------|-------------|
| Portability | No CGO dependencies required |
| Security | Well-audited standard library |
| Simplicity | No external build requirements |
| Performance | Sufficient for most workloads |

### HMAC-SHA256

Used for general signature operations:

```go
func ComputeHMAC(key, data string) string {
    h := hmac.New(sha256.New, []byte(key))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}
```

### Optional C++ Module

For high-throughput scenarios (>100k webhook verifications/second), an optional C++ module is available:

- Location: `crypto_cpp/`
- Build: CMake + OpenSSL
- Integration: CGO wrapper

**When to use C++:**
- Batch signature verification
- High-frequency webhook processing
- CPU-bound crypto operations

---

## Secrets Management

### Environment Variables

All secrets are loaded from environment variables:

```bash
MIDTRANS_SERVER_KEY=your-server-key
XENDIT_API_KEY=your-api-key
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

### Best Practices

| Environment | Recommendation |
|-------------|----------------|
| Development | `.env` file (gitignored) |
| CI/CD | GitHub Secrets |
| Production | Secret manager (Vault, AWS Secrets Manager) |

### What NOT to Do

- Never commit secrets to git
- Never log secrets
- Never include secrets in error messages
- Never expose secrets in API responses

---

## OWASP Compliance

PayLink follows OWASP security guidelines:

### A1: Injection

- Parameterized SQL queries via database/sql
- Input validation before use
- No dynamic query construction

### A2: Broken Authentication

- API key required for all sensitive endpoints
- Keys stored as hashes
- Rate limiting recommended

### A3: Sensitive Data Exposure

- Secrets never logged
- API keys masked in responses
- HTTPS required in production

### A5: Security Misconfiguration

- Environment-based configuration
- No default credentials
- Secure defaults

### A7: Cross-Site Scripting (XSS)

- JSON-only API (no HTML rendering)
- Proper Content-Type headers

### A8: Insecure Deserialization

- JSON validation before parsing
- Type-safe unmarshaling

### A9: Using Components with Known Vulnerabilities

- Minimal dependencies
- Regular dependency audit (`go mod verify`)

### A10: Insufficient Logging & Monitoring

- Structured JSON logging
- Request/response logging
- Error tracking
- Metrics endpoint

---

## Security Checklist

Before deploying to production:

- [ ] All secrets loaded from environment/secret manager
- [ ] HTTPS enabled and enforced
- [ ] API key authentication active
- [ ] Rate limiting configured
- [ ] Input validation enabled
- [ ] Webhook signatures verified
- [ ] Logging configured (secrets masked)
- [ ] Monitoring/alerting set up
- [ ] Dependency audit completed

---

## Reporting Security Issues

If you discover a security vulnerability, please report it via:

- Email: security@paylink.dev
- Do NOT create public GitHub issues for security vulnerabilities

---

*Document maintained by PayLink Security Team*
