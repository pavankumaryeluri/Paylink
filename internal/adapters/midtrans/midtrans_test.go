package midtrans_adapter_test

import (
	"context"
	"net/http/httptest"
	"testing"

	midtrans_adapter "github.com/vibeswithkk/paylink/internal/adapters/midtrans"
	"github.com/vibeswithkk/paylink/internal/models"
)

func TestCreatePayment(t *testing.T) {
	adapter := midtrans_adapter.NewAdapter("test-server-key")

	tx := &models.Transaction{
		MerchantID:   "merchant-1",
		Amount:       100000,
		Currency:     "IDR",
		ProviderTxID: "order-123",
	}

	tokenID, url, err := adapter.CreatePayment(context.Background(), tx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if tokenID == "" {
		t.Error("Expected non-empty token ID")
	}

	if url == "" {
		t.Error("Expected non-empty checkout URL")
	}
}

func TestVerifySignature(t *testing.T) {
	adapter := midtrans_adapter.NewAdapter("test-server-key")

	// Test with invalid signature
	payload := []byte(`{
		"order_id": "order-123",
		"status_code": "200",
		"gross_amount": "50000.00",
		"signature_key": "invalid-signature",
		"transaction_id": "tx-456"
	}`)

	req := httptest.NewRequest("POST", "/webhook", nil)

	_, valid, err := adapter.VerifySignature(req, payload)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if valid {
		t.Error("Expected invalid signature to return false")
	}
}

func TestGetTransactionStatus(t *testing.T) {
	adapter := midtrans_adapter.NewAdapter("test-server-key")

	status, err := adapter.GetTransactionStatus(context.Background(), "tx-123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status == "" {
		t.Error("Expected non-empty status")
	}
}
