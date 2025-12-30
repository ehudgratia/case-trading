package controllers

import (
	"case-trading/app/models"
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func AddMarket(ctx *fiber.Ctx) error {
	var input models.AddMarket
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	s := repository.GetTransaction()
	defer func() {
		if r := recover(); r != nil {
			err := s.Rollback(r)
			ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
	}()

	market, err := s.AddMarket(ctx.Context(), input)
	if err != nil {
		s.Rollback(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	if err := s.Commit(); err != nil {
		s.Rollback(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(models.MarketRespons{
		Success: true,
		Message: "Market created successfully",
		Data:    market,
	})
}

func GetMarkets(ctx *fiber.Ctx) error {
	s := repository.GetService()

	markets, err := s.GetMarkets(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    markets,
	})
}
