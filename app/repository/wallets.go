package repository

import (
	"case-trading/app/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s Service) CreateWallet(ctx context.Context, IDUser int, input models.CreateWallet) (*models.WalletsData, error) {
	if strings.TrimSpace(input.Asset) == "" {
		return nil, fmt.Errorf("asset is required")
	}

	asset := strings.ToUpper(strings.TrimSpace(input.Asset))

	// cek wallet sudah ada
	var wallet models.Wallets
	err := s.DB.WithContext(ctx).
		Where("user_id = ? AND asset = ?", IDUser, asset).
		First(&wallet).Error

	if err == nil {
		return nil, fmt.Errorf("wallet already exists")
	}

	wallets := models.Wallets{
		UserID:    IDUser,
		Asset:     asset,
		Available: 0,
		Locked:    0,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.WithContext(ctx).Create(&wallets).Error; err != nil {
		return nil, err
	}

	var user models.Users
	if err := s.DB.WithContext(ctx).
		Where("id = ?", IDUser).
		First(&user).Error; err != nil {
		return nil, err
	}

	resp := &models.WalletsData{
		ID:        wallets.ID,
		Username:  user.Username,
		Asset:     wallets.Asset,
		Available: wallet.Available,
		Locked:    wallet.Locked,
		Total:     wallet.Available + wallet.Locked,
		CreatedAt: wallet.CreatedAt,
	}
	return resp, nil
}

func (s Service) TopUpWallet(ctx context.Context, userID int, input models.TopUpWallet) (*models.WalletsData, error) {
	if strings.TrimSpace(input.Asset) == "" {
		return nil, fmt.Errorf("asset is required")
	}

	if input.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	asset := strings.ToUpper(strings.TrimSpace(input.Asset))

	// ambil wallet
	var wallet models.Wallets
	if err := s.DB.WithContext(ctx).
		Where("user_id = ? AND asset = ?", userID, asset).
		First(&wallet).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}

	// update saldo
	newAvailable := wallet.Available + input.Amount

	if err := s.DB.WithContext(ctx).
		Model(&wallet).
		Update("available", newAvailable).Error; err != nil {
		return nil, err
	}

	// ambil user
	var user models.Users
	if err := s.DB.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return nil, err
	}

	resp := &models.WalletsData{
		ID:        wallet.ID,
		Username:  user.Username,
		Asset:     wallet.Asset,
		Available: newAvailable,
		Locked:    wallet.Locked,
		Total:     newAvailable + wallet.Locked,
		CreatedAt: wallet.CreatedAt,
	}

	return resp, nil
}
