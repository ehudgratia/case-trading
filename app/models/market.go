package models

import "time"

type Market struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	BaseAsset  string     `json:"base_asset" gorm:"type:varchar(10);not null"`
	QuoteAsset string     `json:"quote_asset" gorm:"type:varchar(10);not null"`
	LastPrice  float64    `json:"last_price" gorm:"not null;default:0"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type AddMarket struct {
	BaseAsset  string `json:"base_asset"`
	QuoteAsset string `json:"quote_asset"`
}

type MarketData struct {
	ID         int     `json:"id"`
	BaseAsset  string  `json:"base_asset"`
	QuoteAsset string  `json:"quote_asset"`
	LastPrice  float64 `json:"last_price"`
	IsActive   bool    `json:"is_active"`
}

type MarketRespons struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    *MarketData `json:"data"`
}
