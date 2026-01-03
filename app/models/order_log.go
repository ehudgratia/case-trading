package models

import "time"

type OrderTrade struct {
	ID          int     `json:"id" gorm:"primaryKey"`
	MarketID    int     `json:"market_id" gorm:"index;not null"`
	BuyOrderID  int     `json:"buy_order_id" gorm:"index;not null"`
	SellOrderID int     `json:"sell_order_id" gorm:"index;not null"`
	BuyerID     int     `json:"buyer_id" gorm:"index;not null"`
	SellerID    int     `json:"seller_id" gorm:"index;not null"`
	Price       float64 `json:"price" gorm:"not null"`
	Qty         float64 `json:"qty" gorm:"not null"`
	QuoteAmount float64 `json:"quote_amount" gorm:"not null"`
	CreatedAt time.Time
}

