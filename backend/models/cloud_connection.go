package models

import (
	"time"
)

type CloudConnection struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	UserID         int64     `json:"user_id" gorm:"not null;index"`
	Provider       string    `json:"provider" gorm:"not null"`
	Name           string    `json:"name" gorm:"not null"`
	Region         string    `json:"region"`
	Zone           string    `json:"zone"`
	Status         string    `json:"status" gorm:"default:inactive"`
	CredentialFile []byte    `json:"-" gorm:"type:bytea"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CloudConnection) TableName() string {
	return "cloud_connections"
} 