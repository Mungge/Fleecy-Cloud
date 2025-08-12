package aggregator

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
)

// AggregatorTrainingService는 Aggregator 학습 관련 비즈니스 로직을 처리합니다
type AggregatorTrainingService struct {
	repo *repository.AggregatorRepository
}

// NewAggregatorTrainingService는 새 AggregatorTrainingService 인스턴스를 생성합니다
func NewAggregatorTrainingService(repo *repository.AggregatorRepository) *AggregatorTrainingService {
	return &AggregatorTrainingService{
		repo: repo,
	}
}

// GetTrainingHistory는 Aggregator의 학습 히스토리를 조회합니다
func (s *AggregatorTrainingService) GetTrainingHistory(aggregatorID string, userID int64) ([]*models.TrainingRound, error) {
	// 권한 확인
	aggregator, err := s.repo.GetAggregatorByID(aggregatorID)
	if err != nil {
		return nil, err
	}
	if aggregator == nil || aggregator.UserID != userID {
		return nil, ErrAggregatorNotFound
	}

	return s.repo.GetTrainingRoundsByAggregatorID(aggregatorID)
}