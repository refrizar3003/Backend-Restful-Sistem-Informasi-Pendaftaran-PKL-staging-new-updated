package repository

import (
	"fmt"
	"internship-telkom/model"

	"gorm.io/gorm"
)

// UserRepository adalah antarmuka untuk operasi database User
type UserRepository interface {
	Save(user model.User) (model.User, error)
	FindByEmail(email string) (model.User, error)
	FindByID(id uint) (model.User, error)
	UpdateUser(user *model.User) error
	UpdatePassword(id uint, newHash string) error
	DeleteByID(id uint) error

	FindStaffByID(id uint) (model.User, error)
	FindAllStaff() ([]model.User, error)
	ExistsByEmailExceptID(email string, excludeID uint) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru dari userRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

// Save menyimpan data pengguna baru ke database
func (r *userRepository) Save(user model.User) (model.User, error) {
	if err := r.db.Create(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("failed to save user: %w", err)
	}
	return user, nil
}

// FindByEmail mencari pengguna berdasarkan alamat email
func (r *userRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("user with email %s not found: %w", email, err)
	}
	return user, nil
}

// FindByID mencari pengguna berdasarkan ID
func (r *userRepository) FindByID(id uint) (model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return model.User{}, fmt.Errorf("user with id %d not found: %w", id, err)
	}
	return user, nil
}

func (r *userRepository) UpdateUser(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) UpdatePassword(id uint, newHash string) error {
	result := r.db.Model(&model.User{}).Where("id = ?", id).Update("password", newHash)
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}

func (r *userRepository) DeleteByID(id uint) error {
	result := r.db.Unscoped().Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *userRepository) FindStaffByID(id uint) (model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND role IN ?", id, []string{model.RoleAdmin, model.RolePegawai}).
		First(&user).Error
	if err != nil {
		return model.User{}, fmt.Errorf("staff with id %d not found: %w", id, err)
	}
	return user, nil
}

func (r *userRepository) FindAllStaff() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("role IN ?", []string{model.RoleAdmin, model.RolePegawai}).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find staff users: %w", err)
	}
	return users, nil
}

func (r *userRepository) ExistsByEmailExceptID(email string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.User{}).Where("email = ?", email)
	if excludeID != 0 {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}