# API Examples

Quick reference for common API operations using curl.

## Health Check

```bash
curl -X GET http://localhost:8080/health
```

**Response:**
```json
{"status":"ok"}
```

---

## Create Checkout

### Midtrans

```bash
curl -X POST http://localhost:8080/v1/checkout \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your-api-key" \
  -d '{
    "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 150000,
    "currency": "IDR",
    "order_id": "ORDER-123",
    "provider_preference": "midtrans"
  }'
```

**Response:**
```json
{
  "checkout_url": "https://app.sandbox.midtrans.com/snap/v3/redirection/snap_ORDER-123_150000",
  "provider_tx_id": "snap_ORDER-123_150000"
}
```

### Xendit

```bash
curl -X POST http://localhost:8080/v1/checkout \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your-api-key" \
  -d '{
    "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 250000,
    "currency": "IDR",
    "order_id": "ORDER-456",
    "provider_preference": "xendit"
  }'
```

**Response:**
```json
{
  "checkout_url": "https://checkout-staging.xendit.co/web/xnd_inv_ORDER-456",
  "provider_tx_id": "xnd_inv_ORDER-456"
}
```

---

## Get Transaction Status

```bash
curl -X GET http://localhost:8080/v1/tx/ORDER-123 \
  -H "X-Api-Key: your-api-key"
```

**Response:**
```json
{
  "id": "ORDER-123",
  "status": "pending"
}
```

---

## Get Metrics

```bash
curl -X GET http://localhost:8080/metrics
```

**Response:**
```
# HELP paylink_uptime_seconds Time since server start
# TYPE paylink_uptime_seconds gauge
paylink_uptime_seconds 3600.00

# HELP paylink_requests_total Total number of requests
# TYPE paylink_requests_total counter
paylink_requests_total{status="success"} 150
paylink_requests_total{status="failed"} 3

# HELP paylink_checkouts_total Total checkouts created
# TYPE paylink_checkouts_total counter
paylink_checkouts_total 100

# HELP paylink_checkouts_by_provider Checkouts by provider
# TYPE paylink_checkouts_by_provider counter
paylink_checkouts_by_provider{provider="midtrans"} 60
paylink_checkouts_by_provider{provider="xendit"} 40
```

---

## Simulate Webhook (Testing)

### Midtrans Webhook

```bash
curl -X POST http://localhost:8080/v1/webhook/midtrans \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORDER-123",
    "status_code": "200",
    "gross_amount": "150000.00",
    "signature_key": "YOUR_COMPUTED_SIGNATURE",
    "transaction_id": "tx-789",
    "transaction_status": "settlement"
  }'
```

### Xendit Webhook

```bash
curl -X POST http://localhost:8080/v1/webhook/xendit \
  -H "Content-Type: application/json" \
  -H "x-callback-token: your-callback-token" \
  -H "webhook-id: evt-123" \
  -d '{
    "id": "inv_456",
    "external_id": "ORDER-456",
    "status": "PAID",
    "amount": 250000
  }'
```

---

## Error Responses

### Invalid Request (400)

```bash
curl -X POST http://localhost:8080/v1/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "amount": -100
  }'
```

**Response:**
```json
{"error":"amount must be positive"}
```

### Unknown Provider (404)

```bash
curl -X POST http://localhost:8080/v1/webhook/unknown
```

**Response:**
```json
{"error":"Unknown provider"}
```

---

## Environment Variables for Testing

```bash
export PAYLINK_URL=http://localhost:8080
export API_KEY=your-api-key

# Then use in curl
curl -X GET ${PAYLINK_URL}/health
```
