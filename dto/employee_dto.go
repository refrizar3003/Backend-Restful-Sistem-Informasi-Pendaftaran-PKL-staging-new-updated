package dto

type CreateEmployeeRequest struct {
	Fullname        string `form:"fullname"          validate:"required"`
	Email           string `form:"email"             validate:"required,email"`
	PhoneNumber     string `form:"phone_number"      validate:"required"`
	Role            string `form:"role"              validate:"required,oneof=admin pegawai"`
	Password        string `form:"password"          validate:"required,min=8"`
	ConfirmPassword string `form:"confirm_password"  validate:"required,min=8"`
}

type UpdateEmployeeRequest struct {
	Fullname        string `form:"fullname"          validate:"required"`
	Email           string `form:"email"             validate:"required,email"`
	PhoneNumber     string `form:"phone_number"      validate:"required"`
	Role            string `form:"role"              validate:"required,oneof=admin pegawai"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

type EmployeeListItem struct {
	ID           uint   `json:"id"`
	Fullname     string `json:"fullname"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	Role         string `json:"role"`
	ProfilePhoto string `json:"profile_photo"`
}

type EmployeeDetailResponse struct {
	ID           uint   `json:"id"`
	Fullname     string `json:"fullname"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	Role         string `json:"role"`
	ProfilePhoto string `json:"profile_photo"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}