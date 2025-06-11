// controllers/controllers.go
package controllers

import (
	"go-payroll/config"
	"go-payroll/models"
	"go-payroll/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SubmitAttendance(c *fiber.Ctx) error {
	type payload struct {
		Date string `json:"date"`
	}
	var body payload
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
			"instruction": "Date should be in YYYY-MM-DD format",
		})
	}

	user,err := GetUserProfile(c)
	if err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
	}

	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot submit attendance on weekends"})
	}

	exists := models.Attendance{}
	err = config.DB.Where("user_id = ? AND date = ?", user.ID, date).First(&exists).Error
	if err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Attendance already submitted for this date"})
	}

	attendance := models.Attendance{
		UserID:    user.ID,
		Date:      date,
		CreatedBy: user.ID,
		IPAddress: utils.GetIPAddress(c),
	}
	if err := config.DB.Create(&attendance).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not save attendance"})
	}

	return c.JSON(fiber.Map{"message": "Attendance submitted"})
}

func SubmitOvertime(c *fiber.Ctx) error {
	type payload struct {
		Date  string  `json:"date"`
		Hours float64 `json:"hours"`
	}
	var body payload
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
			"instruction": "Date should be in YYYY-MM-DD format and hours should be between 1 and 3",
		})
	}

	user,err := GetUserProfile(c)
	if err != nil {
		return err
	}
	date, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
	}
	if body.Hours > 3 || body.Hours <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid number of overtime hours (1-3 allowed)"})
	}
	// Optional: check if submission is after 5PM
	if time.Now().Hour() < 17 {
		return c.Status(400).JSON(fiber.Map{"error": "Overtime can only be submitted after working hours (5PM)"})
	}
	overtime := models.Overtime{
		UserID:    user.ID,
		Date:      date,
		Hours:     body.Hours,
		CreatedBy: user.ID,
		IPAddress: utils.GetIPAddress(c),
	}
	if err := config.DB.Create(&overtime).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not save overtime"})
	}

	return c.JSON(fiber.Map{"message": "Overtime submitted"})
}

func SubmitReimbursement(c *fiber.Ctx) error {
	type payload struct {
		Amount float64 `json:"amount"`
		Desc   string  `json:"desc"`
		Date	 string  `json:"date"` // Optional date field
	}
	var body payload
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
			"instruction": "Amount should be a positive number, description should not be empty, and date should be in YYYY-MM-DD format",
		})
	}

	user,err := GetUserProfile(c)
	if err != nil {
		return err
	}

	reimbursement := models.Reimbursement{
		UserID:    user.ID,
		Amount:    body.Amount,
		Desc:      body.Desc,
		CreatedBy: user.ID,
		IPAddress: utils.GetIPAddress(c),
	}
	if err := config.DB.Create(&reimbursement).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not save reimbursement"})
	}

	return c.JSON(fiber.Map{"message": "Reimbursement submitted"})
}



func GeneratePayslip(c *fiber.Ctx) error {
	/**
		RULES:
		- Employees view only their own payslip
		- Daily rate = Base Salary / 20 days
		- Base Salary = Attendance Days * Daily Rate
		- Overtime Pay = 2 * (Daily Rate / 8 hours) * Overtime Hours
		- Reimbursement Total = SUM of reimbursements in the period
		- Take Home Pay = Base Salary + Overtime Pay + Reimbursement
	**/

	user, err := GetUserProfile(c)
	if err != nil {
		return err
	}

	baseSalary := user.Salary
	dailyRate := baseSalary / 20

	// Attendance (unpaid records only: attendance_period_id == 0)
	var attendanceCount int64
	config.DB.Model(&models.Attendance{}).
		Where("user_id = ? AND attendance_period_id = 0", user.ID).
		Count(&attendanceCount)

	monthlySalary := float64(attendanceCount) * dailyRate

	// Overtime
	var totalOvertime float64
	config.DB.Model(&models.Overtime{}).
		Select("COALESCE(SUM(hours), 0)").
		Where("user_id = ? AND attendance_period_id = 0", user.ID).
		Scan(&totalOvertime)

	overtimePay := 2 * (dailyRate / 8) * totalOvertime

	// Reimbursements
	var reimbursementTotal float64
	config.DB.Model(&models.Reimbursement{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND attendance_period_id = 0", user.ID).
		Scan(&reimbursementTotal)

	// Total take home
	takeHomePay := monthlySalary + overtimePay + reimbursementTotal

	return c.JSON(fiber.Map{
		"employee_id":       user.ID,
		"attendance_days":   attendanceCount,
		"daily_rate":        utils.Round(dailyRate),
		"base_salary_total": utils.Round(monthlySalary),
		"base_salary_note":  "Calculated as attendance_days × daily_rate",
		"base_salary_rate":  utils.Round(baseSalary),

		"overtime_hours":    utils.Round(totalOvertime),
		"overtime_pay":      utils.Round(overtimePay),
		"overtime_note":     "Calculated as 2 × (daily_rate ÷ 8) × overtime_hours",

		"reimbursement_total": utils.Round(reimbursementTotal),
		"reimbursement_note":  "Sum of all reimbursements with no attendance period",

		"take_home_pay":    utils.Round(takeHomePay),
		"take_home_note":   "Calculated as base_salary_total + overtime_pay + reimbursement_total",
	})
}
