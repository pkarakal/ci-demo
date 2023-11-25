package models

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&User{}, &Todo{})
	if err != nil {
		zap.L().Error("Encountered an error while migrating database", zap.Error(err))
		return err
	}
	zap.L().Debug("Successfully executed database migrations")
	return nil
}
