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

	side, err := validateOrderInput(input)
	if err != nil {
		return nil, err
	}

	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	market, err := getActiveMarket(tx, input.MarketID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := lockWalletForOrder(tx, userID, side, market, input.Price, input.Quantity); err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err := createOrderEntity(tx, userID, input, side)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.MatchOrder(tx, order, market); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &models.OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		MarketID:  order.MarketID,
		Side:      order.Side,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
	}, nil
}

func validateOrderInput(input models.OrderRequest) (models.SideType, error) {
	if input.MarketID == 0 {
		return "", fmt.Errorf("market_id is required")
	}
	if input.Price <= 0 || input.Quantity <= 0 {
		return "", fmt.Errorf("price and quantity must be greater than zero")
	}

	side := strings.ToUpper(strings.TrimSpace(input.Side))
	if side != string(models.SideBuy) && side != string(models.SideSell) {
		return "", fmt.Errorf("invalid side")
	}

	return models.SideType(side), nil
}

func getActiveMarket(tx *gorm.DB, marketID int) (models.Market, error) {
	var market models.Market
	err := tx.
		Where("id = ? AND is_active = ?", marketID, true).
		First(&market).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return market, fmt.Errorf("market not found")
		}
		return market, err
	}

	return market, nil
}

func lockWalletForOrder(tx *gorm.DB, userID int, side models.SideType, market models.Market, price, qty float64) error {

	var asset string
	var amount float64

	if side == models.SideBuy {
		asset = market.QuoteAsset
		amount = price * qty
	} else {
		asset = market.BaseAsset
		amount = qty
	}

	wallet, err := getOrCreateWallet(tx, userID, asset)
	if err != nil {
		return err
	}

	if wallet.Available < amount {
		return fmt.Errorf("insufficient available balance")
	}

	return tx.Model(wallet).Updates(map[string]interface{}{
		"available": wallet.Available - amount,
		"locked":    wallet.Locked + amount,
	}).Error
}

func createOrderEntity(tx *gorm.DB, userID int, input models.OrderRequest, side models.SideType) (*models.Order, error) {
	order := models.Order{
		UserID:    userID,
		MarketID:  input.MarketID,
		Side:      side,
		Price:     input.Price,
		Quantity:  input.Quantity,
		FilledQty: 0,
		Status:    models.OrderStatusOpen,
		CreatedAt: time.Now().UTC(),
	}

	if err := tx.Create(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
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
		return s.settleTrade(
			tx, *order, counter, tradePrice, matchQty, market,
		)
	}
	return s.settleTrade(tx, counter, *order, tradePrice, matchQty, market)
}

func (s *Service) settleTrade(tx *gorm.DB, buyOrder models.Order, sellOrder models.Order, price, qty float64, market models.Market) error {

	quoteAmount := price * qty

	// BUYER
	buyerQuote, err := getOrCreateWallet(tx, buyOrder.UserID, market.QuoteAsset)
	if err != nil {
		return err
	}
	buyerBase, err := getOrCreateWallet(tx, buyOrder.UserID, market.BaseAsset)
	if err != nil {
		return err
	}

	tx.Model(buyerQuote).Update("locked", buyerQuote.Locked-quoteAmount)
	tx.Model(buyerBase).Update("available", buyerBase.Available+qty)

	// SELLER
	sellerBase, err := getOrCreateWallet(tx, sellOrder.UserID, market.BaseAsset)
	if err != nil {
		return err
	}
	sellerQuote, err := getOrCreateWallet(tx, sellOrder.UserID, market.QuoteAsset)
	if err != nil {
		return err
	}

	tx.Model(sellerBase).Update("locked", sellerBase.Locked-qty)
	tx.Model(sellerQuote).Update("available", sellerQuote.Available+quoteAmount)

	if err := tx.Model(&models.Market{}).
		Where("id = ?", market.ID).
		Update("last_price", price).Error; err != nil {
		return err
	}

	return nil
}

func getOrCreateWallet(tx *gorm.DB, userID int, asset string) (*models.Wallets, error) {

	var wallet models.Wallets
	err := tx.
		Where("user_id = ? AND asset = ?", userID, asset).
		First(&wallet).Error

	if err == nil {
		return &wallet, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// auto create wallet
	wallet = models.Wallets{
		UserID:    userID,
		Asset:     asset,
		Available: 0,
		Locked:    0,
		CreatedAt: time.Now().UTC(),
	}

	if err := tx.Create(&wallet).Error; err != nil {
		return nil, err
	}

	return &wallet, nil
}
