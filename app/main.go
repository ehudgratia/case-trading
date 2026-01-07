package main

import (
	"case-trading/app/helper/database"
	"case-trading/app/helper/redis"
	"case-trading/app/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	database.InitDB()
	database.SyncDB()
	if err := redis.InitRedis(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	app := fiber.New()
	app.Use(logger.New())

	api := app.Group("/api")

	routes.SetupRouters(api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Print("Listening on port " + port)
	log.Fatal(app.Listen(":" + port))
}
