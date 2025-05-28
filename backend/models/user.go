package models

import (
	"time"
)

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Email        string    `json:"email" gorm:"uniqueIndex:idx_users_email;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	Name         string    `json:"name" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

