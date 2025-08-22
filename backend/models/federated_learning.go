package models

import "time"

type FederatedLearning struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	UserID            int64      `json:"user_id" gorm:"not null;index"`
	CloudConnectionID string     `json:"cloud_connection_id" gorm:"not null;index"` // CloudConnection 참조
	AggregatorID      *string    `json:"aggregator_id,omitempty" gorm:"index"`      // 1:1 관계를 위한 aggregator ID
	Name              string     `json:"name" gorm:"not null"`
	Description       string     `json:"description"`
	Status            string     `json:"status" gorm:"default:inactive"`
	ParticipantCount  int        `json:"participant_count" gorm:"default:0"`
	CompletedAt       *time.Time `json:"completed_at"`
	Accuracy          string     `json:"accuracy"`
	Rounds            int        `json:"rounds" gorm:"default:0"`
	Algorithm         string     `json:"algorithm"`
	ModelType         string     `json:"model_type"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 관계 설정
	User            *User            `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CloudConnection *CloudConnection `json:"cloud_connection,omitempty" gorm:"foreignKey:CloudConnectionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Aggregator      *Aggregator      `json:"aggregator,omitempty" gorm:"foreignKey:AggregatorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Participants    []Participant    `json:"participants" gorm:"many2many:participant_federated_learnings;"`
}

func (FederatedLearning) TableName() string {
	return "federated_learnings"
}
