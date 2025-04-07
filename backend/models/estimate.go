package models

import "math"

// EstimateResources 계산 함수
func EstimateResources(input UserInput) ResourceEstimate {
	// Aggregation 타입에 따라 가중치 설정
	var aggWeight float64
	switch input.AggregationType {
	case "FedAvg":
		aggWeight = 0.1
	case "FedKD":
		aggWeight = 0.5
	case "FedProto":
		aggWeight = 0.7
	default:
		aggWeight = 0.1
	}

	// 복잡도 보정 계수 (간단한 추정)
	complexityFactor := 1.0 + input.ModelSizeMB/50.0

	// RAM (MB): 클라이언트 수 × 모델 크기 × 복잡도 계수
	ram := float64(input.MaxClients) * input.ModelSizeMB * complexityFactor

	// CPU (%): 클라이언트 수 × (FLOPs / 1e9) × aggregation 계수
	cpu := float64(input.MaxClients) * (input.FLOPs / 1e9) * aggWeight

	// Network (MBps): 업로드 + 다운로드 기준 (분당 트래픽 계산 후 초당 전환)
	net := float64(input.MaxClients) * input.ModelSizeMB * 2 / float64(input.UploadFreqMin) / 60

	return ResourceEstimate{
		RAMMB:      int(math.Ceil(ram)),
		CPUPercent: int(math.Ceil(cpu * 100)), // 소수 → 정수 %
		NetMBps:    int(math.Round(net*100) / 100), // 소수 2자리로 반올림 후 정수 변환
	}
}