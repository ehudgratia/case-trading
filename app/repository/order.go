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
	// validasi
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

	// transaksi
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// cek market
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

	// nentukan asset dan lock
	var lockAsset string
	var lockAmount float64

	if side == string(models.SideBuy) {
		lockAsset = market.QuoteAsset
		lockAmount = input.Price * input.Quantity
	} else {
		lockAsset = market.BaseAsset
		lockAmount = input.Quantity
	}

	// ambil wallet
	var wallet models.Wallets
	if err := tx.
		Where("user_id = ? AND asset = ?", userID, lockAsset).
		First(&wallet).Error; err != nil {

		tx.Rollback()
		return nil, fmt.Errorf("wallet %s not found", lockAsset)
	}

	// cek saldo available
	if wallet.Available < lockAmount {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient available balance")
	}

	// lock saldo
	if err := tx.Model(&wallet).Updates(map[string]interface{}{
		"available": wallet.Available - lockAmount,
		"locked":    wallet.Locked + lockAmount,
	}).Error; err != nil {

		tx.Rollback()
		return nil, err
	}

	// save order
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

	// matching
	if err := s.MatchOrder(tx, &order, market); err != nil {
		tx.Rollback()
		return nil, err
	}

	// commit
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// respon
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

func (s *Service) MatchOrder(tx *gorm.DB, order *models.Order, market models.Market) error {
	var counter models.Order

	if order.Side == models.SideBuy {
		// last sell
		err := tx.
			Where("market_id = ? AND side = ? AND status = ? AND price <= ?",
				order.MarketID,
				models.SideSell,
				models.OrderStatusOpen,
				order.Price,
			).
			Order("created_at ASC").
			First(&counter).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // tidak ada match
			}
			return err
		}

	} else {
		// last buy
		err := tx.
			Where("market_id = ? AND side = ? AND status = ? AND price >= ?",
				order.MarketID,
				models.SideBuy,
				models.OrderStatusOpen,
				order.Price,
			).
			Order("created_at ASC").
			First(&counter).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}
	}

	// ================= FULL MATCH (PHASE 2) =================
	matchQty := order.Quantity
	tradePrice := counter.Price

	// update incoming order
	if err := tx.Model(order).Updates(map[string]interface{}{
		"filled_qty": matchQty,
		"status":     models.OrderStatusFilled,
	}).Error; err != nil {
		return err
	}

	// update counter order
	if err := tx.Model(&counter).Updates(map[string]interface{}{
		"filled_qty": matchQty,
		"status":     models.OrderStatusFilled,
	}).Error; err != nil {
		return err
	}

	// ================= SETTLEMENT =================
	if order.Side == models.SideBuy {
		// BUY (order) vs SELL (counter)
		if err := s.settleTrade(tx, *order, counter, tradePrice, matchQty, market); err != nil {
			return err
		}
	} else {
		// SELL (order) vs BUY (counter)
		if err := s.settleTrade(tx, counter, *order, tradePrice, matchQty, market); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) settleTrade(tx *gorm.DB, buyOrder models.Order, sellOrder models.Order, price float64, qty float64, market models.Market) error {

	quoteAmount := price * qty

	// ===== BUYER =====
	var buyerQuote, buyerBase models.Wallets
	tx.Where("user_id = ? AND asset = ?", buyOrder.UserID, market.QuoteAsset).First(&buyerQuote)
	tx.Where("user_id = ? AND asset = ?", buyOrder.UserID, market.BaseAsset).First(&buyerBase)

	tx.Model(&buyerQuote).Updates(map[string]interface{}{
		"locked": buyerQuote.Locked - quoteAmount,
	})
	tx.Model(&buyerBase).Updates(map[string]interface{}{
		"available": buyerBase.Available + qty,
	})

	// ===== SELLER =====
	var sellerBase, sellerQuote models.Wallets
	tx.Where("user_id = ? AND asset = ?", sellOrder.UserID, market.BaseAsset).First(&sellerBase)
	tx.Where("user_id = ? AND asset = ?", sellOrder.UserID, market.QuoteAsset).First(&sellerQuote)

	tx.Model(&sellerBase).Updates(map[string]interface{}{
		"locked": sellerBase.Locked - qty,
	})
	tx.Model(&sellerQuote).Updates(map[string]interface{}{
		"available": sellerQuote.Available + quoteAmount,
	})

	return nil
}
