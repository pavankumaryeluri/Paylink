package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibeswithkk/paylink/internal/adapters"
	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/jobs"
	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/vibeswithkk/paylink/internal/util"
)

type Handler struct {
	Config   *config.Config
	DB       *db.DB
	Enqueuer *jobs.Enqueuer
}

func NewHandler(cfg *config.Config, database *db.DB) *Handler {
	return &Handler{
		Config:   cfg,
		DB:       database,
		Enqueuer: jobs.NewEnqueuer(database.Redis),
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/checkout", h.Checkout)
	r.Post("/webhook/{provider}", h.HandleWebhook)
	return r
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Logger.Error("Decode error", "err", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	adapter, err := adapters.NewAdapter(req.ProviderPreference, h.Config)
	if err != nil {
		util.Logger.Error("Adapter init failed", "error", err)
		http.Error(w, "Provider unavailable", http.StatusBadRequest)
		return
	}

	merchantUUID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		http.Error(w, "Invalid Merchant ID", http.StatusBadRequest)
		return
	}

	tx := &models.Transaction{
		MerchantID:   merchantUUID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Provider:     req.ProviderPreference,
		ProviderTxID: req.OrderID, // Temp
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

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	body, _ := io.ReadAll(r.Body)

	adapter, err := adapters.NewAdapter(provider, h.Config)
	if err != nil {
		http.Error(w, "Unknown provider", http.StatusNotFound)
		return
	}

	_, valid, err := adapter.VerifySignature(r, body)
	if !valid || err != nil {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Enqueue job to worker
	err = h.Enqueuer.EnqueueWebhook(r.Context(), provider, body)
	if err != nil {
		util.Logger.Error("Enqueue failed", "error", err)
		// We might still return 200 IF we want to allow retry later by provider,
		// but usually 500 triggers retry.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	util.Logger.Info("Webhook enqueued", "provider", provider)
	w.WriteHeader(http.StatusOK)
}
