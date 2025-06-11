package seed

import (
	"fmt"
	"go-payroll/models"
	"go-payroll/utils"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		fmt.Println("Users already seeded")
		return nil
	}

	fmt.Println("Seeding users...")

	rand.Seed(time.Now().UnixNano())

	// Create 100 employees
	for i := 1; i <= 100; i++ {
		username := fmt.Sprintf("employee%03d", i)
		password := "password123"
		salary := float64(rand.Intn(4_000_000) + 3_000_000) // 3m - 7m

		hashedPassword, _ := utils.HashPassword(password)

		user := models.User{
			Username:  username,
			Password:  string(hashedPassword),
			Role:      "employee",
			Salary:    salary,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	// Create 1 admin
	hashedPassword, _ := utils.HashPassword("admin123") // preset password
	admin := models.User{
		Username:  "admin",
		Password:  hashedPassword, // preset
		Role:      "admin",
		Salary:    0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	fmt.Println("Users seeded successfully")
	return nil
}

