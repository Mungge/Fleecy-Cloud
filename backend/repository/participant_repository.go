package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ParticipantRepository는 참여자 모델의 데이터 액세스 계층입니다
type ParticipantRepository struct {
	db *gorm.DB
}

// NewParticipantRepository는 새 ParticipantRepository 인스턴스를 생성합니다
func NewParticipantRepository(db *gorm.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

// Create는 새로운 참여자를 생성합니다
func (r *ParticipantRepository) Create(participant *models.Participant) error {
	// 고유한 ID 생성
	if participant.ID == "" {
		participant.ID = uuid.New().String()
	}
	return r.db.Create(participant).Error
}

// GetByUserID는 사용자 ID로 참여자 목록을 조회합니다
func (r *ParticipantRepository) GetByUserID(userID int64) ([]*models.Participant, error) {
	var participants []*models.Participant
	err := r.db.Where("user_id = ?", userID).Find(&participants).Error
	return participants, err
}

// GetByID는 ID로 참여자를 조회합니다
func (r *ParticipantRepository) GetByID(id string) (*models.Participant, error) {
	var participant models.Participant
	err := r.db.Where("id = ?", id).First(&participant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &participant, nil
}

// Update는 참여자 정보를 업데이트합니다
func (r *ParticipantRepository) Update(participant *models.Participant) error {
	return r.db.Save(participant).Error
}

// UpdateStatus는 참여자 상태를 업데이트합니다
func (r *ParticipantRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.Participant{}).Where("id = ?", id).Update("status", status).Error
}

// Delete는 참여자를 삭제합니다
func (r *ParticipantRepository) Delete(id string) error {
	return r.db.Delete(&models.Participant{}, "id = ?", id).Error
}

// GetByStatus는 상태로 참여자 목록을 조회합니다
func (r *ParticipantRepository) GetByStatus(userID int64, status string) ([]*models.Participant, error) {
	var participants []*models.Participant
	err := r.db.Where("user_id = ? AND status = ?", userID, status).Find(&participants).Error
	return participants, err
}

// GetAvailable는 사용 가능한 참여자 목록을 조회합니다 (active 상태)
func (r *ParticipantRepository) GetAvailable(userID int64) ([]*models.Participant, error) {
	return r.GetByStatus(userID, "active")
}
