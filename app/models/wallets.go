package models

import (
	"time"
)

type Wallets struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	UserID    int     `json:"user_id" gorm:"not null;index"`
	Asset     string  `json:"asset" gorm:"not null"`
	Available float64 `json:"available" gorm:"type:decimal(20,2); default:0"`
	Locked    float64 `json:"locked" gorm:"type:decimal(20,2); not null; default:0"`

	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	User Users `gorm:"foreignKey:UserID;references:ID"`
}

type CreateWallet struct {
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
}

type TopUpWallet struct {
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
}

type WalletsData struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Asset     string    `json:"asset"`
	Available float64   `json:"available"`
	Locked    float64   `json:"locked"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *WalletsData `json:"data"`
}
