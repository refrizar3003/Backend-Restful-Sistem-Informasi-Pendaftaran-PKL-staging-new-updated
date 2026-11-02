package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusActive   = "active"
	StatusDone     = "done"
)

type InternshipApplication struct {
	gorm.Model
	UserID              uint       `gorm:"not null;index"                                         json:"user_id"`
	User                User       `gorm:"foreignKey:UserID"                                      json:"-"`
	EducationLevel      string     `gorm:"type:varchar(50);not null"                              json:"education_level"`
	IdentityNumber      string     `gorm:"type:varchar(50);not null"                              json:"identity_number"`
	InstitutionName     string     `gorm:"type:varchar(255);not null"                             json:"institution_name"`
	Major               string     `gorm:"type:varchar(100);not null"                             json:"major"`
	Semester            int        `gorm:"not null"                                               json:"semester"`
	StartDate           time.Time  `gorm:"not null"                                               json:"start_date"`
	EndDate             time.Time  `gorm:"not null"                                               json:"end_date"`
	Status              string     `gorm:"type:varchar(20);default:'pending'"                     json:"status"`
	AdminNotes          *string    `gorm:"type:text;default:null"                                 json:"admin_notes,omitempty"`
	AttendanceDate      *time.Time `gorm:"default:null"                                           json:"attendance_date,omitempty"`
	IntegrityPactAgreed bool       `gorm:"default:false"                                          json:"integrity_pact_agreed"`
	UnitID              *uint      `gorm:"default:null"                                           json:"unit_id,omitempty"`
	MentorID            *uint      `gorm:"default:null"                                           json:"mentor_id,omitempty"`
	Unit                *Unit      `gorm:"foreignKey:UnitID"                                      json:"-"`
	Mentor              *Mentor    `gorm:"foreignKey:MentorID"                                    json:"-"`
	Documents           []Document `gorm:"foreignKey:InternshipApplicationID;constraint:OnDelete:CASCADE" json:"documents,omitempty"`
	Letter              *Letter    `gorm:"foreignKey:InternshipApplicationID"                     json:"letter,omitempty"`
}

type Letter struct {
	gorm.Model
	InternshipApplicationID uint      `gorm:"not null;index"              json:"internship_application_id"`
	LetterNumber            string    `gorm:"type:varchar(100);not null"  json:"letter_number"`
	LetterDate              time.Time `gorm:"not null"                    json:"letter_date"`
	FileURL                 string    `gorm:"type:varchar(255);not null"  json:"file_url"`
	GeneratedAt             time.Time `gorm:"autoCreateTime"              json:"generated_at"`
}

type Document struct {
	gorm.Model
	InternshipApplicationID uint      `gorm:"not null;index"              json:"internship_application_id"`
	DocumentType            string `gorm:"type:varchar(50);not null"  json:"document_type"`
	FileURL                 string    `gorm:"type:varchar(255);not null"  json:"file_url"`
}

type Unit struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null" json:"name"`
}

type Mentor struct {
	gorm.Model
	UnitID   uint   `gorm:"not null;index"             json:"unit_id"`
	Unit     Unit   `gorm:"foreignKey:UnitID"          json:"-"`
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Position string `gorm:"type:varchar(100);not null" json:"position"`
}

type Notification struct {
	gorm.Model
	UserID                  uint                   `gorm:"not null;index"              json:"user_id"`
	User                    User                   `gorm:"foreignKey:UserID"           json:"-"`
	InternshipApplicationID uint                   `gorm:"not null;index"              json:"internship_application_id"`
	InternshipApplication   InternshipApplication  `gorm:"foreignKey:InternshipApplicationID" json:"-"`
	Description             string                 `gorm:"type:text;not null"          json:"description"`
	IsRead                  bool                   `gorm:"default:false"               json:"is_read"`
}
