package models

import "time"

// ParticipantFederatedLearning은 참여자와 연합학습 간의 ManyToMany 관계를 나타내는 중간 테이블
type ParticipantFederatedLearning struct {
	ParticipantID       string    `json:"participant_id" gorm:"primaryKey"`
	FederatedLearningID string    `json:"federated_learning_id" gorm:"primaryKey"`
	JoinedAt            time.Time `json:"joined_at" gorm:"autoCreateTime"`
	Status              string    `json:"status" gorm:"default:active"` // active, inactive, completed, failed
	TasksCompleted      int       `json:"tasks_completed" gorm:"default:0"`
	LastTaskCompletedAt *time.Time `json:"last_task_completed_at,omitempty"`
	
	// 관계 설정
	Participant       Participant       `json:"participant,omitempty" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FederatedLearning FederatedLearning `json:"federated_learning,omitempty" gorm:"foreignKey:FederatedLearningID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ParticipantFederatedLearning) TableName() string {
	return "participant_federated_learnings"
}
