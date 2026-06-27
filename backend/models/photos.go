package models

import (
	"time"
)

type Album struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`
	Photos []Photo `gorm:"foreignKey:AlbumID" json:"photos,omitempty"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Photo struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Title string `gorm:"type:varchar(100);not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	URL string `gorm:"not null" json:"url"`
	AlbumID *uint `json:"album_id"`
	UserID uint `gorm:"not null" json:"user_id"`
	User User `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}