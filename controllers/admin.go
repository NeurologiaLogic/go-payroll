// controllers/controllers.go
package controllers

import (
	"go-payroll/config"
	"go-payroll/models"
	"go-payroll/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Generate PayslipSummary generates a summary of payslips for all employees that have not been processed yet.
func PayslipSummary(c *fiber.Ctx) error {
	type Result struct {
		UserID         uint
		BaseSalary     float64
		AttendanceCount int64
		MonthlySalary float64
		OvertimeHours  float64
		OvertimePay    float64
		Reimbursement  float64
		TakeHomePay    float64
	}

	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch users")
	}

	var results []Result
	for _, user := range users {
		var attendanceCount int64
		var totalOvertime float64
		var totalReimburse float64

		config.DB.Model(&models.Attendance{}).
			Where("user_id = ? AND payroll_processed_id = 0", user.ID).
			Count(&attendanceCount)

		config.DB.Model(&models.Overtime{}).
			Where("user_id = ? AND payroll_processed_id = 0", user.ID).
			Select("COALESCE(SUM(hours), 0)").Scan(&totalOvertime)

		config.DB.Model(&models.Reimbursement{}).
			Where("user_id = ? AND payroll_processed_id = 0", user.ID).
			Select("COALESCE(SUM(amount), 0)").Scan(&totalReimburse)

		monthlySalary := float64(attendanceCount) * user.Salary
		overtimePay := totalOvertime * (user.Salary / 8) // assume 8 hours/day
		takeHome := monthlySalary + overtimePay + totalReimburse

		results = append(results, Result{
			UserID:        user.ID,
			BaseSalary:    user.Salary,
			AttendanceCount: attendanceCount,
			MonthlySalary: monthlySalary,
			OvertimeHours: totalOvertime,
			OvertimePay:   overtimePay,
			Reimbursement: totalReimburse,
			TakeHomePay:   takeHome,
		})
	}

	total := 0.0
	for _, r := range results {
		total += r.TakeHomePay
	}

	return c.JSON(fiber.Map{
		"summary": results,
		"total_take_home_all_employees": utils.Round(total*100) / 100,
		"note": "Sum of unpaid base salary, overtime, and reimbursement",
	})
}


// CreateAttendancePeriod creates attendance records for a specified date range for multiple employees
func CreateAttendancePeriod(c *fiber.Ctx) error {
	type Input struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		// array of employee IDs (optional)
		Employees []uint `json:"employees,omitempty"`
	}
	var body Input
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
			"instruction": "Start date and end date should be in YYYY-MM-DD format, and employees should be an array of user IDs",
		})
	}
	// Validate dates
	if body.StartDate == "" || body.EndDate == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Start date and end date are required")
	}
	// Create attendance records for each employee
	user, err := GetUserProfile(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}
	if len(body.Employees) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No employees provided")
	}

	for _, empID := range body.Employees {

		// Loop through each day in the range and create attendance records
		startDate, err := time.Parse("2006-01-02", body.StartDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid start date format")
		}
		endDate, err := time.Parse("2006-01-02", body.EndDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid end date format")
		}
		for d := startDate; d.Before(endDate.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
			attendance := models.Attendance{
				UserID: empID,
				// Set the date range
				CreatedBy:  user.ID, // Assuming admin ID is 1
				IPAddress:  c.IP(),
				Date: 		 d,
			}
			if err := config.DB.Create(&attendance).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create attendance record")
			}
		}
	}
	return c.JSON(fiber.Map{
		"message": "Attendance period created successfully",
		"start_date": body.StartDate,
		"end_date": body.EndDate,
		"employees": body.Employees,
	})
}

// RunPayroll processes the payroll for all users based on attendance, overtime, and reimbursements
func RunPayroll(c *fiber.Ctx) error {
	// Input validation
	type Input struct {
		Date string `json:"date"` // Using string for date format consistency
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
			"instruction": "Date should be in YYYY-MM-DD format",
		})
	}

	payrollDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid date format")
	}

	// Create PayrollProcessed
	user, err := GetUserProfile(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
	pp := models.PayrollProcessed{
		Date:      payrollDate,
		CreatedBy: user.ID,
		UpdatedBy: user.ID,
		IPAddress: c.IP(),
	}
	if err := config.DB.Create(&pp).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create payroll period")
	}

	// Process per user
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch users")
	}

	for _, u := range users {
		// Count unpaid attendance
		var attendanceCount int64
		config.DB.Model(&models.Attendance{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Count(&attendanceCount)

		// Sum overtime hours
		var overtimeHours float64
		config.DB.Model(&models.Overtime{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Select("COALESCE(SUM(hours), 0)").Scan(&overtimeHours)

		// Sum reimbursements
		var reimburseTotal float64
		config.DB.Model(&models.Reimbursement{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Select("COALESCE(SUM(amount), 0)").Scan(&reimburseTotal)

		// baseSalary := float64(attendanceCount) * u.Salary
		// overtimePay := overtimeHours * (u.Salary / 8) // assume 8 hours/day
		// takeHome := baseSalary + overtimePay + reimburseTotal

		// // Create DailyPayroll record
		// dpr := models.DailyPayroll{
		// 	UserID:             u.ID,
		// 	PayrollProcessedID: pp.ID,
		// 	TotalAttendance:    int(attendanceCount),
		// 	BaseSalary:         baseSalary,
		// 	TotalOvertime:      overtimeHours,
		// 	OvertimePay:        overtimePay,
		// 	ReimbursementTotal: reimburseTotal,
		// 	TakeHomePay:        takeHome,
		// 	CreatedAt:          time.Now(),
		// 	UpdatedAt:          time.Now(),
		// }
		// if err := config.DB.Create(&dpr).Error; err != nil {
		// 	return fiber.NewError(fiber.StatusInternalServerError, "Failed to save payroll summary")
		// }

		// Update references
		config.DB.Model(&models.Attendance{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Update("payroll_processed_id", pp.ID)
		config.DB.Model(&models.Overtime{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Update("payroll_processed_id", pp.ID)
		config.DB.Model(&models.Reimbursement{}).
			Where("user_id = ? AND payroll_processed_id = 0", u.ID).
			Update("payroll_processed_id", pp.ID)
	}

	return c.JSON(fiber.Map{"message": "Payroll processed", "payroll_processed_id": pp.ID})
}
