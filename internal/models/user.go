package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // Hidden from JSON response
	Email    string `gorm:"uniqueIndex" json:"email"`
	Role     string `gorm:"not null;default:'user'" json:"role"`
	Remember bool   `gorm:"not null;default:false" json:"remember"`
}
