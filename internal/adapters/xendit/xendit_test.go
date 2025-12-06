package xendit_adapter_test

import (
	"context"
	"net/http/httptest"
	"testing"

	xendit_adapter "github.com/vibeswithkk/paylink/internal/adapters/xendit"
	"github.com/vibeswithkk/paylink/internal/models"
	"github.com/vibeswithkk/paylink/internal/util"
)

func init() {
	util.InitLogger()
}

func TestXenditCreatePayment(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")

	tx := &models.Transaction{
		MerchantID:   "merchant-1",
		Amount:       250000,
		Currency:     "IDR",
		ProviderTxID: "order-xendit-123",
	}

	invoiceID, url, err := adapter.CreatePayment(context.Background(), tx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if invoiceID == "" {
		t.Error("Expected non-empty invoice ID")
	}

	if url == "" {
		t.Error("Expected non-empty checkout URL")
	}

	// Verify URL format
	if len(url) < 10 {
		t.Error("Checkout URL seems too short")
	}
}

func TestXenditCreatePaymentValidation(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")

	// Test nil transaction
	_, _, err := adapter.CreatePayment(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil transaction")
	}

	// Test zero amount
	tx := &models.Transaction{
		Amount:       0,
		ProviderTxID: "order-123",
	}
	_, _, err = adapter.CreatePayment(context.Background(), tx)
	if err == nil {
		t.Error("Expected error for zero amount")
	}
}

func TestXenditVerifySignatureValid(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")
	adapter.Config.WebhookToken = "test-callback-token"

	req := httptest.NewRequest("POST", "/webhook/xendit", nil)
	req.Header.Set("x-callback-token", "test-callback-token")
	req.Header.Set("webhook-id", "evt_123456")

	payload := []byte(`{"id": "inv_123", "status": "PAID"}`)

	eventID, valid, err := adapter.VerifySignature(req, payload)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !valid {
		t.Error("Expected valid signature")
	}

	if eventID != "evt_123456" {
		t.Errorf("Expected event ID 'evt_123456', got '%s'", eventID)
	}
}

func TestXenditVerifySignatureInvalid(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")
	adapter.Config.WebhookToken = "correct-token"

	req := httptest.NewRequest("POST", "/webhook/xendit", nil)
	req.Header.Set("x-callback-token", "wrong-token")

	payload := []byte(`{"id": "inv_123"}`)

	_, valid, _ := adapter.VerifySignature(req, payload)

	if valid {
		t.Error("Expected invalid signature for wrong token")
	}
}

func TestXenditVerifySignatureEmptyPayload(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")

	req := httptest.NewRequest("POST", "/webhook/xendit", nil)
	_, _, err := adapter.VerifySignature(req, []byte{})

	if err == nil {
		t.Error("Expected error for empty payload")
	}
}

func TestXenditGetTransactionStatus(t *testing.T) {
	adapter := xendit_adapter.NewAdapter("xnd_development_test_key")

	status, err := adapter.GetTransactionStatus(context.Background(), "inv_123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status == "" {
		t.Error("Expected non-empty status")
	}
}
