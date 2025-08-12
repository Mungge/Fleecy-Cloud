package aggregator

import (
	"github.com/Mungge/Fleecy-Cloud/services"
)

// OptimizationServiceAdapter는 기존 services.OptimizationService를 새로운 인터페이스에 맞게 어댑트합니다
type OptimizationServiceAdapter struct {
	service *services.OptimizationService
}

// NewOptimizationServiceAdapter는 새로운 어댑터 인스턴스를 생성합니다
func NewOptimizationServiceAdapter(service *services.OptimizationService) OptimizationService {
	return &OptimizationServiceAdapter{
		service: service,
	}
}

// ValidatePythonEnvironment은 Python 환경을 검증합니다
func (a *OptimizationServiceAdapter) ValidatePythonEnvironment() error {
	return a.service.ValidatePythonEnvironment()
}

// ValidatePythonScript는 Python 스크립트를 검증합니다
func (a *OptimizationServiceAdapter) ValidatePythonScript() error {
	return a.service.ValidatePythonScript()
}

// RunOptimization은 최적화를 실행합니다
func (a *OptimizationServiceAdapter) RunOptimization(request OptimizationRequest) (interface{}, error) {
	// 타입 변환: aggregator.OptimizationRequest -> services.OptimizationRequest
	serviceRequest := services.OptimizationRequest{
		FederatedLearning: struct {
			Name          string                `json:"name"`
			Description   string                `json:"description"`
			ModelType     string                `json:"modelType"`
			Algorithm     string                `json:"algorithm"`
			Rounds        int                   `json:"rounds"`
			Participants  []services.Participant `json:"participants"`
			ModelFileName *string               `json:"modelFileName,omitempty"`
		}{
			Name:          request.FederatedLearning.Name,
			Description:   request.FederatedLearning.Description,
			ModelType:     "default", // 기본값 설정
			Algorithm:     request.FederatedLearning.Algorithm,
			Rounds:        request.FederatedLearning.Rounds,
			ModelFileName: nil, // 기본값 nil
		},
		AggregatorConfig: struct {
			MaxBudget  int `json:"maxBudget"`
			MaxLatency int `json:"maxLatency"`
		}{
			MaxBudget:  request.AggregatorConfig.MaxBudget,
			MaxLatency: request.AggregatorConfig.MaxLatency,
		},
	}

	// Participants 변환
	for _, p := range request.FederatedLearning.Participants {
		participant := services.Participant{
			ID:                p.ID,
			Name:              p.Name,
			Status:            "active", // 기본값 설정
			Region:            "default", // 기본값 설정 (필요시 요청에서 추가)
			OpenstackEndpoint: p.OpenstackEndpoint,
		}
		serviceRequest.FederatedLearning.Participants = append(
			serviceRequest.FederatedLearning.Participants,
			participant,
		)
	}

	return a.service.RunOptimization(serviceRequest)
}
