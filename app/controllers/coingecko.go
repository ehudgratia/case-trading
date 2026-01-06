package controllers

import (
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func GetMarketsCoin(ctx *fiber.Ctx) error {
	// Ambil query param 'ids', jika kosong gunakan default
	coinIDs := ctx.Query("ids", "bitcoin,ethereum,solana")

	repo := repository.NewMarketRepository()
	data, err := repo.GetLivePrice(ctx.Context(), coinIDs)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data dari CoinGecko",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}
