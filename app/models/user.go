package models

import (
	"time"
)

type Users struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	Username  string     `json:"username" gorm:"not null;unique"`
	Email     string     `json:"email" gorm:"not null;unique;type:varchar(100)"`
	Password  string     `json:"password" gorm:"not null;type:varchar(100)"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success   bool      `json:"success"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}
