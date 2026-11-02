package controller

import (
	"internship-telkom/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NotificationController struct {
	srv service.NotificationService
}

func NewNotificationController(srv service.NotificationService) *NotificationController {
	return &NotificationController{srv}
}

// GetNotifications godoc
// @Summary      Get user notifications
// @Description  Retrieve all notifications for the currently logged-in user
// @Tags         Notification
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /notifications [get]
func (h *NotificationController) GetNotifications(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	res, err := h.srv.GetNotifications(userID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil notifikasi")
	}

	return successResponse(c, fiber.StatusOK, "Berhasil mengambil notifikasi", res)
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Tags         Notification
// @Security     ApiKeyAuth
// @Param        id   path  int  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /notifications/{id}/read [patch]
func (h *NotificationController) MarkAsRead(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID notifikasi tidak valid")
	}

	if err := h.srv.MarkAsRead(uint(id), userID); err != nil {
		return errorResponse(c, fiber.StatusNotFound, "Notifikasi tidak ditemukan")
	}

	return successResponse(c, fiber.StatusOK, "Notifikasi berhasil ditandai sebagai dibaca", nil)
}

// MarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Tags         Notification
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /notifications/read-all [patch]
func (h *NotificationController) MarkAllAsRead(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	if err := h.srv.MarkAllAsRead(userID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal menandai semua notifikasi")
	}

	return successResponse(c, fiber.StatusOK, "Semua notifikasi berhasil ditandai sebagai dibaca", nil)
}