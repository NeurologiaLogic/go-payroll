package middleware

import (
	"go-payroll/config"
	"go-payroll/models"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)


func JWTProtected(requiredRole string) fiber.Handler {
	secret := os.Getenv("JWT_SECRET")

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
			})
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		claims := token.Claims.(jwt.MapClaims)

		// Extract role from token and compare
		role, ok := claims["role"].(string)
		if !ok || role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access forbidden: insufficient role",
			})
		}

		c.Locals("user_id", claims["user_id"])
		c.Locals("role", role)
		stringRequestID := uuid.New().String()

		config.DB.Create(&models.AuditLog{
			RequestID:  stringRequestID,
			Endpoint:   c.Path(),
			UserID:     uint(claims["user_id"].(float64)),
			IPAddress:  c.IP(),
			CreatedAt:  time.Now(),
		})
		return c.Next()
	}
}