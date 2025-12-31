package controllers

import (
	"case-trading/app/models"
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func AddWallet(ctx *fiber.Ctx) error {
	var input models.CreateWallet
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}
	User_ID := ctx.Locals("id").(int)
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

	wallet, err := s.CreateWallet(ctx.Context(), User_ID, input)
	if err != nil {
		s.Rollback(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Wallet created successfully",
		"data":    wallet,
	})
}

func TopUpWallet(ctx *fiber.Ctx) error {
	var input models.TopUpWallet
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	userID := ctx.Locals("id").(int)

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

	wallet, err := s.TopUpWallet(ctx.Context(), userID, input)
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

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "top up success",
		"data":    wallet,
	})
}
