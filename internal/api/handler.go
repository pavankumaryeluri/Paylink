package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vibeswithkk/paylink/internal/adapters"
	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/jobs"
	"github.com/vibeswithkk/paylink/internal/metrics"
	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/vibeswithkk/paylink/internal/util"
)

// Handler handles HTTP requests
type Handler struct {
	Config   *config.Config
	DB       *db.DB
	Enqueuer *jobs.Enqueuer
}

// NewHandler creates a new HTTP handler
func NewHandler(cfg *config.Config, database *db.DB) *Handler {
	return &Handler{
		Config:   cfg,
		DB:       database,
		Enqueuer: jobs.NewEnqueuer(database.Redis),
	}
}

// Routes returns the HTTP router
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// API v1 routes
	mux.HandleFunc("POST /v1/checkout", h.withMetrics(h.Checkout))
	mux.HandleFunc("POST /v1/webhook/", h.withMetrics(h.HandleWebhook))
	mux.HandleFunc("GET /v1/tx/", h.withMetrics(h.GetTransaction))

	// System routes
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	return mux
}

// withMetrics wraps a handler with metrics collection
func (h *Handler) withMetrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(rw, r)

		duration := time.Since(start)
		success := rw.statusCode >= 200 && rw.statusCode < 400

		m := metrics.GetMetrics()
		m.IncrementRequests(success)
		m.RecordLatency(duration)

		util.Logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds())
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Health returns server health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Checkout creates a new payment checkout
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Logger.Error("Failed to decode checkout request", "error", err)
		h.errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := h.validateCheckoutRequest(&req); err != nil {
		util.Logger.Warn("Checkout validation failed", "error", err)
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	adapter, err := adapters.NewAdapter(req.ProviderPreference, h.Config)
	if err != nil {
		util.Logger.Error("Failed to create adapter", "provider", req.ProviderPreference, "error", err)
		h.errorResponse(w, "Provider unavailable", http.StatusBadRequest)
		return
	}

	tx := &models.Transaction{
		MerchantID:   req.MerchantID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Provider:     req.ProviderPreference,
		ProviderTxID: req.OrderID,
	}

	pid, url, err := adapter.CreatePayment(r.Context(), tx)
	if err != nil {
		util.Logger.Error("Failed to create payment", "error", err)
		h.errorResponse(w, "Payment creation failed", http.StatusInternalServerError)
		return
	}

	// Track metrics
	metrics.GetMetrics().IncrementCheckout(req.ProviderPreference)

	util.Logger.Info("Checkout created",
		"provider", req.ProviderPreference,
		"order_id", req.OrderID,
		"amount", req.Amount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CheckoutResponse{
		CheckoutURL:  url,
		ProviderTxID: pid,
	})
}

func (h *Handler) validateCheckoutRequest(req *models.CheckoutRequest) error {
	if req.MerchantID == "" {
		return fmt.Errorf("merchant_id is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if req.Amount > 999999999 {
		return fmt.Errorf("amount exceeds maximum")
	}
	if req.Currency == "" || len(req.Currency) != 3 {
		return fmt.Errorf("currency must be 3-character ISO code")
	}
	if req.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if len(req.OrderID) > 50 {
		return fmt.Errorf("order_id too long")
	}
	if req.ProviderPreference == "" {
		return fmt.Errorf("provider_preference is required")
	}
	return nil
}

// HandleWebhook processes incoming webhooks from payment providers
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract provider from path: /v1/webhook/{provider}
	path := strings.TrimPrefix(r.URL.Path, "/v1/webhook/")
	provider := strings.TrimSuffix(path, "/")

	if provider == "" {
		h.errorResponse(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	// Read body with size limit (1MB max)
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		util.Logger.Error("Failed to read webhook body", "error", err)
		h.errorResponse(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		h.errorResponse(w, "Empty request body", http.StatusBadRequest)
		return
	}

	adapter, err := adapters.NewAdapter(provider, h.Config)
	if err != nil {
		util.Logger.Warn("Unknown provider in webhook", "provider", provider)
		h.errorResponse(w, "Unknown provider", http.StatusNotFound)
		return
	}

	eventID, valid, err := adapter.VerifySignature(r, body)
	if err != nil {
		util.Logger.Error("Webhook verification error", "provider", provider, "error", err)
		metrics.GetMetrics().IncrementWebhook(false)
		h.errorResponse(w, "Verification failed", http.StatusBadRequest)
		return
	}

	if !valid {
		util.Logger.Warn("Invalid webhook signature", "provider", provider)
		metrics.GetMetrics().IncrementWebhook(false)
		h.errorResponse(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Enqueue job asynchronously
	err = h.Enqueuer.EnqueueWebhook(r.Context(), provider, body)
	if err != nil {
		util.Logger.Error("Failed to enqueue webhook", "error", err)
		metrics.GetMetrics().IncrementWebhook(false)
		h.errorResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	metrics.GetMetrics().IncrementWebhook(true)

	util.Logger.Info("Webhook enqueued", "provider", provider, "event_id", eventID)
	w.WriteHeader(http.StatusOK)
}

// GetTransaction retrieves a transaction by ID
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /v1/tx/{id}
	path := strings.TrimPrefix(r.URL.Path, "/v1/tx/")
	txID := strings.TrimSuffix(path, "/")

	if txID == "" {
		h.errorResponse(w, "Transaction ID not specified", http.StatusBadRequest)
		return
	}

	util.Logger.Info("Transaction status requested", "tx_id", txID)

	// TODO: Query database for transaction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     txID,
		"status": "pending",
	})
}

func (h *Handler) errorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
