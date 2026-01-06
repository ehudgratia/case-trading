package models

type CoinGeckoPriceResponse map[string]map[string]float64

type MarketPriceResponse struct {
	ID       int     `json:"id"`
	CoinName string  `json:"coin_name"`
	PriceUSD float64 `json:"price_usd"`
}
