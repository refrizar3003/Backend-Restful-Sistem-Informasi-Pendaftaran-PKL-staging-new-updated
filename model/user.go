package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleApplicant = "applicant"
	RolePegawai   = "pegawai"
	RoleAdmin     = "admin"
)

type User struct {
	ID                 uint           `gorm:"primaryKey"                             json:"id"`
	Email              string         `gorm:"type:varchar(255);unique;not null"      json:"email"`
	Password           string         `gorm:"not null"                               json:"-"`
	Fullname           string         `gorm:"type:varchar(255);not null"             json:"fullname"`
	Gender             string         `gorm:"type:varchar(10)"                       json:"gender"`
	DateOfBirth        *time.Time     `gorm:"default:null"                           json:"date_of_birth"`
	PhoneNumber        string         `gorm:"type:varchar(20)"                       json:"phone_number"`
	Address            string         `gorm:"type:text"                              json:"address"`
	Province           string         `gorm:"type:varchar(100)"                      json:"province"`
	Role               string         `gorm:"type:varchar(50);default:'applicant'"   json:"role"`
	ProfilePhoto       string         `gorm:"type:varchar(255);default:null"         json:"profile_photo"`
	IsVerified         bool           `gorm:"default:false"                          json:"is_verified"`
	VerificationToken  *string        `gorm:"size:255;default:null"                  json:"-"`
	IsActive           bool           `gorm:"default:true"                           json:"is_active"`
	CreatedAt          time.Time      `                                              json:"created_at"`
	UpdatedAt          time.Time      `                                              json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index"                                  json:"-"`
	UserTokens         []UserToken    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TrustedDevices []TrustedDevice `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}


type UserToken struct {
	ID        uint           `gorm:"primaryKey"                                         json:"id"`
	UserID    uint           `gorm:"not null;index"                                     json:"user_id"`
	FcmToken  string         `gorm:"column:fcm_token;not null"                          json:"fcm_token"`
	CreatedAt time.Time      `                                                          json:"created_at"`
	UpdatedAt time.Time      `                                                          json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                              json:"-"`
}

type TrustedDevice struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"not null;index"`
	TokenHash  string     `gorm:"type:char(64);not null;uniqueIndex"`
	UserAgent  string     `gorm:"type:varchar(255)"`
	ExpiresAt  time.Time  `gorm:"not null;index"`
	LastUsedAt *time.Time `gorm:"default:null"`
	CreatedAt  time.Time
}