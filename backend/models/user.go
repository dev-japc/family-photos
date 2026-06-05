package models

import (
	"time"
)

// AllowedNationalIdNumber representa la lista blanca de familiares autorizados.
type AllowedNationalIdNumber struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	NationalIdNumber string    `gorm:"unique;not null;type:varchar(20)" json:"national_id_number"`
	Name             string    `gorm:"not null" json:"name"`
	CreatedAt        time.Time `json:"created_at"`
}

// User representa a los usuarios registrados.
type User struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	FullName         string    `gorm:"not null" json:"full_name"`
	Email            string    `gorm:"unique;not null" json:"email"`
	NationalIdNumber string    `gorm:"unique;not null;type:varchar(20)" json:"national_id_number"`
	Password         string    `gorm:"not null" json:"-"`
	Role             string    `gorm:"default:'family_member';not null" json:"role"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}