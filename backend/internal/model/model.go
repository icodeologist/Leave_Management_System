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

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Email        string    `json:"email" gorm:"not null;uniqueIndex"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role" gorm:"not null;index;check:role IN ('EMPLOYEE','ADMIN')"`
	LeaveBalance int       `json:"leave_balance" gorm:"not null;check:leave_balance >= 0"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
}

type LeaveRequest struct {
	ID           int64       `json:"id" gorm:"primaryKey"`
	UserID       int64       `json:"user_id" gorm:"not null;index"`
	User         User        `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	LeaveType    string      `json:"leave_type" gorm:"not null"`
	StartDate    time.Time   `json:"start_date" gorm:"type:date;not null"`
	EndDate      time.Time   `json:"end_date" gorm:"type:date;not null"`
	NumberOfDays int         `json:"number_of_days" gorm:"not null;check:number_of_days > 0"`
	Reason       string      `json:"reason" gorm:"not null"`
	Status       LeaveStatus `json:"status" gorm:"not null;default:PENDING;index;check:status IN ('PENDING','APPROVED','REJECTED')"`
	ReviewedBy   *int64      `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time  `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at" gorm:"not null"`
}
