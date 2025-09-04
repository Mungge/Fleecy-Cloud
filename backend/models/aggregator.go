package models

import "time"

type Aggregator struct {
	ID            string `json:"id" gorm:"primaryKey"`
	UserID        int64  `json:"user_id" gorm:"not null;index"`
	Name          string `json:"name" gorm:"not null"`
	Status        string `json:"status" gorm:"default:pending"` // pending, running, completed, error
	Algorithm     string `json:"algorithm" gorm:"not null"`     // FedAvg, FedProx, FedAdam, etc.
	CloudProvider string `json:"cloud_provider" gorm:"not null"`

	// 클라우드 공통 필드
	ProjectName  string `json:"project_name" gorm:"not null"`
	Region       string `json:"region" gorm:"not null"`
	Zone         string `json:"zone" gorm:"not null"`
	InstanceType string `json:"instance_type" gorm:"not null"`

	// GCP 전용 필드 (nullable)
	ProjectID *string `json:"project_id,omitempty"` // GCP에서만 사용

	// MLflow 연동 필드 추가
    MLflowExperimentID   *string `json:"mlflow_experiment_id,omitempty"`
    MLflowExperimentName *string `json:"mlflow_experiment_name,omitempty"`
    MLflowRunID         *string `json:"mlflow_run_id,omitempty"`

	ParticipantCount int      `json:"participant_count" gorm:"default:0"`
	CurrentRound     int      `json:"current_round" gorm:"default:0"`
	Accuracy         *float64 `json:"accuracy,omitempty"`
	CurrentCost      float64  `json:"current_cost" gorm:"default:0"`
	EstimatedCost    float64  `json:"estimated_cost" gorm:"default:0"`
	CPUSpecs         string   `json:"cpu_specs"`
	MemorySpecs      string   `json:"memory_specs"`
	StorageSpecs     string   `json:"storage_specs"`
	CPUUsage         float64  `json:"cpu_usage" gorm:"default:0"`
	MemoryUsage      float64  `json:"memory_usage" gorm:"default:0"`
	NetworkUsage     float64  `json:"network_usage" gorm:"default:0"`
	InstanceID       string   `json:"instance_id,omitempty"`
	PublicIP         string   `json:"public_ip,omitempty"`
	PrivateIP        string   `json:"private_ip,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User              User               `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FederatedLearning *FederatedLearning `json:"federated_learning,omitempty" gorm:"foreignKey:AggregatorID"`
}

func (Aggregator) TableName() string {
	return "aggregators"
}

// AggregatorMetrics는 실시간 메트릭을 위한 구조체입니다
type AggregatorMetrics struct {
	AggregatorID string    `json:"aggregator_id"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	NetworkUsage float64   `json:"network_usage"`
	Timestamp    time.Time `json:"timestamp"`
}

// TrainingRound는 학습 라운드 정보를 위한 구조체입니다
type TrainingRound struct {
	ID                string      `json:"id" gorm:"primaryKey"`
	AggregatorID      string      `json:"aggregator_id" gorm:"not null;index"`
	Round             int         `json:"round" gorm:"not null"`
	ModelMetrics      ModelMetric `json:"model_metrics" gorm:"embedded;embeddedPrefix:model_metrics_"`
	Duration          int         `json:"duration"`           // seconds
	ParticipantsCount int         `json:"participants_count"` // 왜 있는거지?
	StartedAt         time.Time   `json:"started_at"`
	CompletedAt       *time.Time  `json:"completed_at,omitempty"`
	CreatedAt         time.Time   `json:"created_at" gorm:"autoCreateTime"` // 삭제?

	// Relationships
	Aggregator *Aggregator `json:"aggregator,omitempty" gorm:"foreignKey:AggregatorID"`
}

type ModelMetric struct {
	Accuracy  *float64 `json:"accuracy,omitempty"`
	Loss      *float64 `json:"loss,omitempty"`
	Precision *float64 `json:"precision,omitempty"`
	Recall    *float64 `json:"recall,omitempty"`
	F1Score   *float64 `json:"f1_score,omitempty"`
}
