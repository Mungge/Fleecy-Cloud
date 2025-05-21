package models

import (
	"time"
)

type CloudConnection struct {
	ID             string    `json:"id"`
	UserID         int64     `json:"user_id"`
	Provider       string    `json:"provider"`
	Name           string    `json:"name"`
	Region         string    `json:"region"`
	Zone           string    `json:"zone"`
	Status         string    `json:"status"`
	CredentialFile []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
} 