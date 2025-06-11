package config

import (
	"fmt"
	"go-payroll/models"
	"go-payroll/seed"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(dsn string) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }

    DB = db
		// Log the connection
		fmt.Println("Connected to database successfully")

		//drop all tables if you want to reset the database
		// err = db.Migrator().DropTable(&models.User{}, &models.Attendance{}, &models.PayrollProcessed{}, &models.AuditLog{}, &models.Overtime{}, &models.Reimbursement{})
		// if err != nil {
		// 	panic("failed to drop tables: " + err.Error())
		// }

    // Auto-migrate your models here
    AutoMigrate()
		err = seed.SeedUsers(db)
		if err != nil {
			panic("failed to seed users: " + err.Error())
		}
}

func AutoMigrate() {
    err := DB.AutoMigrate(
        &models.User{},
				&models.Attendance{},
				&models.PayrollProcessed{},
				&models.AuditLog{},
				&models.Overtime{},
				&models.Reimbursement{},
        // Add other models here
    )
    if err != nil {
        panic("failed to auto migrate models: " + err.Error())
    }
		fmt.Println("Database migration completed successfully")
}
