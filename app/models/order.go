package models

import "time"

type SideType string

const (
	SideBuy  SideType = "BUY"
	SideSell SideType = "SELL"
)

type Order struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null; index"`
	MarketID  int       `json:"market_id" gorm:"not null; index"`
	Side      SideType  `json:"side" gorm:"type:varchar(4); not null"`
	Price     float64   `json:"price" gorm:"not null"`
	Quantity  float64   `json:"quantity" gorm:"not null"`
	Status    string    `json:"status" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at"`

	User   Users  `gorm:"foreignKey:UserID;references:ID"`
	Market Market `gorm:"foreignKey:MarketID;references:ID"`
}

type OrderRequest struct {
	MarketID int     `json:"market_id"`
	Side     string  `json:"side"`
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}
