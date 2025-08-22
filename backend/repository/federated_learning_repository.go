package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FederatedLearningRepository는 연합학습 모델의 데이터 액세스 계층입니다
type FederatedLearningRepository struct {
	db *gorm.DB
}

// NewFederatedLearningRepository는 새 FederatedLearningRepository 인스턴스를 생성합니다
func NewFederatedLearningRepository(db *gorm.DB) *FederatedLearningRepository {
	return &FederatedLearningRepository{db: db}
}

// Create는 새로운 연합학습 작업을 생성합니다
func (r *FederatedLearningRepository) Create(fl *models.FederatedLearning) error {
	// 고유한 ID 생성
	if fl.ID == "" {
		fl.ID = uuid.New().String()
	}

	return r.db.Create(fl).Error
}

// CreateWithParticipants는 참여자와 함께 연합학습을 생성합니다
func (r *FederatedLearningRepository) CreateWithParticipants(fl *models.FederatedLearning, participantIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 고유한 ID 생성
		if fl.ID == "" {
			fl.ID = uuid.New().String()
		}

		// 연합학습 생성
		if err := tx.Create(fl).Error; err != nil {
			return err
		}

		// 참여자 관계 생성
		for _, participantID := range participantIDs {
			pfl := &models.ParticipantFederatedLearning{
				ParticipantID:       participantID,
				FederatedLearningID: fl.ID,
				Status:              "active",
			}
			if err := tx.Create(pfl).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByUserID는 사용자 ID로 연합학습 목록을 조회합니다
func (r *FederatedLearningRepository) GetByUserID(userID int64) ([]*models.FederatedLearning, error) {
	var learnings []*models.FederatedLearning
	err := r.db.Preload("Participants").Where("user_id = ?", userID).Find(&learnings).Error
	return learnings, err
}

// GetByID는 ID로 연합학습을 조회합니다
func (r *FederatedLearningRepository) GetByID(id string) (*models.FederatedLearning, error) {
	var learning models.FederatedLearning
	err := r.db.Preload("Participants").Where("id = ?", id).First(&learning).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &learning, nil
}

// Update는 연합학습 정보를 업데이트합니다
func (r *FederatedLearningRepository) Update(fl *models.FederatedLearning) error {
	return r.db.Save(fl).Error
}

// UpdateStatus는 연합학습 상태를 업데이트합니다
func (r *FederatedLearningRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.FederatedLearning{}).Where("id = ?", id).Update("status", status).Error
}

// Delete는 연합학습을 삭제합니다
func (r *FederatedLearningRepository) Delete(id string) error {
	return r.db.Delete(&models.FederatedLearning{}, "id = ?", id).Error
}

// GetByStatus는 상태로 연합학습 목록을 조회합니다
func (r *FederatedLearningRepository) GetByStatus(status string) ([]*models.FederatedLearning, error) {
	var learnings []*models.FederatedLearning
	err := r.db.Where("status = ?", status).Find(&learnings).Error
	return learnings, err
}

// GetAll은 모든 연합학습을 조회합니다
func (r *FederatedLearningRepository) GetAll() ([]*models.FederatedLearning, error) {
	var learnings []*models.FederatedLearning
	err := r.db.Find(&learnings).Error
	return learnings, err
}
