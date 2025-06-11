// controllers/auth.go
package controllers

import (
	"go-payroll/config"
	"go-payroll/models"
	"go-payroll/utils"

	"github.com/gofiber/fiber/v2"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user models.User
	result := config.DB.Where("username = ?", input.Username).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
	})
}

func GetUserProfile(c *fiber.Ctx) (*models.User, error) {
	floatUserID, ok := c.Locals("user_id").(float64)
	if !ok {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid user ID in token")
	}
	userID := uint(floatUserID)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	user.Password = ""
	return &user, nil
}
