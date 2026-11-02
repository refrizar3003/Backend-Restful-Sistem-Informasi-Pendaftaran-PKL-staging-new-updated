package repository

import (
	"fmt"
	"internship-telkom/model"

	"gorm.io/gorm"
)

type LetterRepository interface {
	Save(letter model.Letter) (model.Letter, error)
	FindByApplicationID(applicationID uint) (model.Letter, error)
}

type letterRepository struct {
	db *gorm.DB
}

func NewLetterRepository(db *gorm.DB) LetterRepository {
	return &letterRepository{db}
}

func (r *letterRepository) Save(letter model.Letter) (model.Letter, error) {
	if err := r.db.Create(&letter).Error; err != nil {
		return model.Letter{}, fmt.Errorf("failed to save letter: %w", err)
	}
	return letter, nil
}

func (r *letterRepository) FindByApplicationID(applicationID uint) (model.Letter, error) {
	var letter model.Letter
	err := r.db.Where("internship_application_id = ?", applicationID).First(&letter).Error
	if err != nil {
		return model.Letter{}, fmt.Errorf("letter for application %d not found: %w", applicationID, err)
	}
	return letter, nil
}
