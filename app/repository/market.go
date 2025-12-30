package repository

import (
	"case-trading/app/models"
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Service) AddMarket(ctx context.Context, input models.AddMarket) (*models.MarketData, error) {
	if strings.TrimSpace(input.BaseAsset) == "" || strings.TrimSpace(input.QuoteAsset) == "" {
		return nil, fmt.Errorf("base_asset and quote_asset are required")
	}

	base := strings.ToUpper(strings.TrimSpace(input.BaseAsset))
	quote := strings.ToUpper(strings.TrimSpace(input.QuoteAsset))

	if base == quote {
		return nil, fmt.Errorf("base_asset and quote_asset must be different")
	}

	// cek market sudah ada
	var existing models.Market
	err := s.DB.WithContext(ctx).
		Where("base_asset = ? AND quote_asset = ?", base, quote).
		First(&existing).Error

	if err == nil {
		return nil, fmt.Errorf("market already exists")
	}

	market := models.Market{
		BaseAsset:  base,
		QuoteAsset: quote,
		IsActive:   true,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.DB.WithContext(ctx).Create(&market).Error; err != nil {
		return nil, err
	}

	resp := &models.MarketData{
		ID:         market.ID,
		BaseAsset:  market.BaseAsset,
		QuoteAsset: market.QuoteAsset,
		IsActive:   market.IsActive,
		CreatedAt:  market.CreatedAt,
		UpdatedAt:  market.UpdatedAt,
	}

	return resp, nil
}

func (s *Service) GetMarkets(ctx context.Context) ([]models.MarketData, error) {
	var markets []models.Market

	if err := s.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Order("id ASC").
		Find(&markets).Error; err != nil {
		return nil, err
	}

	result := make([]models.MarketData, 0, len(markets))
	for _, m := range markets {
		result = append(result, models.MarketData{
			ID:         m.ID,
			BaseAsset:  m.BaseAsset,
			QuoteAsset: m.QuoteAsset,
			IsActive:   m.IsActive,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}

	return result, nil
}
