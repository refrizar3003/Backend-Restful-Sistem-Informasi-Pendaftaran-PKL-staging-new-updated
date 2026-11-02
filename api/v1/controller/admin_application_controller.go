package controller

import (
	"errors"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AdminApplicationController struct {
	srv service.AdminApplicationService
}

func NewAdminApplicationController(srv service.AdminApplicationService) *AdminApplicationController {
	return &AdminApplicationController{srv}
}

// GetApplications godoc
// @Summary      View PKL Registration Data (Admin)
// @Description  Retrieves a list of incoming PKL registrations. Supports status filters (to differentiate between "New Registration" and "Registration History" tabs), categories (Student/Student), name search, and pagination.
// @Tags         Admin - Application
// @Security     ApiKeyAuth
// @Produce      json
// @Param        status    query  string  false  "Filter status: diproses|diterima|ditolak|aktif|selesai"
// @Param        category  query  string  false  "Category filter: Mahasiswa|Siswa"
// @Param        search    query  string  false  "Search by registrant's full name"
// @Param        page      query  int     false  "Page (default 1)"
// @Param        limit     query  int     false  "Number of items per page (default 10, maximum 100))"
// @Success      200  {object}  map[string]interface{} "PKL registration list"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      403  {object}  map[string]interface{} "Forbidden - not admin"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /admin/applications [get]
func (h *AdminApplicationController) GetApplications(c *fiber.Ctx) error {
	var input dto.AdminApplicationListRequest
	if err := c.QueryParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid query parameters")
	}

	res, err := h.srv.GetApplications(input)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pendaftaran")
	}

	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data pendaftaran", res)
}

// GetApplicationDetail godoc
// @Summary      View PKL Registration Details (Admin)
// @Description  Retrieve complete details of one PKL registration based on ID, including the applicant's personal data, academic information, document files, and unit/mentor data if assigned
// @Tags         Admin - Application
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path  int  true  "ID Pendaftaran PKL"
// @Success      200  {object}  map[string]interface{} "Registration details"
// @Failure      400  {object}  map[string]interface{} "Invalid ID"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      403  {object}  map[string]interface{} "Forbidden"
// @Failure      404  {object}  map[string]interface{} "Registration not found"
// @Router       /admin/applications/{id} [get]
func (h *AdminApplicationController) GetApplicationDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pendaftaran tidak valid")
	}

	res, err := h.srv.GetApplicationDetail(uint(id))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrApplicationNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail pendaftaran")
		}
	}

	return successResponse(c, fiber.StatusOK, "Berhasil mengambil detail pendaftaran", res)
}

// UpdateApplicationStatus godoc
// @Summary      Verify and Change PKL Registration Status (Admin)
// @Description  The "approved" status requires unit_id, mentor_id, and attendance_date. The "rejected" status requires admin_notes
// @Tags         Admin - Application
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                                true  "Registration ID"
// @Param        request  body  dto.UpdateApplicationStatusRequest  true  "Data update status"
// @Success      200  {object}  map[string]interface{} "Status updated successfully"
// @Failure      400  {object}  map[string]interface{} "Validation failed"
// @Failure      404  {object}  map[string]interface{} "Registration not found"
// @Failure      409  {object}  map[string]interface{} "Registration is not pending status"
// @Router       /admin/applications/{id}/status [patch]
func (h *AdminApplicationController) UpdateApplicationStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pendaftaran tidak valid")
	}

	var input dto.UpdateApplicationStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	actorIDClaim := c.Locals("user_id")
	actorID, _ := actorIDClaim.(uint)

	res, err := h.srv.UpdateApplicationStatus(uint(id), input, actorID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrApplicationNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, apperrors.ErrApplicationNotPending):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		case errors.Is(err, apperrors.ErrInvalidApplicationStatus),
			errors.Is(err, apperrors.ErrApprovedRequiresUnitMentor),
			errors.Is(err, apperrors.ErrRejectedRequiresNotes),
			errors.Is(err, apperrors.ErrUnitNotFound),
			errors.Is(err, apperrors.ErrMentorNotFound),
			errors.Is(err, apperrors.ErrMentorNotInUnit):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui status pendaftaran")
		}
	}

	return successResponse(c, fiber.StatusOK, "Status pendaftaran berhasil diperbarui", res)
}

// GetUnits godoc
// @Summary      Unit/Division List (for Accept modal dropdown)
// @Tags         Admin - Application
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/units [get]
func (h *AdminApplicationController) GetUnits(c *fiber.Ctx) error {
	units, err := h.srv.GetUnits()
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data unit")
	}
	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data unit", units)
}

// GetMentorsByUnit godoc
// @Summary      List of Mentors by Unit (for Accept modal dropdown)
// @Tags         Admin - Application
// @Security     ApiKeyAuth
// @Produce      json
// @Param        unit_id  path  int  true  "ID Unit"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /admin/units/{unit_id}/mentors [get]
func (h *AdminApplicationController) GetMentorsByUnit(c *fiber.Ctx) error {
	unitIDStr := c.Params("unit_id")
	unitID, err := strconv.ParseUint(unitIDStr, 10, 64)
	if err != nil || unitID == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "unit_id tidak valid")
	}

	mentors, err := h.srv.GetMentorsByUnit(uint(unitID))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrUnitNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pembimbing")
		}
	}

	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data pembimbing", mentors)
}

// GetDashboardSummary godoc
// @Summary      Dashboard Summary (Admin)
// @Description  Displays the number of applicants currently being processed by level (SMK/D3/D4/S1), registration trends by year, and a list of recent registrations
// @Tags         Admin - Dashboard
// @Security     ApiKeyAuth
// @Produce      json
// @Param        year  query  int  false  "Number of past years for the trend chart (default: 5)"
// @Success      200  {object}  map[string]interface{} "Data dashboard"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      403  {object}  map[string]interface{} "Forbidden"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /admin/dashboard/summary [get]
func (h *AdminApplicationController) GetDashboardSummary(c *fiber.Ctx) error {
	years, err := strconv.Atoi(c.Query("years"))
	if err != nil || years <= 0 {
		years = 5
	}

	res, err := h.srv.GetDashboardSummary(years)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data dashboard")
	}

	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data dashboard", res)
}