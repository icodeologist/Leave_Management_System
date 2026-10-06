package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"leave-management/internal/database"
	"leave-management/internal/model"
)

func TestCreateLeaveWithAvailableBalance(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Name:         "Balance Test Employee",
		Email:        "balance-test-" + time.Now().Format("20060102150405.000000000") + "@example.com",
		Role:         model.RoleEmployee,
		LeaveBalance: 5,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&model.LeaveRequest{})
		db.Delete(&user)
	})

	startDate := time.Now().UTC().AddDate(0, 0, 2).Format(model.DateFormat)
	leave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeCasual,
		DayType:      model.DayTypeFull,
		StartDate:    startDate,
		EndDate:      startDate,
		NumberOfDays: 1,
		Reason:       "Balance test",
		Status:       model.StatusPending,
	}

	if err := New(db).CreateLeave(context.Background(), &leave); err != nil {
		t.Fatalf("expected leave request to be created: %v", err)
	}
	if leave.ID == 0 {
		t.Fatal("expected created leave request to have an ID")
	}
	if leave.Status != model.StatusPending {
		t.Fatalf("expected status PENDING, got %s", leave.Status)
	}
}

func TestCreateLeaveRejectsInsufficientBalance(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Name:         "Insufficient Balance Employee",
		Email:        "insufficient-test-" + time.Now().Format("20060102150405.000000000") + "@example.com",
		Role:         model.RoleEmployee,
		LeaveBalance: 1,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&model.LeaveRequest{})
		db.Delete(&user)
	})

	startDate := time.Now().UTC().AddDate(0, 0, 3).Format(model.DateFormat)
	leave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeAnnual,
		DayType:      model.DayTypeFull,
		StartDate:    startDate,
		EndDate:      startDate,
		NumberOfDays: 2,
		Reason:       "Insufficient balance test",
		Status:       model.StatusPending,
	}

	err = New(db).CreateLeave(context.Background(), &leave)
	if !errors.Is(err, ErrInsufficientLeave) {
		t.Fatalf("expected insufficient balance error, got %v", err)
	}

	var saved int64
	if err := db.Model(&model.LeaveRequest{}).Where("user_id = ?", user.ID).Count(&saved).Error; err != nil {
		t.Fatal(err)
	}
	if saved != 0 {
		t.Fatalf("expected no leave request to be saved, got %d", saved)
	}
}
