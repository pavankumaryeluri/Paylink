package models

import (
	"time"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID           string                 `json:"id"`
	MerchantID   string                 `json:"merchant_id"`
	Provider     string                 `json:"provider"`
	ProviderTxID string                 `json:"provider_tx_id"`
	Amount       int64                  `json:"amount"` // in smallest currency unit
	Currency     string                 `json:"currency"`
	Status       string                 `json:"status"` // PENDING, PAID, FAILED, EXPIRED
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CheckoutRequest represents a checkout creation request
type CheckoutRequest struct {
	MerchantID         string `json:"merchant_id"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	OrderID            string `json:"order_id"`
	ProviderPreference string `json:"provider_preference"`
}

// CheckoutResponse represents a checkout creation response
type CheckoutResponse struct {
	CheckoutURL  string `json:"checkout_url"`
	ProviderTxID string `json:"provider_tx_id"`
}

// WebhookEvent represents a received webhook event
type WebhookEvent struct {
	ID         string                 `json:"id"`
	Provider   string                 `json:"provider"`
	EventID    string                 `json:"event_id"`
	Payload    map[string]interface{} `json:"payload"`
	Processed  bool                   `json:"processed"`
	ReceivedAt time.Time              `json:"received_at"`
}

// Merchant represents a registered merchant
type Merchant struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	APIKeyHash string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}
