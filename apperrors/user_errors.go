package apperrors

import "errors"

// User errors
var (
	ErrEmailAlreadyVerified  = errors.New("email is already registered and verified")
	ErrEmailPendingVerify    = errors.New("please verify your email")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrUserNotFound          = errors.New("user not found")
	ErrFailedGenerateToken   = errors.New("failed to create authentication token")
	ErrCacheUnavailable      = errors.New("internal server error: cache service unavailable")
	ErrInvalidOTP            = errors.New("OTP expired or not requested")
	ErrOTPMismatch           = errors.New("invalid OTP")
	ErrInvalidResetToken     = errors.New("invalid or expired token")
	ErrPasswordMismatch      = errors.New("password and confirm password do not match")
	ErrInvalidTokenScope     = errors.New("invalid token scope")
	ErrAccountAlreadyVerified = errors.New("this account is already verified")
	ErrInvalidVerifyToken    = errors.New("the verification link is invalid or has expired")
	ErrAccountNotVerified = errors.New("your account is not verified yet, please check your email")
	ErrResendVerificationCooldown = errors.New("please wait 2 minute before requesting a new verification email")
	ErrNoActiveOTPRequest = errors.New("no active OTP request found, please use forgot-password first")
	ErrResendOTPCooldown  = errors.New("please wait 2 minute before requesting a new OTP")
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	ErrInvalidChallengeToken = errors.New("login session is invalid or has expired, please log in again")
	ErrTooManyOTPAttempts    = errors.New("too many incorrect OTP attempts, please log in again")
)

var (
	ErrApplicationStillActive = errors.New("anda memiliki pendaftaran yang sedang diproses")
	ErrApplicationNotFound    = errors.New("application not found")

	ErrInvalidApplicationStatus  = errors.New("status tidak valid, gunakan: approved atau rejected")
	ErrApplicationNotPending     = errors.New("hanya pendaftaran dengan status 'pending' yang dapat diverifikasi")

	ErrApprovedRequiresUnitMentor = errors.New("unit_id, mentor_id, dan attendance_date wajib diisi saat menyetujui pendaftaran")
	ErrRejectedRequiresNotes     = errors.New("alasan penolakan (admin_notes) wajib diisi saat menolak pendaftaran")
	ErrUnitNotFound              = errors.New("unit tidak ditemukan")
	ErrMentorNotFound            = errors.New("mentor tidak ditemukan")
	ErrMentorNotInUnit           = errors.New("mentor tidak terdaftar di unit yang dipilih")

	ErrLetterOnlyForApproved  = errors.New("surat penerimaan hanya dapat dibuat untuk pendaftaran yang telah diterima")
	ErrLetterAlreadyGenerated = errors.New("surat penerimaan untuk pendaftaran ini sudah pernah dibuat")
	ErrLetterRequiresFinalStatus = errors.New("surat hanya dapat dibuat untuk pengajuan berstatus diterima atau ditolak")
	ErrAcceptanceLetterMissingFields = errors.New("nomor surat, penerima, dan data penandatangan wajib diisi untuk surat penerimaan")
)