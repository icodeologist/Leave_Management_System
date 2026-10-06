package database

import (
	"fmt"

	"leave-management/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LeaveRequest{}); err != nil {
		return nil, fmt.Errorf("set up database: %w", err)
	}
	return db, nil
}
