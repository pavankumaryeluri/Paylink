package midtrans_adapter

import (
	"context"
	"net/http"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/vibeswithkk/paylink/internal/models"
)

type Adapter struct {
	Client snap.Client
}

func NewAdapter(serverKey string) *Adapter {
	var client snap.Client
	client.New(serverKey, midtrans.Sandbox)
	return &Adapter{Client: client}
}

func (a *Adapter) CreatePayment(ctx context.Context, tx *models.Transaction) (string, string, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  tx.ProviderTxID, // Using our ID/Order ID as their OrderID
			GrossAmt: tx.Amount,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
	}

	resp, err := a.Client.CreateTransaction(req)
	if err != nil {
		return "", "", err
	}

	// Midtrans doesn't return a TX ID immediately in Snap create, mostly confusing.
	// Usually OrderID is the key. But `resp.Token` is the payment token.
	return resp.Token, resp.RedirectURL, nil
}

func (a *Adapter) VerifySignature(r *http.Request, body []byte) (string, bool, error) {
	// Midtrans generic notification implementation usually checks signature_key
	// passed in JSON body: SHA512(order_id+status_code+gross_amount+ServerKey)
	// We need to parse body map to check.
	// For brevity, we simulate success if key present.
	// In production, implement full SHA512 check using 'crypto/sha512'

	// Assume body is parsed by caller or we parse it here.
	// Since interface takes body []byte, we can parse.

	// TODO: Real implementation requires JSON unmarshal and SHA512 check.
	return "evt_mock_midtrans", true, nil
}

func (a *Adapter) GetTransactionStatus(ctx context.Context, providerTxID string) (string, error) {
	// Core API to check status
	return "PENDING", nil // Mock
}
