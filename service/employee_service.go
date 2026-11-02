package service

import (
	"context"
	"fmt"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/model"
	"internship-telkom/repository"
	"internship-telkom/utils"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type EmployeeService interface {
	CreateEmployee(input dto.CreateEmployeeRequest, photo *multipart.FileHeader) (dto.EmployeeDetailResponse, error)
	GetEmployees() ([]dto.EmployeeListItem, error)
	GetEmployeeDetail(id uint) (dto.EmployeeDetailResponse, error)
	UpdateEmployee(id uint, input dto.UpdateEmployeeRequest, photo *multipart.FileHeader) (dto.EmployeeDetailResponse, error)
	DeleteEmployee(id uint, requesterID uint) error
}

type employeeService struct {
	repo repository.UserRepository
	cld  *cloudinary.Cloudinary
}

func NewEmployeeService(repo repository.UserRepository, cld *cloudinary.Cloudinary) EmployeeService {
	return &employeeService{repo, cld}
}

func (s *employeeService) CreateEmployee(input dto.CreateEmployeeRequest, photo *multipart.FileHeader) (dto.EmployeeDetailResponse, error) {
	if !isValidStaffRole(input.Role) {
		return dto.EmployeeDetailResponse{}, apperrors.ErrInvalidRole
	}
	if input.Password != input.ConfirmPassword {
		return dto.EmployeeDetailResponse{}, apperrors.ErrPasswordConfirmMismatch
	}

	taken, err := s.repo.ExistsByEmailExceptID(input.Email, 0)
	if err != nil {
		return dto.EmployeeDetailResponse{}, err
	}
	if taken {
		return dto.EmployeeDetailResponse{}, apperrors.ErrEmployeeEmailTaken
	}

	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return dto.EmployeeDetailResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := model.User{
		Fullname:    input.Fullname,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Role:        input.Role,
		Password:    passwordHash,
		IsVerified:  true,
		IsActive:    true,
	}

	if photo != nil {
		photoURL, err := s.uploadPhoto(photo)
		if err != nil {
			return dto.EmployeeDetailResponse{}, err
		}
		user.ProfilePhoto = photoURL
	}

	created, err := s.repo.Save(user)
	if err != nil {
		return dto.EmployeeDetailResponse{}, fmt.Errorf("failed to save employee: %w", err)
	}

	return toEmployeeDetailResponse(created), nil
}

func (s *employeeService) GetEmployees() ([]dto.EmployeeListItem, error) {
	users, err := s.repo.FindAllStaff()
	if err != nil {
		return nil, err
	}

	items := make([]dto.EmployeeListItem, 0, len(users))
	for _, u := range users {
		items = append(items, dto.EmployeeListItem{
			ID:           u.ID,
			Fullname:     u.Fullname,
			Email:        u.Email,
			PhoneNumber:  u.PhoneNumber,
			Role:         u.Role,
			ProfilePhoto: u.ProfilePhoto,
		})
	}
	return items, nil
}

func (s *employeeService) GetEmployeeDetail(id uint) (dto.EmployeeDetailResponse, error) {
	user, err := s.repo.FindStaffByID(id)
	if err != nil {
		return dto.EmployeeDetailResponse{}, apperrors.ErrEmployeeNotFound
	}
	return toEmployeeDetailResponse(user), nil
}

func (s *employeeService) UpdateEmployee(id uint, input dto.UpdateEmployeeRequest, photo *multipart.FileHeader) (dto.EmployeeDetailResponse, error) {
	if !isValidStaffRole(input.Role) {
		return dto.EmployeeDetailResponse{}, apperrors.ErrInvalidRole
	}

	user, err := s.repo.FindStaffByID(id)
	if err != nil {
		return dto.EmployeeDetailResponse{}, apperrors.ErrEmployeeNotFound
	}

	taken, err := s.repo.ExistsByEmailExceptID(input.Email, id)
	if err != nil {
		return dto.EmployeeDetailResponse{}, err
	}
	if taken {
		return dto.EmployeeDetailResponse{}, apperrors.ErrEmployeeEmailTaken
	}

	user.Fullname = input.Fullname
	user.Email = input.Email
	user.PhoneNumber = input.PhoneNumber
	user.Role = input.Role

	if input.Password != "" || input.ConfirmPassword != "" {
		if len(input.Password) < 8 {
			return dto.EmployeeDetailResponse{}, fmt.Errorf("password minimal 8 karakter")
		}
		if input.Password != input.ConfirmPassword {
			return dto.EmployeeDetailResponse{}, apperrors.ErrPasswordConfirmMismatch
		}
		passwordHash, err := utils.HashPassword(input.Password)
		if err != nil {
			return dto.EmployeeDetailResponse{}, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = passwordHash
	}

	if photo != nil {
		photoURL, err := s.uploadPhoto(photo)
		if err != nil {
			return dto.EmployeeDetailResponse{}, err
		}
		user.ProfilePhoto = photoURL
	}

	if err := s.repo.UpdateUser(&user); err != nil {
		return dto.EmployeeDetailResponse{}, fmt.Errorf("failed to update employee: %w", err)
	}

	return toEmployeeDetailResponse(user), nil
}

func (s *employeeService) DeleteEmployee(id uint, requesterID uint) error {
	if id == requesterID {
		return apperrors.ErrCannotDeleteSelf
	}

	if _, err := s.repo.FindStaffByID(id); err != nil {
		return apperrors.ErrEmployeeNotFound
	}

	if err := s.repo.DeleteByID(id); err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

func (s *employeeService) uploadPhoto(photo *multipart.FileHeader) (string, error) {
	file, err := photo.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open photo file: %w", err)
	}
	defer file.Close()

	uploadRes, err := s.cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
		Folder: "employee_photos",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}
	return uploadRes.SecureURL, nil
}

func isValidStaffRole(role string) bool {
	return role == model.RoleAdmin || role == model.RolePegawai
}

func toEmployeeDetailResponse(u model.User) dto.EmployeeDetailResponse {
	return dto.EmployeeDetailResponse{
		ID:           u.ID,
		Fullname:     u.Fullname,
		Email:        u.Email,
		PhoneNumber:  u.PhoneNumber,
		Role:         u.Role,
		ProfilePhoto: u.ProfilePhoto,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt.Format("2006-01-02"),
	}
}