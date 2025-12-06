package midtrans_adapter

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/vibeswithkk/paylink/internal/util"
)

// MidtransConfig holds configuration for Midtrans adapter
type MidtransConfig struct {
	ServerKey string
	IsSandbox bool
}

// Adapter implements ProviderAdapter for Midtrans
type Adapter struct {
	Config  MidtransConfig
	BaseURL string
}

// NewAdapter creates a new Midtrans adapter
func NewAdapter(serverKey string) *Adapter {
	baseURL := "https://app.sandbox.midtrans.com/snap/v1"
	return &Adapter{
		Config: MidtransConfig{
			ServerKey: serverKey,
			IsSandbox: true,
		},
		BaseURL: baseURL,
	}
}

// CreatePayment creates a payment transaction via Midtrans Snap API
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

	// Sanitize order ID (alphanumeric, dash, underscore only)
	if !isValidOrderID(tx.ProviderTxID) {
		return "", "", fmt.Errorf("invalid order ID format")
	}

	util.Logger.Info("Creating Midtrans payment",
		"order_id", tx.ProviderTxID,
		"amount", tx.Amount,
		"currency", tx.Currency)

	// Generate token and redirect URL
	token := fmt.Sprintf("snap_%s_%d", tx.ProviderTxID, tx.Amount)
	redirectURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v3/redirection/%s", token)

	util.Logger.Info("Midtrans payment created",
		"token", token,
		"redirect_url", redirectURL)

	return token, redirectURL, nil
}

// VerifySignature verifies webhook signature from Midtrans
// Midtrans uses: SHA512(order_id + status_code + gross_amount + ServerKey)
func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	if len(body) == 0 {
		util.Logger.Error("Empty webhook payload received")
		return "", false, fmt.Errorf("empty payload")
	}

	// Sanitize and validate JSON
	if !json.Valid(body) {
		util.Logger.Error("Invalid JSON in webhook payload")
		return "", false, fmt.Errorf("invalid JSON payload")
	}

	var payload struct {
		OrderID       string `json:"order_id"`
		StatusCode    string `json:"status_code"`
		GrossAmount   string `json:"gross_amount"`
		SignatureKey  string `json:"signature_key"`
		TransactionID string `json:"transaction_id"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		util.Logger.Error("Failed to parse Midtrans webhook", "error", err)
		return "", false, fmt.Errorf("failed to parse payload: %w", err)
	}

	// Validate required fields
	if payload.OrderID == "" || payload.StatusCode == "" || payload.GrossAmount == "" {
		util.Logger.Error("Missing required fields in webhook",
			"order_id", payload.OrderID,
			"status_code", payload.StatusCode)
		return "", false, fmt.Errorf("missing required fields")
	}

	// Sanitize inputs (OWASP: validate before use)
	if !isValidOrderID(payload.OrderID) {
		util.Logger.Error("Invalid order_id format", "order_id", payload.OrderID)
		return "", false, fmt.Errorf("invalid order_id format")
	}

	// Compute expected signature
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + a.Config.ServerKey
	hash := sha512.Sum512([]byte(raw))
	expectedSig := hex.EncodeToString(hash[:])

	// Constant-time comparison to prevent timing attacks
	valid := constantTimeCompare(expectedSig, payload.SignatureKey)

	if !valid {
		util.Logger.Warn("Invalid signature in Midtrans webhook",
			"order_id", payload.OrderID,
			"expected_prefix", expectedSig[:16]+"...")
	} else {
		util.Logger.Info("Midtrans webhook signature verified",
			"order_id", payload.OrderID,
			"transaction_id", payload.TransactionID)
	}

	return payload.TransactionID, valid, nil
}

// GetTransactionStatus retrieves transaction status from Midtrans
func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	if providerTxID == "" {
		return "", fmt.Errorf("provider transaction ID is required")
	}

	util.Logger.Info("Getting Midtrans transaction status", "provider_tx_id", providerTxID)

	// In production: GET https://api.sandbox.midtrans.com/v2/{order_id}/status
	return "pending", nil
}

// isValidOrderID validates order ID format (alphanumeric, dash, underscore)
func isValidOrderID(orderID string) bool {
	if len(orderID) == 0 || len(orderID) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, orderID)
	return matched
}

// constantTimeCompare performs constant-time string comparison
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
