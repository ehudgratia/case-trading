package controllers

import (
	"case-trading/app/models"
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func CreateOrder(ctx *fiber.Ctx) error {
	var input models.OrderRequest
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	userID := ctx.Locals("id").(int)
	s := repository.GetService()

	order, err := s.CreateOrder(ctx.Context(), userID, input)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    order,
	})
}
