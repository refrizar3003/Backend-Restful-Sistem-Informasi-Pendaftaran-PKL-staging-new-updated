package repository

import (
	"fmt"
	"internship-telkom/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationRepository interface {
	Create(notif model.Notification) (model.Notification, error)
	FindByUserID(userID uint) ([]model.Notification, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
	CountUnread(userID uint) (int64, error)
	FindByID(id uint) (model.Notification, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db}
}

func (r *notificationRepository) Create(notif model.Notification) (model.Notification, error) {
	if err := r.db.Omit(clause.Associations).Create(&notif).Error; err != nil {
		return model.Notification{}, fmt.Errorf("failed to create notification: %w", err)
	}
	return notif, nil
}

func (r *notificationRepository) FindByUserID(userID uint) ([]model.Notification, error) {
	var notifs []model.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find notifications for user %d: %w", userID, err)
	}
	return notifs, nil
}

func (r *notificationRepository) MarkAsRead(id uint, userID uint) error {
	result := r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notifikasi dengan id %d tidak ditemukan", id)
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID uint) error {
	result := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", result.Error)
	}
	return nil
}

func (r *notificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

func (r *notificationRepository) FindByID(id uint) (model.Notification, error) {
	var notif model.Notification
	if err := r.db.First(&notif, id).Error; err != nil {
		return model.Notification{}, fmt.Errorf("notification with id %d not found: %w", id, err)
	}
	return notif, nil
}