package repository

import (
	"case-trading/app/models"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Service) CreateOrder(ctx context.Context, userID int, input models.OrderRequest) (*models.OrderResponse, error) {
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

	// ================= TRANSACTION =================
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// ===== CEK MARKET =====
	var market models.Market
	if err := tx.
		Where("id = ? AND is_active = ?", input.MarketID, true).
		First(&market).Error; err != nil {

		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("market not found")
		}
		return nil, err
	}

	// ================= TENTUKAN ASSET & JUMLAH LOCK =================
	var lockAsset string
	var lockAmount float64

	if side == string(models.SideBuy) {
		lockAsset = market.QuoteAsset
		lockAmount = input.Price * input.Quantity
	} else {
		lockAsset = market.BaseAsset
		lockAmount = input.Quantity
	}

	// ================= AMBIL WALLET =================
	var wallet models.Wallets
	if err := tx.
		Where("user_id = ? AND asset = ?", userID, lockAsset).
		First(&wallet).Error; err != nil {

		tx.Rollback()
		return nil, fmt.Errorf("wallet %s not found", lockAsset)
	}

	// ================= CEK SALDO AVAILABLE =================
	if wallet.Available < lockAmount {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient available balance")
	}

	// ================= LOCK SALDO =================
	if err := tx.Model(&wallet).Updates(map[string]interface{}{
		"available": wallet.Available - lockAmount,
		"locked":    wallet.Locked + lockAmount,
	}).Error; err != nil {

		tx.Rollback()
		return nil, err
	}

	// ================= SIMPAN ORDER =================
	order := models.Order{
		UserID:    userID,
		MarketID:  input.MarketID,
		Side:      models.SideType(side),
		Price:     input.Price,
		Quantity:  input.Quantity,
		FilledQty: 0,
		Status:    models.OrderStatusOpen,
		CreatedAt: time.Now().UTC(),
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ================= COMMIT =================
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// ================= RESPONSE =================
	resp := &models.OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		MarketID:  order.MarketID,
		Side:      order.Side,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
	}

	return resp, nil
}
