package models

import (
	"time"
)

type Photo struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Title string `gorm:"type:varchar(100);not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	URL string `gorm:"not null" json:"url"`
	UserID uint `gorm:"not null" json:"user_id"`
	User User `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}