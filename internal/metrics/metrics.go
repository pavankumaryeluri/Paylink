package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Checkout metrics
	CheckoutsCreated    int64
	CheckoutsByProvider map[string]int64

	// Webhook metrics
	WebhooksReceived  int64
	WebhooksProcessed int64
	WebhooksFailed    int64

	// Latency tracking (in milliseconds)
	AvgLatencyMs float64
	totalLatency int64
	latencyCount int64

	// Uptime
	StartTime time.Time
}

var globalMetrics *Metrics
var once sync.Once

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			CheckoutsByProvider: make(map[string]int64),
			StartTime:           time.Now(),
		}
	})
	return globalMetrics
}

// IncrementRequests increments the request counter
func (m *Metrics) IncrementRequests(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
	if success {
		m.SuccessfulRequests++
	} else {
		m.FailedRequests++
	}
}

// IncrementCheckout increments checkout counter for a provider
func (m *Metrics) IncrementCheckout(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CheckoutsCreated++
	m.CheckoutsByProvider[provider]++
}

// IncrementWebhook increments webhook counters
func (m *Metrics) IncrementWebhook(processed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WebhooksReceived++
	if processed {
		m.WebhooksProcessed++
	} else {
		m.WebhooksFailed++
	}
}

// RecordLatency records a request latency
func (m *Metrics) RecordLatency(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ms := duration.Milliseconds()
	m.totalLatency += ms
	m.latencyCount++
	if m.latencyCount > 0 {
		m.AvgLatencyMs = float64(m.totalLatency) / float64(m.latencyCount)
	}
}

// GetSnapshot returns a copy of current metrics
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime).Seconds()

	providersCopy := make(map[string]int64)
	for k, v := range m.CheckoutsByProvider {
		providersCopy[k] = v
	}

	return map[string]interface{}{
		"uptime_seconds":        uptime,
		"total_requests":        m.TotalRequests,
		"successful_requests":   m.SuccessfulRequests,
		"failed_requests":       m.FailedRequests,
		"checkouts_created":     m.CheckoutsCreated,
		"checkouts_by_provider": providersCopy,
		"webhooks_received":     m.WebhooksReceived,
		"webhooks_processed":    m.WebhooksProcessed,
		"webhooks_failed":       m.WebhooksFailed,
		"avg_latency_ms":        m.AvgLatencyMs,
	}
}

// PrometheusHandler returns metrics in Prometheus format
func PrometheusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := GetMetrics()
		m.mu.RLock()
		defer m.mu.RUnlock()

		uptime := time.Since(m.StartTime).Seconds()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Prometheus format: metric_name{labels} value
		writeMetric(w, "# HELP paylink_uptime_seconds Time since server start\n")
		writeMetric(w, "# TYPE paylink_uptime_seconds gauge\n")
		writeMetric(w, fmt.Sprintf("paylink_uptime_seconds %.2f\n\n", uptime))

		writeMetric(w, "# HELP paylink_requests_total Total number of requests\n")
		writeMetric(w, "# TYPE paylink_requests_total counter\n")
		writeMetric(w, fmt.Sprintf("paylink_requests_total{status=\"success\"} %d\n", m.SuccessfulRequests))
		writeMetric(w, fmt.Sprintf("paylink_requests_total{status=\"failed\"} %d\n\n", m.FailedRequests))

		writeMetric(w, "# HELP paylink_checkouts_total Total checkouts created\n")
		writeMetric(w, "# TYPE paylink_checkouts_total counter\n")
		writeMetric(w, fmt.Sprintf("paylink_checkouts_total %d\n\n", m.CheckoutsCreated))

		writeMetric(w, "# HELP paylink_checkouts_by_provider Checkouts by provider\n")
		writeMetric(w, "# TYPE paylink_checkouts_by_provider counter\n")
		for provider, count := range m.CheckoutsByProvider {
			writeMetric(w, fmt.Sprintf("paylink_checkouts_by_provider{provider=\"%s\"} %d\n", provider, count))
		}
		writeMetric(w, "\n")

		writeMetric(w, "# HELP paylink_webhooks_total Total webhooks\n")
		writeMetric(w, "# TYPE paylink_webhooks_total counter\n")
		writeMetric(w, fmt.Sprintf("paylink_webhooks_total{status=\"received\"} %d\n", m.WebhooksReceived))
		writeMetric(w, fmt.Sprintf("paylink_webhooks_total{status=\"processed\"} %d\n", m.WebhooksProcessed))
		writeMetric(w, fmt.Sprintf("paylink_webhooks_total{status=\"failed\"} %d\n\n", m.WebhooksFailed))

		writeMetric(w, "# HELP paylink_latency_avg_ms Average request latency\n")
		writeMetric(w, "# TYPE paylink_latency_avg_ms gauge\n")
		writeMetric(w, fmt.Sprintf("paylink_latency_avg_ms %.2f\n", m.AvgLatencyMs))
	}
}

func writeMetric(w io.Writer, s string) {
	w.Write([]byte(s))
}
