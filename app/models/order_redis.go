package models

import "time"

type RedisOrder struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MarketID  int       `json:"market_id"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	FilledQty float64   `json:"filled_qty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type MatchResult struct {
	CounterID int     `json:"counter_id"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
}
