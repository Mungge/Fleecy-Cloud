package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
)

type SSHKeypairRepository struct {
	db *gorm.DB
}

func NewSSHKeypairRepository(db *gorm.DB) *SSHKeypairRepository {
	return &SSHKeypairRepository{db: db}
}

// CreateKeypair SSH 키페어 생성
func (r *SSHKeypairRepository) CreateKeypair(keypair *models.SSHKeypair) error {
	return r.db.Create(keypair).Error
}

// GetKeypairByAggregatorID 집계자 ID로 SSH 키페어 조회
func (r *SSHKeypairRepository) GetKeypairByAggregatorID(aggregatorID string) (*models.SSHKeypair, error) {
	var keypair models.SSHKeypair
	err := r.db.Where("aggregator_id = ?", aggregatorID).First(&keypair).Error
	if err != nil {
		return nil, err
	}
	return &keypair, nil
}

// GetKeypairByID ID로 SSH 키페어 조회
func (r *SSHKeypairRepository) GetKeypairByID(id int64) (*models.SSHKeypair, error) {
	var keypair models.SSHKeypair
	err := r.db.Where("id = ?", id).First(&keypair).Error
	if err != nil {
		return nil, err
	}
	return &keypair, nil
}

// UpdateKeypair SSH 키페어 업데이트
func (r *SSHKeypairRepository) UpdateKeypair(keypair *models.SSHKeypair) error {
	return r.db.Save(keypair).Error
}

// DeleteKeypairByAggregatorID 집계자 ID로 SSH 키페어 삭제
func (r *SSHKeypairRepository) DeleteKeypairByAggregatorID(aggregatorID string) error {
	return r.db.Where("aggregator_id = ?", aggregatorID).Delete(&models.SSHKeypair{}).Error
}

// ListKeypairsByUserID 사용자의 모든 SSH 키페어 목록 조회 (집계자를 통해)
func (r *SSHKeypairRepository) ListKeypairsByUserID(userID int64) ([]*models.SSHKeypair, error) {
	var keypairs []*models.SSHKeypair

	// 집계자 테이블과 조인하여 사용자의 키페어만 조회
	err := r.db.Table("ssh_keypairs").
		Joins("JOIN aggregators ON ssh_keypairs.aggregator_id = aggregators.id").
		Where("aggregators.user_id = ?", userID).
		Find(&keypairs).Error

	return keypairs, err
}
