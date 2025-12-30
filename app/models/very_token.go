package models

import "time"

type VerifyToken struct {
	User_ID  int    `json:"user_id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"type:varchar(100);not null"`
	Email    string `json:"email" gorm:"type:varchar(150);not null"`
	Token    string `json:"token" gorm:"type:varchar(255);not null"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	ExpiredAt time.Time `json:"expired_at" gorm:"not null"`
}
