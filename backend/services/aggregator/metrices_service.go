package aggregator

import (
	"github.com/Mungge/Fleecy-Cloud/repository"
)

// AggregatorMetricsService는 Aggregator 메트릭 관련 비즈니스 로직을 처리합니다
type AggregatorMetricsService struct {
	repo *repository.AggregatorRepository
}

// NewAggregatorMetricsService는 새 AggregatorMetricsService 인스턴스를 생성합니다
func NewAggregatorMetricsService(repo *repository.AggregatorRepository) *AggregatorMetricsService {
	return &AggregatorMetricsService{
		repo: repo,
	}
}

// UpdateStatus는 Aggregator의 상태를 업데이트합니다
func (s *AggregatorMetricsService) UpdateStatus(aggregatorID string, userID int64, status string) error {
	// 권한 확인
	aggregator, err := s.repo.GetAggregatorByID(aggregatorID)
	if err != nil {
		return err
	}
	if aggregator == nil || aggregator.UserID != userID {
		return ErrAggregatorNotFound
	}

	return s.repo.UpdateAggregatorStatus(aggregatorID, status)
}

// UpdateMetrics는 Aggregator의 메트릭을 업데이트합니다
func (s *AggregatorMetricsService) UpdateMetrics(aggregatorID string, userID int64, cpuUsage, memoryUsage, networkUsage float64) error {
	// 권한 확인
	aggregator, err := s.repo.GetAggregatorByID(aggregatorID)
	if err != nil {
		return err
	}
	if aggregator == nil || aggregator.UserID != userID {
		return ErrAggregatorNotFound
	}

	return s.repo.UpdateAggregatorMetrics(aggregatorID, cpuUsage, memoryUsage, networkUsage)
}

// GetStats는 사용자의 Aggregator 통계를 조회합니다
func (s *AggregatorMetricsService) GetStats(userID int64) (map[string]interface{}, error) {
	return s.repo.GetAggregatorStats(userID)
}