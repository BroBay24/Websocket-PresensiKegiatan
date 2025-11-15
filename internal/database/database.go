package database

import (
	"fmt"

	"github.com/BroBay24/WebsocketUTS/internal/config"
	"github.com/BroBay24/WebsocketUTS/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	gormCfg := &gorm.Config{}
	if cfg.DB.LogMode {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Attendance{}); err != nil {
		return nil, err
	}

	return db, nil
}
