package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(path, logLevel string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger:         newGormLogger(logLevel),
		TranslateError: true,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	initialized := false
	defer func() {
		if !initialized {
			_ = sqlDB.Close()
		}
	}()
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("protect database file: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return fmt.Errorf("enable sqlite WAL: %w", err)
	}
	if err := db.Exec("PRAGMA busy_timeout=5000").Error; err != nil {
		return fmt.Errorf("set sqlite busy timeout: %w", err)
	}
	DB = db
	initialized = true
	return nil
}

func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	DB = nil
	return sqlDB.Close()
}
