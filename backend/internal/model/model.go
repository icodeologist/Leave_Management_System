package model

import "time"

type Role string

const (
	RoleEmployee Role = "EMPLOYEE"
	RoleAdmin    Role = "ADMIN"
)

type LeaveStatus string

const (
	StatusPending  LeaveStatus = "PENDING"
	StatusApproved LeaveStatus = "APPROVED"
	StatusRejected LeaveStatus = "REJECTED"
)

type LeaveType string

const (
	LeaveTypeAnnual LeaveType = "ANNUAL"
	LeaveTypeCasual LeaveType = "CASUAL"
	LeaveTypeSick   LeaveType = "SICK"
)

type DayType string

const (
	DayTypeFull DayType = "FULL_DAY"
	DayTypeHalf DayType = "HALF_DAY"
)

const DateFormat = "2006-01-02"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Email        string    `json:"email" gorm:"not null;uniqueIndex"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role" gorm:"not null;index;check:role IN ('EMPLOYEE','ADMIN')"`
	LeaveBalance float64   `json:"leave_balance" gorm:"type:numeric(4,1);not null;check:leave_balance >= 0 AND leave_balance <= 20"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
}

type LeaveRequest struct {
	ID                   int64       `json:"id" gorm:"primaryKey"`
	UserID               int64       `json:"user_id" gorm:"not null;index"`
	User                 User        `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	EmployeeName         string      `json:"employee_name,omitempty" gorm:"-"`
	EmployeeEmail        string      `json:"employee_email,omitempty" gorm:"-"`
	EmployeeLeaveBalance *float64    `json:"leave_balance,omitempty" gorm:"-"`
	BalanceAfterApproval *float64    `json:"balance_after_approval,omitempty" gorm:"-"`
	LeaveType            LeaveType   `json:"leave_type" gorm:"not null;check:leave_type IN ('ANNUAL','CASUAL','SICK')"`
	DayType              DayType     `json:"day_type" gorm:"not null;check:day_type IN ('FULL_DAY','HALF_DAY')"`
	StartDate            string      `json:"start_date" gorm:"type:date;not null"`
	EndDate              string      `json:"end_date" gorm:"type:date;not null"`
	NumberOfDays         float64     `json:"number_of_days" gorm:"type:numeric(4,1);not null;check:number_of_days > 0"`
	Reason               string      `json:"reason" gorm:"not null"`
	Status               LeaveStatus `json:"status" gorm:"not null;default:PENDING;index;check:status IN ('PENDING','APPROVED','REJECTED')"`
	ReviewedBy           *int64      `json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time  `json:"reviewed_at,omitempty"`
	CreatedAt            time.Time   `json:"created_at" gorm:"not null"`
}

type Dashboard struct {
	TotalEmployees   int64 `json:"total_employees"`
	TotalRequests    int64 `json:"total_leave_requests"`
	PendingRequests  int64 `json:"pending_requests"`
	ApprovedRequests int64 `json:"approved_requests"`
	RejectedRequests int64 `json:"rejected_requests"`
}
