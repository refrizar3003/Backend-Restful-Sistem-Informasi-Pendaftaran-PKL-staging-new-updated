package dto

type RegisterRequest struct {
	Fullname        string `json:"fullname"     validate:"required"`
	Gender          string `json:"gender"       validate:"required,oneof=male female"`
	DateOfBirth     string `json:"date_of_birth" validate:"required"`
	Email           string `json:"email"        validate:"required,email"`
	PhoneNumber     string `json:"phone_number" validate:"required"`
	Address         string `json:"address"      validate:"required"`
	Province        string `json:"province"     validate:"required"`
	Password        string `json:"password"     validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required"`
	DeviceToken string `json:"device_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type ResetPasswordRequest struct {
	ResetToken      string `json:"reset_token" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

type UpdateProfileRequest struct {
	Fullname    string `form:"fullname" validate:"required"`
	Gender      string `form:"gender" validate:"required,oneof=male female"`
	DateOfBirth string `form:"date_of_birth" validate:"required"`
	Email       string `form:"email" validate:"required,email"`
	PhoneNumber string `form:"phone_number" validate:"required"`
	Province    string `form:"province" validate:"required"`
	Address     string `form:"address" validate:"required"`
}

type UserProfileResponse struct {
	ID           uint   `json:"id"`
	Email        string `json:"email"`
	Fullname     string `json:"fullname"`
	Gender       string `json:"gender"`
	DateOfBirth  string `json:"date_of_birth"`
	PhoneNumber  string `json:"phone_number"`
	Address      string `json:"address"`
	Province     string `json:"province"`
	ProfilePhoto string `json:"profile_photo"`
	Role         string `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type LoginResponse struct {
	OTPRequired    bool   `json:"otp_required"`
	Token          string `json:"token,omitempty"`
	ChallengeToken string `json:"challenge_token,omitempty"`
}

type VerifyLoginOTPRequest struct {
	ChallengeToken string `json:"challenge_token" validate:"required"`
	OTP            string `json:"otp"             validate:"required,len=6"`
	RememberDevice bool   `json:"remember_device"`
}

type ResendLoginOTPRequest struct {
	ChallengeToken string `json:"challenge_token" validate:"required"`
}

type VerifyLoginOTPResponse struct {
	Token       string `json:"token"`
	DeviceToken string `json:"device_token,omitempty"`
}
