package repository

import (
	"fmt"
	"internship-telkom/model"

	"gorm.io/gorm"
)

type UnitRepository interface {
	FindAll() ([]model.Unit, error)
	FindByID(id uint) (model.Unit, error)
}

type MentorRepository interface {
	FindByUnitID(unitID uint) ([]model.Mentor, error)
	FindByID(id uint) (model.Mentor, error)
}

type unitRepository struct {
	db *gorm.DB
}

func NewUnitRepository(db *gorm.DB) UnitRepository {
	return &unitRepository{db}
}

func (r *unitRepository) FindAll() ([]model.Unit, error) {
	var units []model.Unit
	if err := r.db.Order("name ASC").Find(&units).Error; err != nil {
		return nil, fmt.Errorf("failed to find units: %w", err)
	}
	return units, nil
}

func (r *unitRepository) FindByID(id uint) (model.Unit, error) {
	var unit model.Unit
	if err := r.db.First(&unit, id).Error; err != nil {
		return model.Unit{}, fmt.Errorf("unit with id %d not found: %w", id, err)
	}
	return unit, nil
}

type mentorRepository struct {
	db *gorm.DB
}

func NewMentorRepository(db *gorm.DB) MentorRepository {
	return &mentorRepository{db}
}

func (r *mentorRepository) FindByUnitID(unitID uint) ([]model.Mentor, error) {
	var mentors []model.Mentor
	if err := r.db.Where("unit_id = ?", unitID).Order("name ASC").Find(&mentors).Error; err != nil {
		return nil, fmt.Errorf("failed to find mentors for unit %d: %w", unitID, err)
	}
	return mentors, nil
}

func (r *mentorRepository) FindByID(id uint) (model.Mentor, error) {
	var mentor model.Mentor
	if err := r.db.First(&mentor, id).Error; err != nil {
		return model.Mentor{}, fmt.Errorf("mentor with id %d not found: %w", id, err)
	}
	return mentor, nil
}