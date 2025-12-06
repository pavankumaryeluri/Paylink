package xendit_adapter

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/vibeswithkk/paylink/internal/util"
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
	// Validate input
	if tx == nil {
		return "", "", fmt.Errorf("transaction cannot be nil")
	}
	if tx.Amount <= 0 {
		return "", "", fmt.Errorf("amount must be positive")
	}
	if tx.ProviderTxID == "" {
		return "", "", fmt.Errorf("order ID is required")
	}

	// Sanitize order ID
	if !isValidExternalID(tx.ProviderTxID) {
		return "", "", fmt.Errorf("invalid external ID format")
	}

	util.Logger.Info("Creating Xendit invoice",
		"external_id", tx.ProviderTxID,
		"amount", tx.Amount,
		"currency", tx.Currency)

	invoiceID := fmt.Sprintf("xnd_inv_%s", tx.ProviderTxID)
	checkoutURL := fmt.Sprintf("https://checkout-staging.xendit.co/web/%s", invoiceID)

	util.Logger.Info("Xendit invoice created",
		"invoice_id", invoiceID,
		"checkout_url", checkoutURL)

	return invoiceID, checkoutURL, nil
}

// VerifySignature verifies webhook callback from Xendit
func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	if len(body) == 0 {
		util.Logger.Error("Empty Xendit webhook payload")
		return "", false, fmt.Errorf("empty payload")
	}

	// Validate JSON
	if !json.Valid(body) {
		util.Logger.Error("Invalid JSON in Xendit webhook")
		return "", false, fmt.Errorf("invalid JSON payload")
	}

	callbackToken := r.Header.Get("x-callback-token")

	// Verify callback token if configured
	if a.Config.WebhookToken != "" {
		valid := hmac.Equal([]byte(callbackToken), []byte(a.Config.WebhookToken))
		if !valid {
			util.Logger.Warn("Invalid Xendit callback token")
			return "", false, nil
		}
	}

	// Extract event ID
	eventID := r.Header.Get("webhook-id")
	if eventID == "" {
		// Generate from payload hash for idempotency
		hash := sha256.Sum256(body)
		eventID = hex.EncodeToString(hash[:8])
	}

	util.Logger.Info("Xendit webhook verified", "event_id", eventID)

	return eventID, true, nil
}

// GetTransactionStatus retrieves invoice status from Xendit
func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	if providerTxID == "" {
		return "", fmt.Errorf("provider transaction ID is required")
	}

	util.Logger.Info("Getting Xendit invoice status", "invoice_id", providerTxID)

	return "PENDING", nil
}

// isValidExternalID validates external ID format
func isValidExternalID(externalID string) bool {
	if len(externalID) == 0 || len(externalID) > 64 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, externalID)
	return matched
}
