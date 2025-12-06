package xendit_adapter

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/vibeswithkk/paylink/internal/models"
)

// XenditConfig holds configuration for Xendit adapter
type XenditConfig struct {
	APIKey       string
	WebhookToken string
}

// Adapter implements ProviderAdapter for Xendit
type Adapter struct {
	Config  XenditConfig
	BaseURL string
}

// NewAdapter creates a new Xendit adapter
func NewAdapter(apiKey string) *Adapter {
	return &Adapter{
		Config: XenditConfig{
			APIKey: apiKey,
		},
		BaseURL: "https://api.xendit.co",
	}
}

// CreatePayment creates an invoice via Xendit API
func (a *Adapter) CreatePayment(ctx context.Context, tx *models.Transaction) (string, string, error) {
	// In production, this would POST to https://api.xendit.co/v2/invoices
	// For sandbox/demo, return simulated response

	invoiceID := fmt.Sprintf("xnd_inv_%s", tx.ProviderTxID)
	checkoutURL := fmt.Sprintf("https://checkout-staging.xendit.co/web/%s", invoiceID)

	return invoiceID, checkoutURL, nil
}

// VerifySignature verifies webhook callback from Xendit
// Xendit uses x-callback-token header for verification
func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	callbackToken := r.Header.Get("x-callback-token")

	// If webhook token is configured, verify it
	if a.Config.WebhookToken != "" {
		valid := hmac.Equal([]byte(callbackToken), []byte(a.Config.WebhookToken))
		if !valid {
			return "", false, nil
		}
	}

	// Extract event ID from header or generate from body hash
	eventID := r.Header.Get("webhook-id")
	if eventID == "" {
		hash := sha256.Sum256(body)
		eventID = hex.EncodeToString(hash[:8])
	}

	return eventID, true, nil
}

// GetTransactionStatus retrieves invoice status from Xendit
func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	// In production: GET https://api.xendit.co/v2/invoices/{invoice_id}
	return "PENDING", nil
}
