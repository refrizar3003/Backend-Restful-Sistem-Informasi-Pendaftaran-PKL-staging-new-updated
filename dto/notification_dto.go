package dto

type NotificationResponse struct {
	ID                      uint   `json:"id"`
	Description             string `json:"description"`
	IsRead                  bool   `json:"is_read"`
	InternshipApplicationID uint   `json:"internship_application_id"`
	CreatedAt               string `json:"created_at"`
}

type NotificationListResponse struct {
	Unread int                    `json:"unread"`
	Items  []NotificationResponse `json:"items"`
}