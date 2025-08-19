package models

import (
	"time"
)

// SSHKeypair SSH 키페어 모델 (암호화 저장)
type SSHKeypair struct {
	ID                  int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	AggregatorID        string    `json:"aggregator_id" gorm:"not null;index;size:255"`
	KeyName             string    `json:"key_name" gorm:"not null;size:255"`
	CloudProvider       string    `json:"cloud_provider" gorm:"not null;size:20"` // aws, gcp
	Region              string    `json:"region" gorm:"not null;size:50"`
	PublicKey           string    `json:"public_key" gorm:"type:text"`
	EncryptedPrivateKey string    `json:"-" gorm:"type:text;not null"` // 암호화된 Private Key (JSON 응답에서 제외)
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 테이블 이름 설정
func (SSHKeypair) TableName() string {
	return "ssh_keypairs"
}

// SSHKeypairResponse SSH 키페어 응답 (Private Key는 별도 요청으로만 제공)
type SSHKeypairResponse struct {
	ID            int64     `json:"id"`
	AggregatorID  string    `json:"aggregator_id"`
	KeyName       string    `json:"key_name"`
	CloudProvider string    `json:"cloud_provider"`
	Region        string    `json:"region"`
	PublicKey     string    `json:"public_key"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SSHKeypairWithPrivateKey Private Key가 포함된 응답 (다운로드용)
type SSHKeypairWithPrivateKey struct {
	SSHKeypairResponse
	PrivateKey string `json:"private_key"`
}
