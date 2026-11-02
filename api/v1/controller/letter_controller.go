package controller

import (
	"errors"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/service"

	"github.com/gofiber/fiber/v2"
)

type LetterController struct {
	srv service.LetterService
}

func NewLetterController(srv service.LetterService) *LetterController {
	return &LetterController{srv}
}

// GenerateLetter godoc
// @Summary      Generate PKL Acceptance Letter
// @Description  Generating a PKL acceptance letter in PDF format based on approved registration data. The letter is uploaded to Cloudinary and its URL is saved to the database
// @Tags         Admin - Letter
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                         true  "PKL Registration ID"
// @Param        request  body  dto.GenerateLetterRequest   true  "Letter data"
// @Success      201  {object}  map[string]interface{} "Letter successfully created"
// @Failure      400  {object}  map[string]interface{} "Validation failed / wrong date format"
// @Failure      404  {object}  map[string]interface{} "Registration not found"
// @Failure      409  {object}  map[string]interface{} "The letter has been created / status is not approved"
// @Router       /admin/applications/{id}/letter [post]
func (h *LetterController) GenerateLetter(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pendaftaran tidak valid")
	}

	var input dto.GenerateLetterRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.LetterNumber == "" || input.LetterDate == "" ||
		input.RecipientName == "" || input.SignerName == "" ||
		input.SignerTitle == "" || input.SignerNik == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Semua field wajib diisi")
	}

	res, err := h.srv.GenerateLetter(uint(id), input)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrApplicationNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, apperrors.ErrLetterOnlyForApproved):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		case errors.Is(err, apperrors.ErrLetterAlreadyGenerated):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Gagal membuat surat penerimaan")
		}
	}

	return successResponse(c, fiber.StatusCreated, "Surat penerimaan berhasil dibuat", res)
}