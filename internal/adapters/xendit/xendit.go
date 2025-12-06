package xendit_adapter

import (
	"context"
	"net/http"

	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/xendit/xendit-go/v4"
)

type Adapter struct {
	Client *xendit.APIClient
}

func NewAdapter(apiKey string) *Adapter {
	client := xendit.NewClient(apiKey)
	return &Adapter{Client: client}
}

func (a *Adapter) CreatePayment(ctx context.Context, tx *models.Transaction) (string, string, error) {
	// Mock implementation for skeleton
	return "xnd_inv_mock_123", "https://checkout.xendit.co/web/123", nil
}

func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	// Xendit uses x-callback-token header usually, or JWS signature
	return "evt_xnd_mock", true, nil
}

func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	return "PENDING", nil
}
