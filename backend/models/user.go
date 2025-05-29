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

	// Releationships
	CloudConnections   []CloudConnection   `json:"cloud_connections,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participants       []Participant       `json:"participants,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FederatedLearnings []FederatedLearning `json:"federated_learnings,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Aggregators        []Aggregator        `json:"aggregators,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}

