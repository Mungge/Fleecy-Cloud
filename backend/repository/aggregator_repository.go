package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errors"
)

type AggregatorRepository struct {
	db *gorm.DB
}

func NewAggregatorRepository(db *gorm.DB) *AggregatorRepository {
	return &AggregatorRepository{db: db}
}

func (r *AggregatorRepository) CreateAggregator(aggregator *models.Aggregator) error {
	// UUID 생성
	aggregator.ID = uuid.New().String()
	return r.db.Create(aggregator).Error
}

func (r *AggregatorRepository) GetAggregatorsByUserID(userID int64) ([]*models.Aggregator, error) {
	var aggregators []*models.Aggregator
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&aggregators).Error
	return aggregators, err
}

func (r *AggregatorRepository) GetAggregatorByID(id string) (*models.Aggregator, error) {
	var aggregator models.Aggregator
	err := r.db.Where("id = ?", id).
		First(&aggregator).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &aggregator, nil
}

func (r *AggregatorRepository) GetAggregatorByIDWithFederatedLearning(id string) (*models.Aggregator, error) {
	var aggregator models.Aggregator
	err := r.db.Where("id = ?", id).
		Preload("FederatedLearning").
		First(&aggregator).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &aggregator, nil
}

func (r *AggregatorRepository) UpdateAggregator(aggregator *models.Aggregator) error {
	return r.db.Save(aggregator).Error
}

func (r *AggregatorRepository) UpdateAggregatorStatus(id string, status string) error {
	return r.db.Model(&models.Aggregator{}).Where("id = ?", id).Update("status", status).Error
}

func (r *AggregatorRepository) UpdateAggregatorMLflowInfo(id, experimentID, experimentName string) error {
	return r.db.Model(&models.Aggregator{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"mlflow_experiment_id":   experimentID,
			"mlflow_experiment_name": experimentName,
		}).Error
}

func (r *AggregatorRepository) UpdateAggregatorMetrics(id string, cpuUsage, memoryUsage, networkUsage float64) error {
	return r.db.Model(&models.Aggregator{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"cpu_usage":     cpuUsage,
			"memory_usage":  memoryUsage,
			"network_usage": networkUsage,
		}).Error
}

func (r *AggregatorRepository) UpdateAggregatorIPInfo(id, instanceID, publicIP, privateIP string) error {
	return r.db.Model(&models.Aggregator{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"instance_id": instanceID,
			"public_ip":   publicIP,
			"private_ip":  privateIP,
		}).Error
}

func (r *AggregatorRepository) UpdateAggregatorProgress(id string, currentRound int, accuracy *float64) error {
	updates := map[string]interface{}{
		"current_round": currentRound,
	}
	if accuracy != nil {
		updates["accuracy"] = *accuracy
	}
	return r.db.Model(&models.Aggregator{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AggregatorRepository) DeleteAggregator(id string) error {
	return r.db.Delete(&models.Aggregator{}, "id = ?", id).Error
}

// Training Round 관련 메서드들
func (r *AggregatorRepository) CreateTrainingRound(round *models.TrainingRound) error {
	var existing models.TrainingRound
    err := r.db.Where("aggregator_id = ? AND round = ?", round.AggregatorID, round.Round).
        First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 새로 생성
		round.ID = uuid.New().String()
		return r.db.Create(round).Error
	} else if err != nil {
		return err
	}
	
	// 기존 데이터 업데이트
	round.ID = existing.ID
	return r.db.Save(round).Error
}

func (r *AggregatorRepository) GetTrainingRoundsByAggregatorID(aggregatorID string) ([]*models.TrainingRound, error) {
	var rounds []*models.TrainingRound
	err := r.db.Where("aggregator_id = ?", aggregatorID).
		Order("round ASC").
		Find(&rounds).Error
	return rounds, err
}

func (r *AggregatorRepository) UpdateTrainingRound(round *models.TrainingRound) error {
	return r.db.Save(round).Error
}

// Aggregator 통계 메서드들
func (r *AggregatorRepository) GetAggregatorStats(userID int64) (map[string]interface{}, error) {
	var total, running, completed, pending int64

	r.db.Model(&models.Aggregator{}).Where("user_id = ?", userID).Count(&total)
	r.db.Model(&models.Aggregator{}).Where("user_id = ? AND status = ?", userID, "running").Count(&running)
	r.db.Model(&models.Aggregator{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completed)
	r.db.Model(&models.Aggregator{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&pending)

	var totalCost float64
	r.db.Model(&models.Aggregator{}).Where("user_id = ?", userID).Select("COALESCE(SUM(current_cost), 0)").Scan(&totalCost)

	return map[string]interface{}{
		"total":      total,
		"running":    running,
		"completed":  completed,
		"pending":    pending,
		"total_cost": totalCost,
	}, nil
}

// GetAggregatorsByStatus는 특정 상태의 집계자들을 조회합니다
func (r *AggregatorRepository) GetAggregatorsByStatus(status string) ([]*models.Aggregator, error) {
	var aggregators []*models.Aggregator
	err := r.db.Where("status = ?", status).
		Find(&aggregators).Error
	return aggregators, err
}
