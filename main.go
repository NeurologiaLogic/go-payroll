// main.go
package main

import (
	"go-payroll/config"
	"go-payroll/routes"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
    godotenv.Load()
    config.ConnectDB(os.Getenv("DB_DSN"))
    app := fiber.New()
		// Fiber logger middleware
		app.Use(logger.New(logger.Config{
				TimeFormat: time.RFC3339,
				Format:     "[${time}] ${latency} ${method} ${path}\n",
		}))
    routes.Setup(app)
    app.Listen(":3000")
}
