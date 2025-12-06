package adapters

import (
	"context"
	"net/http"

	"github.com/vibeswithkk/paylink/internal/models"
)

type ProviderAdapter interface {
	CreatePayment(ctx context.Context, tx *models.Transaction) (providerTxID string, checkoutURL string, err error)
	VerifySignature(r *http.Request, body []byte) (eventID string, valid bool, err error)
	GetTransactionStatus(ctx context.Context, providerTxID string) (status string, err error)
}
