package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID              `json:"id"`
	MerchantID   uuid.UUID              `json:"merchant_id"`
	Provider     string                 `json:"provider"`
	ProviderTxID string                 `json:"provider_tx_id"`
	Amount       int64                  `json:"amount"` // in smallest currency unit
	Currency     string                 `json:"currency"`
	Status       string                 `json:"status"` // PENDING, PAID, FAILED, EXPIRED
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type CheckoutRequest struct {
	MerchantID         string `json:"merchant_id"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	OrderID            string `json:"order_id"`
	ProviderPreference string `json:"provider_preference"`
}

type CheckoutResponse struct {
	CheckoutURL  string `json:"checkout_url"`
	ProviderTxID string `json:"provider_tx_id"`
}
