package dto

type AdminApplicationDetailResponse struct {
	ID               uint   `json:"id"`
	RegistrationDate string `json:"registration_date"`
	Category         string `json:"category"`
	Period           string `json:"period"`
	Status           string `json:"status"`
	AdminNotes       string `json:"admin_notes,omitempty"`
	AttendanceDate   string `json:"attendance_date,omitempty"`

	Applicant AdminApplicantInfo `json:"applicant"`

	Academic AdminAcademicInfo `json:"academic"`

	Documents []DocumentResponse `json:"documents,omitempty"`
	Letter    *LetterResponse    `json:"letter,omitempty"`

	Unit   *AdminUnitInfo   `json:"unit,omitempty"`
	Mentor *AdminMentorInfo `json:"mentor,omitempty"`
}

type AdminApplicantInfo struct {
	ID          uint   `json:"id"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type AdminAcademicInfo struct {
	EducationLevel  string `json:"education_level"`
	InstitutionName string `json:"institution_name"`
	Major           string `json:"major"`
	IdentityNumber  string `json:"identity_number"`
	SemesterOrClass string `json:"semester_or_class"`
}

type AdminUnitInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type AdminMentorInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
}