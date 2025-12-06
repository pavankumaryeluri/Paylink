package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/vibeswithkk/paylink/internal/adapters"
	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/jobs"
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
	mux.HandleFunc("POST /v1/checkout", h.Checkout)
	mux.HandleFunc("POST /v1/webhook/", h.HandleWebhook)
	mux.HandleFunc("GET /v1/tx/", h.GetTransaction)
	mux.HandleFunc("GET /health", h.Health)

	return mux
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
		util.Logger.Error("Decode error", "err", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.MerchantID == "" || req.Amount <= 0 || req.ProviderPreference == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	adapter, err := adapters.NewAdapter(req.ProviderPreference, h.Config)
	if err != nil {
		util.Logger.Error("Adapter init failed", "error", err)
		http.Error(w, "Provider unavailable", http.StatusBadRequest)
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
		util.Logger.Error("Create payment failed", "error", err)
		http.Error(w, "Payment creation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CheckoutResponse{
		CheckoutURL:  url,
		ProviderTxID: pid,
	})
}

// HandleWebhook processes incoming webhooks from payment providers
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract provider from path: /v1/webhook/{provider}
	path := strings.TrimPrefix(r.URL.Path, "/v1/webhook/")
	provider := strings.TrimSuffix(path, "/")

	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	adapter, err := adapters.NewAdapter(provider, h.Config)
	if err != nil {
		http.Error(w, "Unknown provider", http.StatusNotFound)
		return
	}

	eventID, valid, err := adapter.VerifySignature(r, body)
	if err != nil {
		util.Logger.Error("Signature verification error", "error", err)
		http.Error(w, "Verification failed", http.StatusBadRequest)
		return
	}

	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Enqueue job asynchronously
	err = h.Enqueuer.EnqueueWebhook(r.Context(), provider, body)
	if err != nil {
		util.Logger.Error("Enqueue failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	util.Logger.Info("Webhook enqueued", "provider", provider, "event_id", eventID)
	w.WriteHeader(http.StatusOK)
}

// GetTransaction retrieves a transaction by ID
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /v1/tx/{id}
	path := strings.TrimPrefix(r.URL.Path, "/v1/tx/")
	txID := strings.TrimSuffix(path, "/")

	if txID == "" {
		http.Error(w, "Transaction ID not specified", http.StatusBadRequest)
		return
	}

	// TODO: Query database for transaction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     txID,
		"status": "pending",
	})
}
