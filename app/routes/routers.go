package routes

import (
	"case-trading/app/controllers"
	"case-trading/app/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetupRouters(api fiber.Router) {
	public := api.Group("/public")
	{
		// regist
		public.Post("/register", controllers.UserRegister)

		// login
		public.Post("/login", controllers.UserLogin)
	}

	auth := api.Group("/auth", middlewares.AuthMiddleware())
	{
		auth.Post("/wallet", controllers.AddWallet)
		auth.Post("/topup", controllers.TopUpWallet)
	}
}
