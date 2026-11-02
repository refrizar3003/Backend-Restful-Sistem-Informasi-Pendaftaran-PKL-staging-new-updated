package repository

import (
	"fmt"
	"internship-telkom/model"
	"time"

	"gorm.io/gorm"
)

type TrustedDeviceRepository interface {
	Create(device model.TrustedDevice) error
	FindValid(userID uint, tokenHash string, now time.Time) (model.TrustedDevice, error)
	TouchLastUsed(id uint, at time.Time) error
	DeleteByUserID(userID uint) error
	DeleteExpired(before time.Time) error
}

type trustedDeviceRepository struct{ db *gorm.DB }

func NewTrustedDeviceRepository(db *gorm.DB) TrustedDeviceRepository {
	return &trustedDeviceRepository{db}
}

func (r *trustedDeviceRepository) Create(device model.TrustedDevice) error {
	if err := r.db.Create(&device).Error; err != nil {
		return fmt.Errorf("failed to save trusted device: %w", err)
	}
	return nil
}

func (r *trustedDeviceRepository) FindValid(userID uint, tokenHash string, now time.Time) (model.TrustedDevice, error) {
	var d model.TrustedDevice
	err := r.db.Where("user_id = ? AND token_hash = ? AND expires_at > ?", userID, tokenHash, now).
		First(&d).Error
	if err != nil {
		return model.TrustedDevice{}, fmt.Errorf("trusted device not found: %w", err)
	}
	return d, nil
}

func (r *trustedDeviceRepository) TouchLastUsed(id uint, at time.Time) error {
	return r.db.Model(&model.TrustedDevice{}).Where("id = ?", id).Update("last_used_at", at).Error
}

func (r *trustedDeviceRepository) DeleteByUserID(userID uint) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&model.TrustedDevice{}).Error; err != nil {
		return fmt.Errorf("failed to revoke trusted devices: %w", err)
	}
	return nil
}

func (r *trustedDeviceRepository) DeleteExpired(before time.Time) error {
	return r.db.Where("expires_at < ?", before).Delete(&model.TrustedDevice{}).Error
}