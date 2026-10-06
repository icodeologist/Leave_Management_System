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

func TestCreateLeaveCountsExistingPendingBalance(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Name:         "Pending Balance Employee",
		Email:        "pending-balance-test-" + time.Now().Format("20060102150405.000000000") + "@example.com",
		Role:         model.RoleEmployee,
		LeaveBalance: 2,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&model.LeaveRequest{})
		db.Delete(&user)
	})

	store := New(db)
	firstDate := time.Now().UTC().AddDate(0, 0, 4).Format(model.DateFormat)
	firstLeave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeCasual,
		DayType:      model.DayTypeFull,
		StartDate:    firstDate,
		EndDate:      firstDate,
		NumberOfDays: 2,
		Reason:       "First pending request",
		Status:       model.StatusPending,
	}
	if err := store.CreateLeave(context.Background(), &firstLeave); err != nil {
		t.Fatalf("expected first request to be created: %v", err)
	}

	secondDate := time.Now().UTC().AddDate(0, 0, 5).Format(model.DateFormat)
	secondLeave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeCasual,
		DayType:      model.DayTypeFull,
		StartDate:    secondDate,
		EndDate:      secondDate,
		NumberOfDays: 1,
		Reason:       "Second pending request",
		Status:       model.StatusPending,
	}
	if err := store.CreateLeave(context.Background(), &secondLeave); !errors.Is(err, ErrInsufficientLeave) {
		t.Fatalf("expected insufficient balance error, got %v", err)
	}
}

func TestApproveLeaveDeductsBalance(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Name:         "Approval Balance Employee",
		Email:        "approval-balance-test-" + time.Now().Format("20060102150405.000000000") + "@example.com",
		Role:         model.RoleEmployee,
		LeaveBalance: 3,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&model.LeaveRequest{})
		db.Delete(&user)
	})

	date := time.Now().UTC().AddDate(0, 0, 6).Format(model.DateFormat)
	leave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeAnnual,
		DayType:      model.DayTypeFull,
		StartDate:    date,
		EndDate:      date,
		NumberOfDays: 2,
		Reason:       "Approval balance test",
		Status:       model.StatusPending,
	}
	if err := db.Create(&leave).Error; err != nil {
		t.Fatal(err)
	}

	approved, remaining, err := New(db).ApproveLeave(context.Background(), leave.ID, 999)
	if err != nil {
		t.Fatalf("expected approval to succeed: %v", err)
	}
	if approved.Status != model.StatusApproved {
		t.Fatalf("expected status APPROVED, got %s", approved.Status)
	}
	if remaining != 1 {
		t.Fatalf("expected remaining balance 1, got %v", remaining)
	}

	var updated model.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.LeaveBalance != 1 {
		t.Fatalf("expected stored balance 1, got %v", updated.LeaveBalance)
	}
}

func TestApproveHalfDayLeaveDeductsHalfDay(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Name:         "Half Day Balance Employee",
		Email:        "half-day-balance-test-" + time.Now().Format("20060102150405.000000000") + "@example.com",
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

	date := time.Now().UTC().AddDate(0, 0, 7).Format(model.DateFormat)
	leave := model.LeaveRequest{
		UserID:       user.ID,
		LeaveType:    model.LeaveTypeSick,
		DayType:      model.DayTypeHalf,
		StartDate:    date,
		EndDate:      date,
		NumberOfDays: 0.5,
		Reason:       "Half day balance test",
		Status:       model.StatusPending,
	}
	if err := db.Create(&leave).Error; err != nil {
		t.Fatal(err)
	}

	_, remaining, err := New(db).ApproveLeave(context.Background(), leave.ID, 999)
	if err != nil {
		t.Fatalf("expected half-day approval to succeed: %v", err)
	}
	if remaining != 0.5 {
		t.Fatalf("expected remaining balance 0.5, got %v", remaining)
	}

	var updated model.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.LeaveBalance != 0.5 {
		t.Fatalf("expected stored balance 0.5, got %v", updated.LeaveBalance)
	}
}
