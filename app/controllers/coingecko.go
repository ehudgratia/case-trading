package controllers

import (
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func GetMarketsCoin(ctx *fiber.Ctx) error {
	// Ambil query param 'ids', default: bitcoin
	coinIDs := ctx.Query("ids", "bitcoin,ethereum,solana")

	s := repository.GetTransaction()

	// Kita tidak butuh Begin/Commit untuk hit API luar,
	// tapi tetap ikuti pola strukturmu jika ingin konsisten.

	data, err := s.GetLiveMarketPrice(ctx.Context(), coinIDs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch market data",
			"details": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Market data fetched successfully",
		"data":    data,
	})
}
