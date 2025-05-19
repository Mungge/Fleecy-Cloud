package models

import (
	"time"
)

type CloudConnection struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Provider       string    `json:"provider"`
	Name           string    `json:"name"`
	Region         string    `json:"region"`
	Status         string    `json:"status"`
	CredentialFile []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
} 