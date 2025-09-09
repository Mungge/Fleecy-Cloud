package aggregatorvalidator

import (
	"fmt"
	"strings"

	"github.com/Mungge/Fleecy-Cloud/services/aggregator"
)

// ValidateCreateAggregatorRequest는 Aggregator 생성 요청을 검증합니다
func ValidateCreateAggregatorRequest(name, algorithm, region, storage, instanceType, estimatedCost string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("aggregator 이름이 필요합니다")
	}

	if len(name) > 100 {
		return fmt.Errorf("aggregator 이름은 100자를 초과할 수 없습니다")
	}

	if strings.TrimSpace(algorithm) == "" {
		return fmt.Errorf("집계 알고리즘이 필요합니다")
	}

	validAlgorithms := []string{"fedavg", "fedprox", "fedopt", "scaffold"}
	if !contains(validAlgorithms, strings.ToLower(algorithm)) {
		return fmt.Errorf("지원하지 않는 알고리즘입니다: %s", algorithm)
	}

	if strings.TrimSpace(region) == "" {
		return fmt.Errorf("지역 정보가 필요합니다")
	}

	if strings.TrimSpace(storage) == "" {
		return fmt.Errorf("스토리지 정보가 필요합니다")
	}

	if strings.TrimSpace(instanceType) == "" {
		return fmt.Errorf("인스턴스 타입이 필요합니다")
	}

	if strings.TrimSpace(estimatedCost) == "" {
		return fmt.Errorf("예상 비용 정보가 필요합니다")
	}

	return nil
}

// ValidateUpdateStatusRequest는 상태 업데이트 요청을 검증합니다
func ValidateUpdateStatusRequest(status string) error {
	if strings.TrimSpace(status) == "" {
		return fmt.Errorf("상태 정보가 필요합니다")
	}

	validStatuses := []string{"creating", "running", "stopped", "failed", "terminated"}
	if !contains(validStatuses, strings.ToLower(status)) {
		return fmt.Errorf("유효하지 않은 상태입니다: %s", status)
	}

	return nil
}

// ValidateUpdateMetricsRequest는 메트릭 업데이트 요청을 검증합니다
func ValidateUpdateMetricsRequest(cpuUsage, memoryUsage, networkUsage float64) error {
	if cpuUsage < 0 || cpuUsage > 100 {
		return fmt.Errorf("CPU 사용률은 0-100 사이여야 합니다: %.2f", cpuUsage)
	}

	if memoryUsage < 0 || memoryUsage > 100 {
		return fmt.Errorf("메모리 사용률은 0-100 사이여야 합니다: %.2f", memoryUsage)
	}

	if networkUsage < 0 {
		return fmt.Errorf("네트워크 사용량은 0 이상이어야 합니다: %.2f", networkUsage)
	}

	return nil
}

// ValidateOptimizationRequest는 최적화 요청을 검증합니다
func ValidateOptimizationRequest(request *aggregator.OptimizationRequest) error {
	// 1. 연합학습 정보 검증
	if strings.TrimSpace(request.FederatedLearning.Name) == "" {
		return fmt.Errorf("연합학습 이름이 필요합니다")
	}

	if strings.TrimSpace(request.FederatedLearning.Algorithm) == "" {
		return fmt.Errorf("집계 알고리즘이 필요합니다")
	}

	if request.FederatedLearning.Rounds <= 0 {
		return fmt.Errorf("라운드 수는 1 이상이어야 합니다")
	}

	if request.FederatedLearning.Rounds > 1000 {
		return fmt.Errorf("라운드 수가 너무 큽니다 (최대: 1000)")
	}

	// 2. 참여자 정보 검증
	if len(request.FederatedLearning.Participants) == 0 {
		return fmt.Errorf("최소 1명 이상의 참여자가 필요합니다")
	}

	if len(request.FederatedLearning.Participants) > 100 {
		return fmt.Errorf("참여자 수가 너무 많습니다 (최대: 100명)")
	}

	for i, participant := range request.FederatedLearning.Participants {
		if err := validateParticipant(participant, i+1); err != nil {
			return err
		}
	}

	// 3. 제약사항 검증
	if request.AggregatorConfig.MaxBudget <= 0 {
		return fmt.Errorf("최대 예산은 0보다 커야 합니다")
	}

	if request.AggregatorConfig.MaxLatency <= 0 {
		return fmt.Errorf("최대 지연시간은 0보다 커야 합니다")
	}

	// 합리적인 범위 체크
	if request.AggregatorConfig.MaxBudget > 10000000 { // 1천만원
		return fmt.Errorf("최대 예산이 너무 큽니다 (최대: 10,000,000원)")
	}

	if request.AggregatorConfig.MaxLatency > 10000 { // 10초
		return fmt.Errorf("최대 지연시간이 너무 큽니다 (최대: 10,000ms)")
	}

	return nil
}

// validateParticipant는 참여자 정보를 검증합니다
func validateParticipant(participant aggregator.Participant, index int) error {
	if strings.TrimSpace(participant.ID) == "" {
		return fmt.Errorf("참여자 %d의 ID가 필요합니다", index)
	}

	if strings.TrimSpace(participant.Name) == "" {
		return fmt.Errorf("참여자 %d의 이름이 필요합니다", index)
	}

	if strings.TrimSpace(participant.OpenstackEndpoint) == "" {
		return fmt.Errorf("참여자 %d의 OpenStack 엔드포인트가 필요합니다", index)
	}

	// 간단한 URL 형식 검증
	endpoint := strings.TrimSpace(participant.OpenstackEndpoint)
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return fmt.Errorf("참여자 %d의 OpenStack 엔드포인트가 올바른 URL 형식이 아닙니다", index)
	}

	return nil
}

// contains는 슬라이스에 특정 항목이 포함되어 있는지 확인합니다
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}