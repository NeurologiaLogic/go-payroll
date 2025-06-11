// routes/routes.go
package routes

import (
	"go-payroll/controllers"
	"go-payroll/middleware"
	"time"

	"github.com/gofiber/fiber/v2/middleware/cache"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
    // Grouping API
    api := app.Group("/api")
		// Admin Routes
		admin := api.Group("/admin",middleware.JWTProtected("admin"))
		employee := api.Group("/employee",middleware.JWTProtected("employee"))
		cache:=cache.New(cache.Config{
			Expiration: 5 * time.Minute,
		})
    // Health check route (optional)
    api.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("API is running")
    })

		// Authentication Routes
		api.Post("/login", controllers.Login)

    // Attendance Routes
    employee.Post("/attendance", controllers.SubmitAttendance)
    // Overtime Routes
    employee.Post("/overtime", controllers.SubmitOvertime)
    // Reimbursement Routes
    employee.Post("/reimbursement", controllers.SubmitReimbursement)
		// Generate payslip for an employee
		employee.Get("/payslip", cache, controllers.GeneratePayslip)


    // Attendance Period Routes
		admin.Post("/attendance-period", controllers.CreateAttendancePeriod)
		//Generate payslip summary for all employees
		admin.Get("/payslip-summary", controllers.PayslipSummary)
		// Process payroll
		admin.Post("/run-payroll",  cache, controllers.RunPayroll)

}