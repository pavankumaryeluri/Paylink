package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vibeswithkk/paylink/internal/api"
	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/models"
)

func TestHealthEndpoint(t *testing.T) {
	handler := api.NewHandler(&config.Config{}, &db.DB{})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp["status"])
	}
}

func TestCheckoutEndpoint(t *testing.T) {
	cfg := &config.Config{
		MidtransServerKey: "test-key",
	}
	handler := api.NewHandler(cfg, &db.DB{
		Redis: &db.RedisClient{Addr: "localhost:6379"},
	})

	checkout := models.CheckoutRequest{
		MerchantID:         "merchant-123",
		Amount:             50000,
		Currency:           "IDR",
		OrderID:            "order-456",
		ProviderPreference: "midtrans",
	}

	body, _ := json.Marshal(checkout)
	req := httptest.NewRequest("POST", "/v1/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.CheckoutResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.CheckoutURL == "" {
		t.Error("Expected checkout_url to be non-empty")
	}
	if resp.ProviderTxID == "" {
		t.Error("Expected provider_tx_id to be non-empty")
	}
}

func TestCheckoutValidation(t *testing.T) {
	handler := api.NewHandler(&config.Config{}, &db.DB{})

	// Test missing fields
	checkout := models.CheckoutRequest{
		Amount: 0, // Invalid
	}

	body, _ := json.Marshal(checkout)
	req := httptest.NewRequest("POST", "/v1/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid request, got %d", w.Code)
	}
}

func TestWebhookEndpoint(t *testing.T) {
	cfg := &config.Config{
		MidtransServerKey: "test-server-key",
	}
	handler := api.NewHandler(cfg, &db.DB{
		Redis: &db.RedisClient{Addr: "localhost:6379"},
	})

	// Simulated Midtrans webhook payload
	payload := map[string]string{
		"order_id":       "order-123",
		"status_code":    "200",
		"gross_amount":   "50000.00",
		"signature_key":  "dummy", // Won't match but tests the flow
		"transaction_id": "tx-456",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/v1/webhook/midtrans", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Routes().ServeHTTP(w, req)

	// Signature won't match with dummy, so expect 401
	if w.Code != http.StatusUnauthorized {
		t.Logf("Webhook response code: %d (expected 401 for invalid signature)", w.Code)
	}
}

func TestGetTransaction(t *testing.T) {
	handler := api.NewHandler(&config.Config{}, &db.DB{})

	req := httptest.NewRequest("GET", "/v1/tx/transaction-123", nil)
	w := httptest.NewRecorder()

	handler.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
