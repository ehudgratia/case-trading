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

func (s Service) AddWallet(ctx context.Context, IDUser int, input models.AddWallet) (*models.WalletsData, error) {
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
		Amount:    0,
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
		ID:       wallets.ID,
		Username: user.Username,
		Asset:    wallets.Asset,
		Amount:   wallets.Amount,
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
	newAmount := wallet.Amount + input.Amount

	if err := s.DB.WithContext(ctx).
		Model(&wallet).
		Update("amount", newAmount).Error; err != nil {
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
		ID:       wallet.ID,
		Username: user.Username,
		Asset:    wallet.Asset,
		Amount:   newAmount,
	}

	return resp, nil
}
