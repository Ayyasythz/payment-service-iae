package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"payment-service-iae/config"
	"payment-service-iae/models"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate schemas
	err = db.AutoMigrate(
		&models.Payment{},
		&models.PaymentNotification{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
