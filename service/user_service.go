package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/infrastructure"
	"internship-telkom/model"
	"internship-telkom/repository"
	"internship-telkom/utils"
	"log"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/redis/go-redis/v9"
)

const (
	loginOTPTTL         = 5 * time.Minute
	loginOTPCooldown    = 120 * time.Second
	maxLoginOTPAttempts = 5
)

func loginOTPKey(id uint) string         { return fmt.Sprintf("login_otp:%d", id) }
func loginOTPAttemptsKey(id uint) string { return fmt.Sprintf("login_otp:attempts:%d", id) }
func loginOTPCooldownKey(id uint) string { return fmt.Sprintf("login_otp:cooldown:%d", id) }

// UserService adalah antarmuka untuk logika bisnis User
type UserService interface {
	RegisterUser(input dto.RegisterRequest) (model.User, error)
	Login(input dto.LoginRequest) (dto.LoginResponse, error)
	VerifyLoginOTP(input dto.VerifyLoginOTPRequest, userAgent string) (dto.VerifyLoginOTPResponse, error)
	ResendLoginOTP(input dto.ResendLoginOTPRequest) error
	ForgotPassword(input dto.ForgotPasswordRequest) error
	ResendOTP(input dto.ResendOTPRequest) error
	ResendVerification(input dto.ResendVerificationRequest) error
	VerifyCode(input dto.VerifyCodeRequest) (string, error)
	ResetPassword(input dto.ResetPasswordRequest) error
	UpdateProfile(userID uint, input dto.UpdateProfileRequest, photo *multipart.FileHeader) error
	GetProfile(userID uint) (dto.UserProfileResponse, error)
	VerifyAccount(token string) error
}

type userService struct {
	repo repository.UserRepository
	redis *redis.Client
	cloudinary *cloudinary.Cloudinary
	jwtSrv JwtService
	mailer     infrastructure.Mailer
	loginOTPEnabled bool
	deviceRepo                 repository.TrustedDeviceRepository
	trustedDeviceDaysStaff     time.Duration
	trustedDeviceDaysApplicant time.Duration
}

// NewUserService membuat instance baru dari userService
func NewUserService(repo repository.UserRepository, rdb *redis.Client, cld *cloudinary.Cloudinary, jwtSrv JwtService, 
    mailer infrastructure.Mailer, deviceRepo repository.TrustedDeviceRepository) UserService {
	return &userService{
		repo: repo,
		redis:      rdb,
		cloudinary: cld,
		jwtSrv: jwtSrv,
		mailer:     mailer,
		loginOTPEnabled: os.Getenv("LOGIN_OTP_ENABLED") == "true",
		deviceRepo:                 deviceRepo,
		trustedDeviceDaysStaff:     getEnvDuration("TRUSTED_DEVICE_DAYS_STAFF", 7),
		trustedDeviceDaysApplicant: getEnvDuration("TRUSTED_DEVICE_DAYS_APPLICANT", 14),
	}
}

func (s *userService) RegisterUser(input dto.RegisterRequest) (model.User, error) {
	if input.Password != input.ConfirmPassword {
		return model.User{}, apperrors.ErrPasswordMismatch
	}
	
	existingUser, err := s.repo.FindByEmail(input.Email)
	if err == nil {
		// Email ditemukan di database
		if existingUser.IsVerified {
			return model.User{}, apperrors.ErrEmailAlreadyVerified
		}
 
		// Akun belum terverifikasi: hapus jika sudah lewat 24 jam, tolak jika belum
		if time.Since(existingUser.CreatedAt) > 24*time.Hour {
			if err := s.repo.DeleteByID(existingUser.ID); err != nil {
				return model.User{}, fmt.Errorf("failed to cleanup expired account: %w", err)
			}
		} else {
			return model.User{}, apperrors.ErrEmailPendingVerify
		}
	}
 
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to hash password: %w", err)
	}
 
	dateOfBirth, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		return model.User{}, fmt.Errorf("invalid date of birth format, must be YYYY-MM-DD")
	}
 
	user := model.User{
		Fullname:    input.Fullname,
		Gender:      input.Gender,
		DateOfBirth: &dateOfBirth,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Address:     input.Address,
		Province:    input.Province,
		Password:    passwordHash,
		Role:        "applicant",
		IsActive:    true,
	}
 
	newUser, err := s.repo.Save(user)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to save user: %w", err)
	}
 
	// Generate JWT Verification Token
	token, err := s.jwtSrv.GenerateVerificationToken(newUser.Email)
	if err != nil {
		return newUser, apperrors.ErrFailedGenerateToken
	}
 
	// Kirim Email Verifikasi via Mailer Infrastructure
	if err := s.mailer.SendVerificationEmail(newUser.Email, token); err != nil {
		log.Printf("Warning: failed to send verification email to %s: %v", newUser.Email, err)
	}
 
	return newUser, nil
}

func (s *userService) VerifyAccount(token string) error {
	// Validasi token menggunakan jwtService
	claims, err := s.jwtSrv.ValidateToken(token)
	if err != nil {
		return apperrors.ErrInvalidVerifyToken
	}
 
	// Pastikan scope token-nya benar
	if claims["scope"] != "verification" {
		return apperrors.ErrInvalidTokenScope
	}
 
	// Ambil data email dari dalam claims JWT
	email, ok := claims["email"].(string)
	if !ok {
		return apperrors.ErrInvalidTokenScope
	}
 
	// Cari pengguna di database berdasarkan email
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return apperrors.ErrUserNotFound
	}
 
	// Cek jika pengguna sudah pernah melakukan verifikasi
	if user.IsVerified {
		return apperrors.ErrAccountAlreadyVerified
	}
 
	// Ubah status menjadi terverifikasi dan hapus token agar tidak bisa dipakai ulang
	user.IsVerified = true
	user.VerificationToken = nil
 
	if err := s.repo.UpdateUser(&user); err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}
	return nil
}

// UPDATE: Parameter input menggunakan dto.LoginRequest
func (s *userService) Login(input dto.LoginRequest) (dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return dto.LoginResponse{}, apperrors.ErrInvalidCredentials
	}
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return dto.LoginResponse{}, apperrors.ErrInvalidCredentials
	}
	if !user.IsVerified {
		return dto.LoginResponse{}, apperrors.ErrAccountNotVerified
	}

	if s.loginOTPEnabled && !s.isTrustedDevice(user.ID, input.DeviceToken) {
		if s.redis == nil {
			return dto.LoginResponse{}, apperrors.ErrCacheUnavailable
		}
		challenge, err := s.jwtSrv.GenerateLoginChallengeToken(user.ID)
		if err != nil {
			return dto.LoginResponse{}, apperrors.ErrFailedGenerateToken
		}
		if err := s.issueLoginOTP(context.Background(), user); err != nil {
			return dto.LoginResponse{}, err
		}
		return dto.LoginResponse{OTPRequired: true, ChallengeToken: challenge}, nil
	}

	token, err := s.jwtSrv.GenerateLoginToken(user.ID, user.Role)
	if err != nil {
		return dto.LoginResponse{}, apperrors.ErrFailedGenerateToken
	}
	return dto.LoginResponse{Token: token}, nil
}

func (s *userService) isTrustedDevice(userID uint, deviceToken string) bool {
	if deviceToken == "" {
		return false
	}
	now := time.Now()
	device, err := s.deviceRepo.FindValid(userID, utils.HashDeviceToken(deviceToken), now)
	if err != nil {
		return false // gagal tutup: semua error berarti "minta OTP"
	}
	_ = s.deviceRepo.TouchLastUsed(device.ID, now)
	return true
}

func (s *userService) rememberDevice(user model.User, userAgent string) (string, error) {
	token, err := utils.GenerateDeviceToken()
	if err != nil {
		return "", err
	}

	days := s.trustedDeviceDaysApplicant
	if user.Role == model.RoleAdmin || user.Role == model.RolePegawai {
		days = s.trustedDeviceDaysStaff
	}

	if r := []rune(userAgent); len(r) > 255 {
		userAgent = string(r[:255])
	}

	_ = s.deviceRepo.DeleteExpired(time.Now())
	err = s.deviceRepo.Create(model.TrustedDevice{
		UserID:    user.ID,
		TokenHash: utils.HashDeviceToken(token),
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(days * 24 * time.Hour),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) issueLoginOTP(ctx context.Context, user model.User) error {
	otp := utils.GenerateOTP()

	pipe := s.redis.Pipeline()
	pipe.Set(ctx, loginOTPKey(user.ID), otp, loginOTPTTL)
	pipe.Del(ctx, loginOTPAttemptsKey(user.ID))
	pipe.Set(ctx, loginOTPCooldownKey(user.ID), "1", loginOTPCooldown)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store login OTP: %w", err)
	}

	if err := s.mailer.SendLoginOTPEmail(user.Email, otp); err != nil {
		s.redis.Del(ctx, loginOTPKey(user.ID), loginOTPCooldownKey(user.ID))
		return fmt.Errorf("failed to send login OTP email: %w", err)
	}
	return nil
}

func (s *userService) parseChallengeToken(token string) (uint, error) {
	claims, err := s.jwtSrv.ValidateToken(token)
	if err != nil {
		return 0, apperrors.ErrInvalidChallengeToken
	}
	if scope, ok := claims["scope"].(string); !ok || scope != "login_otp" {
		return 0, apperrors.ErrInvalidChallengeToken
	}
	idFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, apperrors.ErrInvalidChallengeToken
	}
	return uint(idFloat), nil
}

func (s *userService) VerifyLoginOTP(input dto.VerifyLoginOTPRequest, userAgent string) (dto.VerifyLoginOTPResponse, error) {
	userID, err := s.parseChallengeToken(input.ChallengeToken)
	if err != nil {
		return dto.VerifyLoginOTPResponse{}, err
	}
	if s.redis == nil {
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrCacheUnavailable
	}

	ctx := context.Background()
	otpKey := loginOTPKey(userID)
	attemptsKey := loginOTPAttemptsKey(userID)

	attempts, err := s.redis.Incr(ctx, attemptsKey).Result()
	if err != nil {
		return dto.VerifyLoginOTPResponse{}, fmt.Errorf("failed to track OTP attempts: %w", err)
	}
	if attempts == 1 {
		s.redis.Expire(ctx, attemptsKey, loginOTPTTL)
	}
	if attempts > maxLoginOTPAttempts {
		s.redis.Del(ctx, otpKey)
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrTooManyOTPAttempts
	}
	
	stored, err := s.redis.Get(ctx, otpKey).Result()
	if err == redis.Nil {
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrInvalidOTP
	} else if err != nil {
		return dto.VerifyLoginOTPResponse{}, fmt.Errorf("failed to read login OTP: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(input.OTP)) != 1 {
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrOTPMismatch
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrUserNotFound
	}

	s.redis.Del(ctx, otpKey, attemptsKey, loginOTPCooldownKey(userID))

	token, err := s.jwtSrv.GenerateLoginToken(user.ID, user.Role)
	if err != nil {
		return dto.VerifyLoginOTPResponse{}, apperrors.ErrFailedGenerateToken
	}
	
	res := dto.VerifyLoginOTPResponse{Token: token}
	if input.RememberDevice {
		deviceToken, err := s.rememberDevice(user, userAgent)
		if err != nil {
			log.Printf("Warning: failed to remember device for user %d: %v", user.ID, err)
		} else {
			res.DeviceToken = deviceToken
		}
	}
	return res, nil
}

func (s *userService) ResendLoginOTP(input dto.ResendLoginOTPRequest) error {
	userID, err := s.parseChallengeToken(input.ChallengeToken)
	if err != nil {
		return err
	}
	if s.redis == nil {
		return apperrors.ErrCacheUnavailable
	}

	ctx := context.Background()
	exists, err := s.redis.Exists(ctx, loginOTPCooldownKey(userID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check cooldown: %w", err)
	}
	if exists > 0 {
		return apperrors.ErrResendOTPCooldown
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}
	return s.issueLoginOTP(ctx, user)
}

func (s *userService) ForgotPassword(input dto.ForgotPasswordRequest) error {
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return nil
	}

	if s.redis == nil {
		log.Println("Error: Redis client is nil. Make sure Redis is running and connected.")
        return apperrors.ErrCacheUnavailable
	}

	otp := utils.GenerateOTP()
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", user.Email)

	if err := s.redis.Set(ctx, key, otp, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to save OTP to cache: %w", err)
	}

	if err := s.mailer.SendResetPasswordEmail(user.Email, otp); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (s *userService) ResendOTP(input dto.ResendOTPRequest) error {
	if s.redis == nil {
		log.Println("Error: Redis client is nil.")
		return apperrors.ErrCacheUnavailable
	}
 
	ctx := context.Background()
	otpKey := fmt.Sprintf("otp:%s", input.Email)
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", input.Email)
 
	exists, err := s.redis.Exists(ctx, otpKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check OTP existence: %w", err)
	}
	if exists == 0 {
		return apperrors.ErrNoActiveOTPRequest
	}
 
	cooldownExists, err := s.redis.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check cooldown: %w", err)
	}
	if cooldownExists > 0 {
		return apperrors.ErrResendOTPCooldown
	}
 
	newOTP := utils.GenerateOTP()
 
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, otpKey, newOTP, 10*time.Minute)
	pipe.Set(ctx, cooldownKey, "1", 60*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update OTP in cache: %w", err)
	}
 
	if err := s.mailer.SendResetPasswordEmail(input.Email, newOTP); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
 
	return nil
}

func (s *userService) ResendVerification(input dto.ResendVerificationRequest) error {
	if s.redis == nil {
		return apperrors.ErrCacheUnavailable
	}

	ctx := context.Background()
	cooldownKey := fmt.Sprintf("verify:cooldown:%s", input.Email)

	cooldownExists, err := s.redis.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check cooldown: %w", err)
	}

	if cooldownExists > 0 {
		return apperrors.ErrResendVerificationCooldown
	}

	user, err := s.repo.FindByEmail(input.Email)
	if err != nil || user.IsVerified {
		s.redis.Set(ctx, cooldownKey, "1", 120*time.Second)
		return nil
	}

	token, err := s.jwtSrv.GenerateVerificationToken(user.Email)
	if err != nil {
		return apperrors.ErrFailedGenerateToken
	}

	if err := s.redis.Set(ctx, cooldownKey, "1", 120*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set cooldown: %w", err)
	}

	if err := s.mailer.SendVerificationEmail(user.Email, token); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *userService) VerifyCode(input dto.VerifyCodeRequest) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", input.Email)
	
	// 1. Ambil OTP dari Redis
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", apperrors.ErrInvalidOTP
	} else if err != nil {
		return "", fmt.Errorf("failed to verify OTP: %w", err)
	}
	// 2. Bandingkan OTP
	if val != input.OTP {
		return "", apperrors.ErrOTPMismatch
	}
	// 3. Ambil data user dari DB
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return "", apperrors.ErrUserNotFound
	}
	// 4. Hapus OTP dari Redis
	s.redis.Del(ctx, key, fmt.Sprintf("otp:cooldown:%s", input.Email))

	// 5. Buat token reset
	resetToken, err := s.jwtSrv.GenerateResetToken(user.ID)
	if err != nil {
		return "", apperrors.ErrFailedGenerateToken
	}
	return resetToken, nil
}

// ResetPassword adalah logika untuk verifikasi OTP dan ganti password
func (s *userService) ResetPassword(input dto.ResetPasswordRequest) error {
	if input.Password != input.ConfirmPassword {
		return apperrors.ErrPasswordMismatch
	}

	claims, err := s.jwtSrv.ValidateToken(input.ResetToken)
	if err != nil {
		return apperrors.ErrInvalidResetToken
	}

	scope, ok := claims["scope"].(string)
	if !ok || scope != "reset_password" {
		return apperrors.ErrInvalidTokenScope
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return apperrors.ErrInvalidTokenScope
	}
	userID := uint(userIDFloat)

	if _, err := s.repo.FindByID(userID); err != nil {
		return apperrors.ErrUserNotFound
	}

	if err := s.deviceRepo.DeleteByUserID(userID); err != nil {
		return fmt.Errorf("failed to revoke trusted devices: %w", err)
	}
 
	newPasswordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
 
	if err := s.repo.UpdatePassword(userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (s *userService) UpdateProfile(userID uint, input dto.UpdateProfileRequest, photo *multipart.FileHeader) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	dateOfBirth, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		return fmt.Errorf("invalid date of birth format, must be YYYY-MM-DD")
	}

	// Update data teks
	user.Fullname = input.Fullname
    user.Gender = input.Gender
	user.DateOfBirth = &dateOfBirth
	user.Email = input.Email
    user.PhoneNumber = input.PhoneNumber
	user.Province = input.Province
    user.Address = input.Address

	if photo != nil {
		file, err := photo.Open()
		if err != nil {
			return fmt.Errorf("failed to open photo file: %w", err)
		}
		defer file.Close()
	
		uploadRes, err := s.cloudinary.Upload.Upload(context.Background(), file, uploader.UploadParams{
			Folder: "profile_photos",
		})
		if err != nil {
			return fmt.Errorf("failed to upload photo: %w", err)
		}
		user.ProfilePhoto = uploadRes.SecureURL
	}

	return s.repo.UpdateUser(&user)
}

func (s *userService) GetProfile(userID uint) (dto.UserProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return dto.UserProfileResponse{}, apperrors.ErrUserNotFound
	}

	dateOfBirth := ""
	if user.DateOfBirth != nil {
		dateOfBirth = user.DateOfBirth.Format("2006-01-02")
	}

    response := dto.UserProfileResponse{
		ID:           user.ID,
        Email:        user.Email,
        Fullname:     user.Fullname,
        Gender:       user.Gender,
        DateOfBirth:  dateOfBirth,
        PhoneNumber:  user.PhoneNumber,
        Address:      user.Address,
        Province:     user.Province,
        ProfilePhoto: user.ProfilePhoto,
		Role: user.Role,
	}
	return response, nil
}