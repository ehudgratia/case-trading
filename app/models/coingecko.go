package models

// Response dari CoinGecko
type CoinGeckoPriceResponse map[string]map[string]float64

// Struct untuk response seragam di API kamu
type MarketPriceResponse struct {
	CoinName string  `json:"coin_name"`
	PriceUSD float64 `json:"price_usd"`
}
