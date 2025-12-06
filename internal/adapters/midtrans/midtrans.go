package midtrans_adapter

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vibeswithkk/paylink/internal/models"
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
	// In production, this would make an HTTP request to Midtrans Snap API
	// For sandbox/demo, we return a simulated response

	// Midtrans Snap API expects:
	// POST https://app.sandbox.midtrans.com/snap/v1/transactions
	// Header: Authorization: Basic base64(ServerKey:)
	// Body: { transaction_details: { order_id, gross_amount }, ... }

	token := fmt.Sprintf("snap_%s_%d", tx.ProviderTxID, tx.Amount)
	redirectURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v3/redirection/%s", token)

	return token, redirectURL, nil
}

// VerifySignature verifies webhook signature from Midtrans
// Midtrans uses: SHA512(order_id + status_code + gross_amount + ServerKey)
func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	var payload struct {
		OrderID       string `json:"order_id"`
		StatusCode    string `json:"status_code"`
		GrossAmount   string `json:"gross_amount"`
		SignatureKey  string `json:"signature_key"`
		TransactionID string `json:"transaction_id"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false, fmt.Errorf("failed to parse payload: %w", err)
	}

	// Compute expected signature
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + a.Config.ServerKey
	hash := sha512.Sum512([]byte(raw))
	expectedSig := hex.EncodeToString(hash[:])

	// Constant-time comparison to prevent timing attacks
	valid := strings.EqualFold(expectedSig, payload.SignatureKey)

	return payload.TransactionID, valid, nil
}

// GetTransactionStatus retrieves transaction status from Midtrans
func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	// In production: GET https://api.sandbox.midtrans.com/v2/{order_id}/status
	// For now, return pending
	return "pending", nil
}
