package infrastructure

import (
	"fmt"
	"internship-telkom/model"
	"log"
	"os"
	"time"
	_ "time/tzdata" // Embed timezone data untuk kompatibilitas lintas platform

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	if err := validateDBEnv(); err != nil {
	return nil, err
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// SETTING CONNECTION POOL
	if err := configureConnectionPool(db); err != nil {
		return nil, err
	}

	log.Println("Running database migration...")
	if err := runMigrations(db); err != nil {
		return nil, err
	}
		
	log.Println("Successfully connected to the database and migrated tables!")
	return db, nil
}

func validateDBEnv() error {
	required := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_PORT":     os.Getenv("DB_PORT"),
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("environment variable %s is not set", key)
		}
	}
	return nil
}

func configureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func runMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.User{},
		&model.UserToken{},
		&model.TrustedDevice{},
		&model.Unit{},
		&model.Mentor{},
		&model.InternshipApplication{},
		&model.Document{},
		&model.Letter{},
		&model.Notification{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}