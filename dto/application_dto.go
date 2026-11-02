package dto

type CreateApplicationRequest struct {
	EducationLevel      string `json:"education_level"       validate:"required,oneof=SMK D3 D4 S1"`
	IdentityNumber      string `json:"identity_number"       validate:"required"`
	InstitutionName     string `json:"institution_name"      validate:"required"`
	Major               string `json:"major"                 validate:"required"`
	Semester            int    `json:"semester"              validate:"required,min=1"`
	StartDate           string `json:"start_date"            validate:"required"` // Format: YYYY-MM-DD
	EndDate             string `json:"end_date"              validate:"required"` // Format: YYYY-MM-DD
	IntegrityPactAgreed bool   `json:"integrity_pact_agreed" validate:"required"`
}

type DocumentResponse struct {
	ID           uint   `json:"id"`
	DocumentType string `json:"document_type"`
	FileURL      string `json:"file_url"`
	UploadedAt   string `json:"uploaded_at"`
}

type LetterResponse struct {
	ID           uint   `json:"id"`
	LetterNumber string `json:"letter_number"`
	LetterDate   string `json:"letter_date"`
	FileURL      string `json:"file_url"`
	GeneratedAt  string `json:"generated_at"`
}

type ApplicationListResponse struct {
	ID               uint   `json:"id"`
	RegistrationDate string `json:"registration_date"`
	EducationLevel   string `json:"education_level"`
	Period           string `json:"period"`
	Status           string `json:"status"`
}

type ApplicationDetailResponse struct {
	ID               uint   `json:"id"`
	RegistrationDate string `json:"registration_date"`
	Period           string `json:"period"`
	Status           string `json:"status"`
	AdminNotes       string `json:"admin_notes,omitempty"`
	AttendanceDate   string `json:"attendance_date,omitempty"`

	Name            string `json:"name"`
	EducationLevel  string `json:"education_level"`
	IdentityNumber  string `json:"identity_number"`
	InstitutionName string `json:"institution_name"`
	Major           string `json:"major"`
	SemesterOrClass string `json:"semester_or_class"`

	Documents []DocumentResponse `json:"documents,omitempty"`
	Letter    *LetterResponse    `json:"letter,omitempty"`
}

type CreateApplicationResponse struct {
	ID                  uint               `json:"id"`
	UserID              uint               `json:"user_id"`
	EducationLevel      string             `json:"education_level"`
	IdentityNumber      string             `json:"identity_number"`
	InstitutionName     string             `json:"institution_name"`
	Major               string             `json:"major"`
	SemesterOrClass     string             `json:"semester_or_class"`
	StartDate           string             `json:"start_date"`
	EndDate             string             `json:"end_date"`
	Status              string             `json:"status"`
	IntegrityPactAgreed bool               `json:"integrity_pact_agreed"`
	Documents           []DocumentResponse `json:"documents,omitempty"`
}