package apperrors

import "errors"

var (
	ErrInvalidRole             = errors.New("hak akses tidak valid, gunakan: admin atau pegawai")
	ErrEmployeeEmailTaken      = errors.New("email sudah digunakan oleh akun lain")
	ErrEmployeeNotFound        = errors.New("akun pegawai tidak ditemukan")
	ErrCannotDeleteSelf        = errors.New("tidak dapat menghapus akun yang sedang digunakan untuk login")
	ErrPasswordConfirmMismatch = errors.New("password dan konfirmasi password tidak sama")
)