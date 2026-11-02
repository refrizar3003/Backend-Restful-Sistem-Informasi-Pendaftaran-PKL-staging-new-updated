package repository

import (
	"fmt"
	"internship-telkom/model"
	"time"

	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Save(app model.InternshipApplication) (model.InternshipApplication, error)
	SaveDocument(doc model.Document) error
	FindActiveByUserID(userID uint) (model.InternshipApplication, error)
	FindByID(id uint) (model.InternshipApplication, error)
	FindByUserID(userID uint) ([]model.InternshipApplication, error)
	FindByIDAndUserID(id uint, userID uint) (model.InternshipApplication, error)
	FindAllForAdmin(status, category, search string, page, limit int) ([]model.InternshipApplication, int64, error)
	FindByIDForAdmin(id uint) (model.InternshipApplication, error)
	UpdateStatus(id uint, status string, adminNotes *string, unitID *uint, mentorID *uint, attendanceDate *time.Time) error

	CountByStatus(status string) (int64, error)
	CountByStatusAndEducationLevel(status, educationLevel string) (int64, error)
	CountByStatusAndYear(status string, year int) (int64, error)
	CountByYear(year int) (int64, error)
	FindRecentApplications(limit int) ([]model.InternshipApplication, error)
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db}
}

func (r *applicationRepository) Save(app model.InternshipApplication) (model.InternshipApplication, error) {
	if err := r.db.Create(&app).Error; err != nil {
		return model.InternshipApplication{}, fmt.Errorf("failed to save application: %w", err)
	}
	return app, nil
}

func (r *applicationRepository) SaveDocument(doc model.Document) error {
	return r.db.Create(&doc).Error
}

func (r *applicationRepository) FindActiveByUserID(userID uint) (model.InternshipApplication, error) {
	var app model.InternshipApplication
	err := r.db.Where(
		"user_id = ? AND status IN ?",
		userID,
		[]string{model.StatusPending, model.StatusActive},
	).First(&app).Error
	if err != nil {
		return model.InternshipApplication{}, fmt.Errorf("no active application found for user %d: %w", userID, err)
	}
	return app, nil
}

func (r *applicationRepository) FindByID(id uint) (model.InternshipApplication, error) {
	var app model.InternshipApplication
	err := r.db.Preload("Documents").First(&app, id).Error
	if err != nil {
		return model.InternshipApplication{}, fmt.Errorf("application with id %d not found: %w", id, err)
	}
	return app, nil
}

func (r *applicationRepository) FindByUserID(userID uint) ([]model.InternshipApplication, error) {
	var apps []model.InternshipApplication
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find applications for user %d: %w", userID, err)
	}
	return apps, nil
}

func (r *applicationRepository) FindByIDAndUserID(id uint, userID uint) (model.InternshipApplication, error) {
	var app model.InternshipApplication
	err := r.db.Preload("Documents").
		Preload("User").
		Preload("Letter").
		Where("id = ? AND user_id = ?", id, userID).
		First(&app).Error
	if err != nil {
		return model.InternshipApplication{}, fmt.Errorf("application with id %d not found for user %d: %w", id, userID, err)
	}
	return app, nil
}

func (r *applicationRepository) FindAllForAdmin(status, category, search string, page, limit int) ([]model.InternshipApplication, int64, error) {
	var apps []model.InternshipApplication
	var total int64

	query := r.db.Model(&model.InternshipApplication{}).Preload("User")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if category != "" {
		query = query.Where("education_level = ?", category)
	}
	if search != "" {
		query = query.Joins("JOIN users ON users.id = internship_applications.user_id").
			Where("users.fullname ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count applications: %w", err)
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("internship_applications.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&apps).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find applications for admin: %w", err)
	}

	return apps, total, nil
}

func (r *applicationRepository) FindByIDForAdmin(id uint) (model.InternshipApplication, error) {
	var app model.InternshipApplication
	err := r.db.
		Preload("User").
		Preload("Documents").
		Preload("Letter").
		Preload("Unit").
		Preload("Mentor").
		First(&app, id).Error
	if err != nil {
		return model.InternshipApplication{}, fmt.Errorf("application with id %d not found: %w", id, err)
	}
	return app, nil
}

func (r *applicationRepository) UpdateStatus(
	id uint,
	status string,
	adminNotes *string,
	unitID *uint,
	mentorID *uint,
	attendanceDate *time.Time,
) error {
	updates := map[string]interface{}{
		"status":          status,
		"admin_notes":     adminNotes,
		"unit_id":         unitID,
		"mentor_id":       mentorID,
		"attendance_date": attendanceDate,
	}

	result := r.db.Model(&model.InternshipApplication{}).
		Where("id = ?", id).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update application status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("application with id %d not found", id)
	}
	return nil
}

func (r *applicationRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.InternshipApplication{}).
		Where("status = ?", status).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count applications by status: %w", err)
	}
	return count, nil
}

func (r *applicationRepository) CountByStatusAndEducationLevel(status, educationLevel string) (int64, error) {
	var count int64
	err := r.db.Model(&model.InternshipApplication{}).
		Where("status = ? AND education_level = ?", status, educationLevel).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count applications by status and education level: %w", err)
	}
	return count, nil
}

func (r *applicationRepository) CountByStatusAndYear(status string, year int) (int64, error) {
	var count int64
	err := r.db.Model(&model.InternshipApplication{}).
		Where("status = ? AND EXTRACT(YEAR FROM created_at) = ?", status, year).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count applications by status and year: %w", err)
	}
	return count, nil
}

func (r *applicationRepository) CountByYear(year int) (int64, error) {
	var count int64
	err := r.db.Model(&model.InternshipApplication{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count applications by year: %w", err)
	}
	return count, nil
}

func (r *applicationRepository) FindRecentApplications(limit int) ([]model.InternshipApplication, error) {
	var apps []model.InternshipApplication
	err := r.db.
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&apps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find recent applications: %w", err)
	}
	return apps, nil
}