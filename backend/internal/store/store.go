package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"leave-management/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound           = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInsufficientLeave  = errors.New("not enough leave balance")
	ErrOverlappingRequest = errors.New("leave dates overlap an existing request")
	ErrLeaveNotFound      = errors.New("leave request not found")
	ErrLeaveNotPending    = errors.New("leave request is not pending")
	ErrInvalidEmployee    = errors.New("leave request does not belong to an employee")
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) UserByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *Store) UserByID(ctx context.Context, userID int64) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *Store) CreateUser(ctx context.Context, user *model.User) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	err := s.db.WithContext(ctx).Create(user).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailExists
	}
	return err
}

func (s *Store) CreateLeave(ctx context.Context, leave *model.LeaveRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, leave.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		var pendingRequests []model.LeaveRequest
		if err := tx.Where("user_id = ? AND status = ?", leave.UserID, model.StatusPending).
			Find(&pendingRequests).Error; err != nil {
			return err
		}

		availableBalance := user.LeaveBalance
		for _, pending := range pendingRequests {
			availableBalance -= pending.NumberOfDays
		}
		if leave.NumberOfDays > availableBalance {
			return ErrInsufficientLeave
		}

		var overlappingRequests int64
		if err := tx.Model(&model.LeaveRequest{}).
			Where("user_id = ? AND status IN ? AND start_date <= ? AND end_date >= ?",
				leave.UserID,
				[]model.LeaveStatus{model.StatusPending, model.StatusApproved},
				leave.EndDate,
				leave.StartDate,
			).
			Count(&overlappingRequests).Error; err != nil {
			return err
		}
		if overlappingRequests > 0 {
			return ErrOverlappingRequest
		}

		return tx.Create(leave).Error
	})
}

func (s *Store) UserLeaves(ctx context.Context, userID int64) ([]model.LeaveRequest, error) {
	var leaves []model.LeaveRequest
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&leaves).Error
	return leaves, err
}

func (s *Store) PendingLeaves(ctx context.Context) ([]model.LeaveRequest, error) {
	var leaves []model.LeaveRequest
	err := s.db.WithContext(ctx).
		Preload("User").
		Where("status = ?", model.StatusPending).
		Order("created_at DESC").
		Find(&leaves).Error
	if err != nil {
		return nil, err
	}

	for i := range leaves {
		leaves[i].EmployeeName = leaves[i].User.Name
		leaves[i].EmployeeEmail = leaves[i].User.Email
		leaves[i].EmployeeLeaveBalance = &leaves[i].User.LeaveBalance
		balanceAfterApproval := leaves[i].User.LeaveBalance - leaves[i].NumberOfDays
		leaves[i].BalanceAfterApproval = &balanceAfterApproval
	}
	return leaves, nil
}

func (s *Store) ApproveLeave(ctx context.Context, leaveID, adminID int64) (model.LeaveRequest, float64, error) {
	var leave model.LeaveRequest
	var remainingBalance float64

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&leave, leaveID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLeaveNotFound
			}
			return err
		}
		if leave.Status != model.StatusPending {
			return ErrLeaveNotPending
		}

		var employee model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&employee, leave.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if employee.Role != model.RoleEmployee {
			return ErrInvalidEmployee
		}
		if employee.LeaveBalance < leave.NumberOfDays {
			return ErrInsufficientLeave
		}

		remainingBalance = employee.LeaveBalance - leave.NumberOfDays
		if err := tx.Model(&employee).Update("leave_balance", remainingBalance).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		leave.Status = model.StatusApproved
		leave.ReviewedBy = &adminID
		leave.ReviewedAt = &now
		return tx.Model(&leave).Updates(map[string]any{
			"status":      model.StatusApproved,
			"reviewed_by": adminID,
			"reviewed_at": now,
		}).Error
	})

	return leave, remainingBalance, err
}

func (s *Store) RejectLeave(ctx context.Context, leaveID, adminID int64) (model.LeaveRequest, error) {
	var leave model.LeaveRequest

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&leave, leaveID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLeaveNotFound
			}
			return err
		}
		if leave.Status != model.StatusPending {
			return ErrLeaveNotPending
		}

		now := time.Now().UTC()
		leave.Status = model.StatusRejected
		leave.ReviewedBy = &adminID
		leave.ReviewedAt = &now
		return tx.Model(&leave).Updates(map[string]any{
			"status":      model.StatusRejected,
			"reviewed_by": adminID,
			"reviewed_at": now,
		}).Error
	})

	return leave, err
}

func (s *Store) Dashboard(ctx context.Context) (model.Dashboard, error) {
	var dashboard model.Dashboard
	db := s.db.WithContext(ctx)

	if err := db.Model(&model.User{}).
		Where("role = ?", model.RoleEmployee).
		Count(&dashboard.TotalEmployees).Error; err != nil {
		return dashboard, err
	}
	if err := db.Model(&model.LeaveRequest{}).
		Count(&dashboard.TotalRequests).Error; err != nil {
		return dashboard, err
	}
	if err := db.Model(&model.LeaveRequest{}).
		Where("status = ?", model.StatusPending).
		Count(&dashboard.PendingRequests).Error; err != nil {
		return dashboard, err
	}
	if err := db.Model(&model.LeaveRequest{}).
		Where("status = ?", model.StatusApproved).
		Count(&dashboard.ApprovedRequests).Error; err != nil {
		return dashboard, err
	}
	if err := db.Model(&model.LeaveRequest{}).
		Where("status = ?", model.StatusRejected).
		Count(&dashboard.RejectedRequests).Error; err != nil {
		return dashboard, err
	}

	return dashboard, nil
}
