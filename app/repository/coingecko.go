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

// Ambil data harga dari CoinGecko
func (s *Service) GetLiveMarketPrice(ctx context.Context, coinIDs string) (models.CoinGeckoPriceResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", CG_BASE_URL, coinIDs)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Setup Header Authentication
	req.Header.Set("x-cg-demo-api-key", CG_API_KEY)
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko api error: status %d", resp.StatusCode)
	}

	var result models.CoinGeckoPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
