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

		// get market
		public.Get("/", controllers.GetMarkets)
	}

	auth := api.Group("/auth", middlewares.AuthMiddleware())
	{
		// user
		auth.Post("/wallet", controllers.AddWallet)
		auth.Post("/topup", controllers.TopUpWallet)

		// admin (manual)
		auth.Post("/market", controllers.AddMarket)

		// order
		auth.Post("/order", controllers.CreateOrder)

		//order log
		auth.Get("/orderlog", controllers.GetMarketTrades)
	}
}
