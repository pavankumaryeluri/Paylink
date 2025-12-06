package adapters

import (
	"fmt"

	midtrans_adapter "github.com/vibeswithkk/paylink/internal/adapters/midtrans"
	xendit_adapter "github.com/vibeswithkk/paylink/internal/adapters/xendit"
	"github.com/vibeswithkk/paylink/internal/config"
)

func NewAdapter(provider string, cfg *config.Config) (ProviderAdapter, error) {
	switch provider {
	case "midtrans":
		return midtrans_adapter.NewAdapter(cfg.MidtransServerKey), nil
	case "xendit":
		return xendit_adapter.NewAdapter(cfg.XenditAPIKey), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
