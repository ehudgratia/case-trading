package repository

import (
	"case-trading/app/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	CG_BASE_URL = "https://api.coingecko.com/api/v3"
	CG_API_KEY  = "CG-XJbdP6jjgk9N9eZLjxNgXaG6"
)

type MarketRepository struct{}

func NewMarketRepository() *MarketRepository {
	return &MarketRepository{}
}

func (r *MarketRepository) GetLivePrice(ctx context.Context, ids string) (models.CoinGeckoPriceResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Gunakan endpoint simple/price
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", CG_BASE_URL, ids)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-cg-demo-api-key", CG_API_KEY)
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko error: status %d", resp.StatusCode)
	}

	var result models.CoinGeckoPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
