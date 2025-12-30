package models

import (
	"time"
)

type Wallets struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	UserID    int        `json:"user_id" gorm:"not null;index"`
	Asset     string     `json:"asset" gorm:"not null"`
	Amount    float64    `json:"amount" gorm:"type:decimal(15,2); default:0"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	User Users `gorm:"foreignKey:UserID;references:ID"`
}

type AddWallet struct {
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
}

type TopUpWallet struct {
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
}

type WalletsData struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Asset    string  `json:"asset"`
	Amount   float64 `json:"amount"`

	CreatedAt time.Time `json:"created_at"`
}

type WalletResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *WalletsData `json:"data"`
}

// type WalletTransaction struct {
// 	ID        int             `json:"id" gorm:"primaryKey"`
// 	WalletID  int             `json:"wallet_id" gorm:"not null;index"`
// 	Type      string          `json:"type" gorm:"size:20;not null"` // Type: "TOPUP", "WITHDRAW", "BUY_STOCK", "SELL_STOCK"
// 	Amount    decimal.Decimal `json:"amount" gorm:"type:decimal(15,2);not null"`
// 	Status    string          `json:"status" gorm:"size:20;default:'PENDING'"` // Status: "PENDING", "SUCCESS", "FAILED"
// 	CreatedAt time.Time       `json:"created_at"`
// }
