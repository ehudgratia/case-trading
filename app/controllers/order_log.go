package controllers

import (
	"case-trading/app/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetMarketTrades(ctx *fiber.Ctx) error {
	marketID, err := strconv.Atoi(ctx.Query("market_id"))
	if err != nil || marketID <= 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid market_id",
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))

	s := repository.GetService()
	trades, err := s.GetMarketTrades(ctx.Context(), marketID, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    trades,
	})
}
