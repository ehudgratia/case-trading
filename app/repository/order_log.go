package repository

import (
	"case-trading/app/models"
	"context"
	"time"

	"gorm.io/gorm"
)

func createTradeLog(tx *gorm.DB, marketID int, buyOrder models.Order, sellOrder models.Order, price, qty float64) error {

	trade := models.OrderTrade{
		MarketID:    marketID,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID:     buyOrder.UserID,
		SellerID:    sellOrder.UserID,
		Price:       price,
		Qty:         qty,
		QuoteAmount: price * qty,
		CreatedAt:   time.Now().UTC(),
	}

	return tx.Create(&trade).Error
}

func (s *Service) GetMarketTrades(ctx context.Context, marketID int, limit int) ([]models.OrderTrade, error) {

	if limit <= 0 {
		limit = 50
	}

	var trades []models.OrderTrade
	err := s.DB.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("created_at DESC").
		Limit(limit).
		Find(&trades).Error

	return trades, err
}
