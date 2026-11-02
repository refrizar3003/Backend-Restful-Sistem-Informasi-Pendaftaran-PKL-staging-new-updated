package controller

import (
	"errors"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/service"
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ApplicationController struct {
	srv service.ApplicationService
}

func NewApplicationController(srv service.ApplicationService) *ApplicationController {
	return &ApplicationController{srv}
}

// Apply godoc
// @Summary      Submit PKL Registration Form
// @Description  Prospective participants submit academic data for internship registration
// @Tags         Application
// @Security     ApiKeyAuth
// @Accept       mpfd
// @Produce      json
// @Param        education_level      formData  string  true   "Education Level (SMK/D3/D4/S1))"
// @Param        identity_number      formData  string  true   "Identity Number (NIM/NISN)"
// @Param        institution_name     formData  string  true   "Institution Name"
// @Param        major                formData  string  true   "Major or Study Program"
// @Param        semester             formData  int     true   "Semester or Class (number only)"
// @Param        start_date           formData  string  true   "Start Date (YYYY-MM-DD)"
// @Param        end_date             formData  string  true   "End Date (YYYY-MM-DD)"
// @Param        integrity_pact_agreed formData bool   true   "Integrity Pact Agreement"
// @Param        recommendation_letter formData file   false  "Recommendation Letter"
// @Param        transcript           formData  file    false  "Academic Transcript"
// @Param        cv                   formData  file    false  "Curriculum Vitae"
// @Param        proposal             formData  file    false  "Internship Proposal"
// @Success      201  {object}  map[string]interface{} "Registration submitted successfully"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      409  {object}  map[string]interface{} "Active application already exists"
// @Router       /applications [post]
func (h *ApplicationController) Apply(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)
	semesterStr := c.FormValue("semester")
	semester, err := strconv.Atoi(semesterStr)
	if err != nil || semester < 1 {
		return errorResponse(c, fiber.StatusBadRequest, "semester harus berupa angka positif")
	}

	agreed, err := strconv.ParseBool(c.FormValue("integrity_pact_agreed"))
	input := dto.CreateApplicationRequest{
		EducationLevel:      c.FormValue("education_level"),
		IdentityNumber:      c.FormValue("identity_number"),
		InstitutionName:     c.FormValue("institution_name"),
		Major:               c.FormValue("major"),
		Semester:     		 semester,
		StartDate:           c.FormValue("start_date"),
		EndDate:             c.FormValue("end_date"),
		IntegrityPactAgreed: agreed,
	}

	files := make(map[string]*multipart.FileHeader)
	docTypes := []string{"recommendation_letter", "transcript", "cv", "proposal"}
	for _, docType := range docTypes {
		file, _ := c.FormFile(docType)
		if file != nil {
			files[docType] = file
		}
	}

	result, err := h.srv.ApplyInternship(userID, input, files)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrApplicationStillActive):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		default:
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		}
	}

	return successResponse(c, fiber.StatusCreated, "Pendaftaran berhasil dikirim", result)
}

// GetMyApplications godoc
// @Summary      View the PKL Application List
// @Description  Retrieve all PKL registration history belonging to the currently logged in user
// @Tags         Application
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{} "List of applications"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /applications [get]
func (h *ApplicationController) GetMyApplications(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	res, err := h.srv.GetMyApplications(userID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pendaftaran")
	}
	
	return successResponse(c, fiber.StatusOK, "Berhasil mengambil data pendaftaran", res)
}

// GetApplicationDetail godoc
// @Summary      View PKL Application Details
// @Description  Retrieve detailed data from a PKL registration form based on submission ID
// @Tags         Application
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path      int  true  "PKL Registration ID"
// @Success      200  {object}  map[string]interface{} "Application detail"
// @Failure      400  {object}  map[string]interface{} "Invalid ID format"
// @Failure      401  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Application not found"
// @Router       /applications/{id} [get]
func (h *ApplicationController) GetApplicationDetail(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	appID, err := c.ParamsInt("id")
	if err != nil || appID <= 0 {
		return errorResponse(c, fiber.StatusBadRequest, "ID pendaftaran tidak valid")
	}

	res, err := h.srv.GetApplicationDetail(uint(appID), userID)
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