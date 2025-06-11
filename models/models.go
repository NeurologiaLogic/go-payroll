package models

import "time"

// User represents an employee or admin in the system
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"` // hashed
	Role      string    `gorm:"not null"` // "employee" or "admin"
	Salary    float64   `gorm:"default:0"`
	//info
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
}

// Attendance represents a daily attendance record for an employee
type Attendance struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null"`
	Date       time.Time `gorm:"not null;index"` // Must be unique per user+date
	//info
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time
	CreatedBy  uint
	IPAddress  string
	PayrollProcessedID uint // Reference to all the attendence records for a period
}

// Overtime represents additional hours worked by an employee
type Overtime struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Date      time.Time `gorm:"index;not null"`
	Hours     float64   `gorm:"not null"`
	//info
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
	IPAddress string
	PayrollProcessedID uint // Reference to all the overtime attendence records for a period
}

// Reimbursement represents an expense claim by an employee
type Reimbursement struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Amount    float64   `gorm:"not null"`
	Desc      string
	//info
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time
	CreatedBy uint
	UpdatedBy uint
	IPAddress string
	PayrollProcessedID uint // Reference to all the reimbursement records for a period
}

//logging all the requests to the API
type AuditLog struct {
	RequestID string `gorm:"type:uuid"`
	Endpoint     string    // e.g. "attendance", "payroll"
	UserID     uint
	IPAddress  string
	//info
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

//run this function every day to log the daily payroll using cron job (not implemented here)
type DailyPayroll struct {
	ID                uint      `gorm:"primaryKey"`
	UserID            uint      `gorm:"index"`
	PayrollProcessedID uint     `gorm:"index"`
	TotalAttendance   int
	BaseSalary        float64
	TotalOvertime     float64
	OvertimePay       float64
	ReimbursementTotal float64
	TakeHomePay       float64
	//info
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AttendancePeriod represents a period for which attendance and payroll are processed
type PayrollProcessed struct {
	ID                uint      `gorm:"primaryKey"`
	Date							time.Time `gorm:"not null;index"` // e.g. "2023-10-01 to 2023-10-15"
	//info
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time
	CreatedBy         uint
	UpdatedBy         uint
	IPAddress         string
}