package repository

import (
	"case-trading/app/models"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Service) CreateOrder(ctx context.Context, userID int, input models.OrderRequest) (*models.Order, error) {
	// ===== VALIDASI DASAR =====
	if input.MarketID == 0 {
		return nil, fmt.Errorf("market_id is required")
	}
	if input.Price <= 0 || input.Quantity <= 0 {
		return nil, fmt.Errorf("price and quantity must be greater than zero")
	}

	side := strings.ToUpper(strings.TrimSpace(input.Side))
	if side != string(models.SideBuy) && side != string(models.SideSell) {
		return nil, fmt.Errorf("invalid side")
	}

	// ===== CEK MARKET =====
	var market models.Market
	if err := s.DB.WithContext(ctx).
		Where("id = ? AND is_active = ?", input.MarketID, true).
		First(&market).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("market not found")
		}
		return nil, err
	}

	// ===== CEK WALLET =====
	var asset string
	if side == string(models.SideBuy) {
		asset = market.QuoteAsset
	} else {
		asset = market.BaseAsset
	}

	var wallet models.Wallets
	if err := s.DB.WithContext(ctx).
		Where("user_id = ? AND asset = ?", userID, asset).
		First(&wallet).Error; err != nil {
		return nil, fmt.Errorf("wallet %s not found", asset)
	}

	// ===== CEK SALDO =====
	var cost float64
	if side == string(models.SideBuy) {
		cost = input.Price * input.Quantity
	} else {
		cost = input.Quantity
	}

	if wallet.Amount < cost {
		return nil, fmt.Errorf("insufficient balance")
	}

	// ===== UPDATE SALDO =====
	if err := s.DB.WithContext(ctx).
		Model(&wallet).
		Update("amount", wallet.Amount-cost).Error; err != nil {
		return nil, err
	}

	// ===== TAMBAH SALDO DIBELI =====
	var targetAsset string
	if side == string(models.SideBuy) {
		targetAsset = market.BaseAsset
	} else {
		targetAsset = market.QuoteAsset
	}

	var targetWallet models.Wallets
	if err := s.DB.WithContext(ctx).
		Where("user_id = ? AND asset = ?", userID, targetAsset).
		First(&targetWallet).Error; err != nil {
		return nil, fmt.Errorf("wallet %s not found", targetAsset)
	}

	addAmount := input.Quantity
	if side == string(models.SideSell) {
		addAmount = input.Price * input.Quantity
	}

	if err := s.DB.WithContext(ctx).
		Model(&targetWallet).
		Update("amount", targetWallet.Amount+addAmount).Error; err != nil {
		return nil, err
	}

	// ===== SIMPAN ORDER =====
	order := models.Order{
		UserID:    userID,
		MarketID:  input.MarketID,
		Side:      models.SideType(side),
		Price:     input.Price,
		Quantity:  input.Quantity,
		Status:    "DONE",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}
