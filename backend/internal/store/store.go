package store

import (
	"context"
	"errors"
	"strings"

	"leave-management/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound           = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInsufficientLeave  = errors.New("not enough leave balance")
	ErrOverlappingRequest = errors.New("leave dates overlap an existing request")
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
	}
	return leaves, nil
}
