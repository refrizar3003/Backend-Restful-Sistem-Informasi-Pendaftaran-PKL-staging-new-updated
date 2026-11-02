package controller

import (
	"errors"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/service"
	"os"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	srv service.UserService
}

func NewUserController(srv service.UserService) *UserController {
	return &UserController{srv: srv}
}

func successResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	res := fiber.Map{
		"status":  "success",
		"message": message,
	}
	if data != nil {
		res["data"] = data
	}
	return c.Status(status).JSON(res)
}

func errorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// Register godoc
// @Summary      Registering a New Account
// @Description  Create a new prospective PKL participant account in the database system
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "User Registration Data"
// @Success      201      {object}  map[string]interface{} "User registered successfully"
// @Failure      400      {object}  map[string]interface{} "Invalid request body"
// @Failure      409      {object}  map[string]interface{} "Email already registered"
// @Failure      400      {object}  map[string]interface{} "Invalid request body / password mismatch
// @Router       /register [post]
func (h *UserController) Register(c *fiber.Ctx) error {
	var input dto.RegisterRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := h.srv.RegisterUser(input)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmailAlreadyVerified):
			return errorResponse(c, fiber.StatusConflict, err.Error())
		case errors.Is(err, apperrors.ErrEmailPendingVerify):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrPasswordMismatch):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}

	dateOfBirth := ""
	if user.DateOfBirth != nil {
		dateOfBirth = user.DateOfBirth.Format("2006-01-02")
	}

	response := dto.UserProfileResponse{
		ID:       	 user.ID,
		Fullname: 	 user.Fullname,
		Gender: 	 user.Gender,
		DateOfBirth: dateOfBirth,
		Email:    	 user.Email,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
		Province:    user.Province,
		Role:     	 user.Role,
	}

	return successResponse(c, fiber.StatusCreated, "User registered successfully", response)
}

// VerifyAccount godoc
// @Summary      User Account Verification
// @Description  Change the 'is_verified' status to true using the JWT token sent via registration email
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        token    query     string  true  "JWT Verification Token"
// @Success      200      {object}  map[string]interface{} "Account successfully verified"
// @Failure      400      {object}  map[string]interface{} "Token is invalid, expired, or not found"
// @Router       /api/v1/verify [get]
func (h *UserController) VerifyAccount(c *fiber.Ctx) error {
	frontendURL := os.Getenv("FRONTEND_URL")
    if frontendURL == "" {
        frontendURL = "http://localhost:3000"
    }
	
	token := c.Query("token")
	if token == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Verification token is missing")
	}

	if err := h.srv.VerifyAccount(token); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrAccountAlreadyVerified):
			return c.Redirect(frontendURL+"/login", fiber.StatusFound)
		default:
			return c.Redirect(frontendURL+"/verify-failed", fiber.StatusFound)
		}
	}

	return c.Redirect(frontendURL+"/login", fiber.StatusFound)
}

// Login godoc
// @Summary      Login to the System
// @Description  Perform email and password authentication to get JWT Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Login Credentials"
// @Success      200      {object}  map[string]interface{} "Login successful (data.token)"
// @Failure      401      {object}  map[string]interface{} "Invalid email or password"
// @Failure      403      {object}  map[string]interface{} "Account not verified"
// @Failure      429      {object}  map[string]interface{} "Too many requests"
// @Failure      503      {object}  map[string]interface{} "Cache service unavailable (OTP cannot be issued)"
// @Router       /login [post]
func (h *UserController) Login(c *fiber.Ctx) error {
	var input dto.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	res, err := h.srv.Login(input)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrAccountNotVerified):
			return errorResponse(c, fiber.StatusForbidden, err.Error())
		case errors.Is(err, apperrors.ErrInvalidCredentials):
			return errorResponse(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrCacheUnavailable):
			return errorResponse(c, fiber.StatusServiceUnavailable, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}

	if res.OTPRequired {
		return successResponse(c, fiber.StatusOK, "OTP has been sent to your email", res)
	}
	return successResponse(c, fiber.StatusOK, "Login successful", res)
}

// VerifyLoginOTP godoc
// @Summary      Verify Login OTP
// @Description  Completes the second step of login. Send the challenge_token from POST /login together with the 6-digit OTP received by email. Returns the JWT token on success.
// @Description  Set "remember_device" to true to also receive a "device_token". Send it as "device_token" in later POST /login requests to skip the OTP on this device until it expires.
// @Description  The OTP is valid for 5 minutes, single use, and allows at most 5 attempts.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyLoginOTPRequest  true  "Challenge token, OTP code, and optional remember_device flag"
// @Success      200      {object}  map[string]interface{} "Login successful (data.token, and data.device_token if remember_device is true)"
// @Failure      401      {object}  map[string]interface{} "Invalid or expired challenge token, or invalid/expired OTP"
// @Failure      429      {object}  map[string]interface{} "Too many incorrect OTP attempts, or too many requests"
// @Failure      503      {object}  map[string]interface{} "Cache service unavailable"
// @Router       /login/verify-otp [post]
func (h *UserController) VerifyLoginOTP(c *fiber.Ctx) error {
	var input dto.VerifyLoginOTPRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	res, err := h.srv.VerifyLoginOTP(input, c.Get("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidChallengeToken),
			errors.Is(err, apperrors.ErrInvalidOTP),
			errors.Is(err, apperrors.ErrOTPMismatch),
			errors.Is(err, apperrors.ErrUserNotFound):
			return errorResponse(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrTooManyOTPAttempts):
			return errorResponse(c, fiber.StatusTooManyRequests, err.Error())
		case errors.Is(err, apperrors.ErrCacheUnavailable):
			return errorResponse(c, fiber.StatusServiceUnavailable, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}
	return successResponse(c, fiber.StatusOK, "Login successful", res)
}

// ResendLoginOTP godoc
// @Summary      Resend Login OTP
// @Description  Sends a new login OTP to the user's email. Requires the challenge_token from POST /login (valid for 10 minutes). Limited to once every 60 seconds. The previous OTP stops working and the attempt counter is reset.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResendLoginOTPRequest  true  "Challenge token"
// @Success      200      {object}  map[string]interface{} "A new OTP has been sent to your email"
// @Failure      401      {object}  map[string]interface{} "Invalid or expired challenge token"
// @Failure      429      {object}  map[string]interface{} "Please wait 2 minute before requesting a new OTP, or too many requests"
// @Failure      503      {object}  map[string]interface{} "Cache service unavailable"
// @Router       /login/resend-otp [post]
func (h *UserController) ResendLoginOTP(c *fiber.Ctx) error {
	var input dto.ResendLoginOTPRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.srv.ResendLoginOTP(input); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidChallengeToken),
			errors.Is(err, apperrors.ErrUserNotFound):
			return errorResponse(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrResendOTPCooldown):
			return errorResponse(c, fiber.StatusTooManyRequests, err.Error())
		case errors.Is(err, apperrors.ErrCacheUnavailable):
			return errorResponse(c, fiber.StatusServiceUnavailable, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}
	return successResponse(c, fiber.StatusOK, "A new OTP has been sent to your email", nil)
}

// ForgotPassword godoc
// @Summary      Forgot Password
// @Description  Sends OTP code to user's email if the email is registered
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ForgotPasswordRequest  true  "Email to Reset Password"
// @Success      200      {object}  map[string]interface{} "If your email is registered, you will receive an OTP"
// @Failure      400      {object}  map[string]interface{} "Invalid request body"
// @Router       /forgot-password [post]
func (h *UserController) ForgotPassword(c *fiber.Ctx) error {
	var input dto.ForgotPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.srv.ForgotPassword(input); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
	}

	return successResponse(c, fiber.StatusOK, "If your email is registered, you will receive an OTP", nil)
}

// ResendOTP godoc
// @Summary      Resend OTP Reset Password
// @Description  Resends a new OTP to the user's email. Requires an active OTP session (must have called /forgot-password first). Limited to once every 60 seconds.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResendOTPRequest  true  "Email address"
// @Success      200      {object}  map[string]interface{} "A new OTP has been sent to your email"
// @Failure      400      {object}  map[string]interface{} "Invalid request body or no active OTP session"
// @Failure      429      {object}  map[string]interface{} "Please wait 60 seconds before requesting a new OTP"
// @Failure      500      {object}  map[string]interface{} "Internal server error"
// @Router       /resend-otp [post]
func (h *UserController) ResendOTP(c *fiber.Ctx) error {
	var input dto.ResendOTPRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}
 
	if err := h.srv.ResendOTP(input); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNoActiveOTPRequest):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrResendOTPCooldown):
			return errorResponse(c, fiber.StatusTooManyRequests, err.Error())
		case errors.Is(err, apperrors.ErrCacheUnavailable):
			return errorResponse(c, fiber.StatusServiceUnavailable, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}
 
	return successResponse(c, fiber.StatusOK, "A new OTP has been sent to your email", nil)
}

// ResendVerification godoc
// @Summary      Resend Account Verification Link
// @Description  Resends an email containing an account verification link. Called by the frontend when a user attempts to log in but receives a 403 (account not verified) response. Limited to once every 120 seconds.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResendVerificationRequest  true  "Email address"
// @Success      200      {object}  map[string]interface{} "Verification email sent"
// @Failure      400      {object}  map[string]interface{} "Invalid request body"
// @Failure      429      {object}  map[string]interface{} "Cooldown active, wait 2 minutes"
// @Router       /resend-verification [post]
func (h *UserController) ResendVerification(c *fiber.Ctx) error {
	var input dto.ResendVerificationRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.srv.ResendVerification(input); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrResendVerificationCooldown):
			return errorResponse(c, fiber.StatusTooManyRequests, err.Error())
		case errors.Is(err, apperrors.ErrCacheUnavailable):
			return errorResponse(c, fiber.StatusServiceUnavailable, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}

	return successResponse(c, fiber.StatusOK, "If your email is registered and unverified, a new verification link has been sent", nil)
}

// VerifyCode godoc
// @Summary      OTP Code Verification
// @Description  Validates the OTP code sent to your email. If correct, it will return a Reset Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyCodeRequest  true  "Email and OTP Code"
// @Success      200      {object}  map[string]interface{} "OTP verified successfully"
// @Failure      401      {object}  map[string]interface{} "Unauthorized / Invalid OTP"
// @Router       /verify-code [post]
func (h *UserController) VerifyCode(c *fiber.Ctx) error {
	var input dto.VerifyCodeRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	resetToken, err := h.srv.VerifyCode(input)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidOTP), errors.Is(err, apperrors.ErrOTPMismatch):
			return errorResponse(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrUserNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}

	return successResponse(c, fiber.StatusOK, "OTP verified successfully", fiber.Map{"reset_token": resetToken})
}

// ResetPassword godoc
// @Summary      Reset New Password
// @Description  Replace the old password with a new password using the Reset Token obtained from OTP verification
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResetPasswordRequest  true  "New Password Data"
// @Success      200      {object}  map[string]interface{} "Password has been reset successfully"
// @Failure      401      {object}  map[string]interface{} "Unauthorized / Invalid Token"
// @Router       /reset-password [post]
func (h *UserController) ResetPassword(c *fiber.Ctx) error {
	var input dto.ResetPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.srv.ResetPassword(input); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrPasswordMismatch):
			return errorResponse(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrInvalidResetToken), errors.Is(err, apperrors.ErrInvalidTokenScope):
			return errorResponse(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrUserNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "Internal server error, please try again later")
		}
	}

	return successResponse(c, fiber.StatusOK, "Password has been reset successfully", nil)
}

// GetProfile godoc
// @Summary      Retrieve user profile data
// @Description  Retrieving the profile information details of the currently logged in student based on the JWT token
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserProfileResponse "Successfully retrieved profile"
// @Failure      401  {object}  map[string]string "Unauthorized - Invalid token"
// @Failure      404  {object}  map[string]string "User not found"
// @Router       /profile [get]
func (h *UserController) GetProfile(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)
 
	userProfile, err := h.srv.GetProfile(userID)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "User not found")
	}
 
	return successResponse(c, fiber.StatusOK, "Successfully retrieved profile", userProfile)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update user profile information and/or upload a new profile photo
// @Tags         Profile
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        fullname      formData  string  true   "Full Name"
// @Param        gender        formData  string  true   "Gender (Male/Female)"
// @Param        date_of_birth formData  string  true   "Date of Birth (YYYY-MM-DD)"
// @Param        email         formData  string  true   "Email Address"
// @Param        phone_number  formData  string  true   "Phone Number"
// @Param        province      formData  string  true   "Province"
// @Param        address       formData  string  true   "Full Address"
// @Param        profile_photo formData  file    false  "Profile Photo File (JPG/PNG)"
// @Success      200  {object}  map[string]string "Profile updated successfully"
// @Failure      400  {object}  map[string]string "Bad Request - Invalid data format"
// @Failure      401  {object}  map[string]string "Unauthorized - Invalid token"
// @Failure      500  {object}  map[string]string "Internal Server Error"
// @Router       /profile [put]
func (h *UserController) UpdateProfile(c *fiber.Ctx) error {
	userIDClaim := c.Locals("user_id")
	if userIDClaim == nil {
		return errorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := userIDClaim.(uint)

	var input dto.UpdateProfileRequest
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid data format")
	}

	file, _ := c.FormFile("profile_photo")

	if err := h.srv.UpdateProfile(userID, input, file); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrUserNotFound):
			return errorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return errorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return successResponse(c, fiber.StatusOK, "Profile updated successfully", nil)
}

