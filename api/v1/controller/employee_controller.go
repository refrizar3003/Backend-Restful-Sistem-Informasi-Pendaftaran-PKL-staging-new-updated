package controller

import (
	"errors"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/service"

	"github.com/gofiber/fiber/v2"
)

type EmployeeController struct {
	srv service.EmployeeService
}

func NewEmployeeController(srv service.EmployeeService) *EmployeeController {
	return &EmployeeController{srv}
}

// CreateEmployee godoc
// @Summary      Add Employee Account (Admin)
// @Description  Administrator creates new staff account (admin/employee)
// @Tags         Admin - Employee
// @Security     ApiKeyAuth
// @Accept       mpfd
// @Produce      json
// @Param        fullname          formData  string  true   "Full name"
// @Param        email             formData  string  true   "Email"
// @Param        phone_number      formData  string  true   "Phone number"
// @Param        role              formData  string  true   "Access Rights (admin/employee)"
// @Param        password          formData  string  true   "Password"
// @Param        confirm_password  formData  string  true   "Confirm Password"
// @Param        profile_photo     formData  file    false  "Profile picture"
// @Success      201  {object}  map[string]interface{} "Employee account created successfully"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      409  {object}  map[string]interface{} "Email already in use"
// @Router       /admin/employees [post]
func (h *EmployeeController) CreateEmployee(c *fiber.Ctx) error {
	var input dto.CreateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validateStruct(input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	photo, _ := c.FormFile("profile_photo")

	res, err := h.srv.CreateEmployee(input, photo)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmployeeEmailTaken):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		case errors.Is(err, apperrors.ErrInvalidRole), errors.Is(err, apperrors.ErrPasswordConfirmMismatch):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal membuat akun pegawai")
		}
	}

	return successResponse(c, fiber.StatusCreated, "Akun pegawai berhasil dibuat", res)
}

// GetEmployees godoc
// @Summary      Employee Account List (Admin)
// @Description  Displays all staff accounts (admin & employee), excluding PKL applicant accounts
// @Tags         Admin - Employee
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{} "Employee list"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /admin/employees [get]
func (h *EmployeeController) GetEmployees(c *fiber.Ctx) error {
	res, err := h.srv.GetEmployees()
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pegawai")
	}
	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data pegawai", res)
}

// GetEmployeeDetail godoc
// @Summary      Employee Account Details (Admin)
// @Tags         Admin - Employee
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  map[string]interface{} "Employee details"
// @Failure      400  {object}  map[string]interface{} "Invalid ID"
// @Failure      404  {object}  map[string]interface{} "Employee not found"
// @Router       /admin/employees/{id} [get]
func (h *EmployeeController) GetEmployeeDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	res, err := h.srv.GetEmployeeDetail(uint(id))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmployeeNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail pegawai")
		}
	}
	return successResponse(c, fiber.StatusOK, "Berhasil mengambil detail pegawai", res)
}

// UpdateEmployee godoc
// @Summary      Edit Employee Account (Admin)
// @Description  Password & password confirmation are optional, leave them blank if you don't want to change the password
// @Tags         Admin - Employee
// @Security     ApiKeyAuth
// @Accept       mpfd
// @Produce      json
// @Param        id                path      int     true   "Employee ID"
// @Param        fullname          formData  string  true   "Full name"
// @Param        email             formData  string  true   "Email"
// @Param        phone_number      formData  string  true   "Phone number"
// @Param        role              formData  string  true   "Access Rights (admin/employee)"
// @Param        password          formData  string  false  "New Password (optional)"
// @Param        confirm_password  formData  string  false  "Confirm New Password (optional)"
// @Param        profile_photo     formData  file    false  "Profile picture"
// @Success      200  {object}  map[string]interface{} "Employee account successfully updated"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      404  {object}  map[string]interface{} "Employee not found"
// @Failure      409  {object}  map[string]interface{} "Email already in use"
// @Router       /admin/employees/{id} [put]
func (h *EmployeeController) UpdateEmployee(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	var input dto.UpdateEmployeeRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validateStruct(input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	photo, _ := c.FormFile("profile_photo")

	res, err := h.srv.UpdateEmployee(uint(id), input, photo)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmployeeNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, apperrors.ErrEmployeeEmailTaken):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		case errors.Is(err, apperrors.ErrInvalidRole), errors.Is(err, apperrors.ErrPasswordConfirmMismatch):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui akun pegawai")
		}
	}

	return successResponse(c, fiber.StatusOK, "Akun pegawai berhasil diperbarui", res)
}

// DeleteEmployee godoc
// @Summary      Delete Employee Account (Admin)
// @Description  Administrators cannot delete their own accounts that are being used to log in
// @Tags         Admin - Employee
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  map[string]interface{} "Employee account successfully deleted"
// @Failure      400  {object}  map[string]interface{} "Invalid ID / can't delete own account"
// @Failure      404  {object}  map[string]interface{} "Employee not found"
// @Router       /admin/employees/{id} [delete]
func (h *EmployeeController) DeleteEmployee(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	requesterIDClaim := c.Locals("user_id")
	if requesterIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	requesterID := requesterIDClaim.(uint)

	if err := h.srv.DeleteEmployee(uint(id), requesterID); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmployeeNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, apperrors.ErrCannotDeleteSelf):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus akun pegawai")
		}
	}

	return successResponse(c, fiber.StatusOK, "Akun pegawai berhasil dihapus", nil)
}