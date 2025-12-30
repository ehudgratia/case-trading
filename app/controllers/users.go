package controllers

import (
	"case-trading/app/models"
	"case-trading/app/repository"

	"github.com/gofiber/fiber/v2"
)

func UserRegister(ctx *fiber.Ctx) error {
	var input models.RegisterRequest

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
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
	user, err := s.Register(ctx.Context(), input)
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
		"message": "User created successfully",
		"data":    user,
	})
}

func UserLogin(ctx *fiber.Ctx) error {
	var input models.LoginRequest

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

	loginResp, err := s.Login(ctx.Context(), input)
	if err != nil {
		s.Rollback(err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
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
		"data":    loginResp,
	})
}
