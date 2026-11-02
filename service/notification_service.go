package service

import (
	"fmt"
	"internship-telkom/dto"
	"internship-telkom/model"
	"internship-telkom/repository"
	"time"
)

type NotificationService interface {
	GetNotifications(userID uint) (dto.NotificationListResponse, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
	CreateNotification(userID uint, applicationID uint, description string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo}
}

func (s *notificationService) GetNotifications(userID uint) (dto.NotificationListResponse, error) {
	notifs, err := s.repo.FindByUserID(userID)
	if err != nil {
		return dto.NotificationListResponse{}, fmt.Errorf("failed to get notifications: %w", err)
	}

	unread, err := s.repo.CountUnread(userID)
	if err != nil {
		return dto.NotificationListResponse{}, fmt.Errorf("failed to count unread: %w", err)
	}

	items := make([]dto.NotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		items = append(items, dto.NotificationResponse{
			ID:                      n.ID,
			Description:             n.Description,
			IsRead:                  n.IsRead,
			InternshipApplicationID: n.InternshipApplicationID,
			CreatedAt:               formatTimeAgo(n.CreatedAt),
		})
	}

	return dto.NotificationListResponse{
		Unread: int(unread),
		Items:  items,
	}, nil
}

func (s *notificationService) MarkAsRead(id uint, userID uint) error {
	return s.repo.MarkAsRead(id, userID)
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *notificationService) CreateNotification(userID uint, applicationID uint, description string) error {
	notif := model.Notification{
		UserID:                  userID,
		InternshipApplicationID: applicationID,
		Description:             description,
		IsRead:                  false,
	}
	_, err := s.repo.Create(notif)
	return err
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Baru saja"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%d menit yang lalu", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d jam yang lalu", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d hari yang lalu", days)
	default:
		return t.Format("2 Jan 2006")
	}
}