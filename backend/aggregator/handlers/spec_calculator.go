package aggregator

import (
	"fmt"
	"math"
	"sort"
)

// 클라우드 지역 정보 구조체
type RegionInfo struct {
	Name            string  `json:"name" yaml:"name"`
	Provider        string  `json:"provider" yaml:"provider"`
	LatencyMS       float64 `json:"latency_ms" yaml:"latency_ms"`
	CostFactorCPU   float64 `json:"cost_factor_cpu" yaml:"cost_factor_cpu"`
	CostFactorRAM   float64 `json:"cost_factor_ram" yaml:"cost_factor_ram"`
	CostFactorDisk  float64 `json:"cost_factor_disk" yaml:"cost_factor_disk"`
	AvailableCPUs   []int   `json:"available_cpus" yaml:"available_cpus"`
	AvailableRAMGB  []int   `json:"available_ram_gb" yaml:"available_ram_gb"`
	AvailableDiskGB []int   `json:"available_disk_gb" yaml:"available_disk_gb"`
}

// Aggregator 설정 구조체
type AggregatorConfig struct {
	NumClients      int     `json:"num_clients"`
	ModelSizeMB     float64 `json:"model_size_mb"`
	FLOPsPerClient  float64 `json:"flops_per_client"`
	UpdateInterval  int     `json:"update_interval_seconds"`
	AggregationType string  `json:"aggregation_type"` // "fedavg", "fedprox", "scaffold" 등
	RegionPriority  string  `json:"region_priority"`  // "latency" 또는 "cost"
}

// Spec 추천 결과 구조체
type SpecRecommendation struct {
	CPU        int     `json:"cpu_cores"`
	RAMInGB    int     `json:"ram_gb"`
	DiskInGB   int     `json:"disk_gb"`
	Region     string  `json:"region"`
	Provider   string  `json:"provider"`
	TotalCost  float64 `json:"estimated_monthly_cost"`
	LatencyMS  float64 `json:"expected_latency_ms"`
	ScalingMin int     `json:"scaling_min_instances"`
	ScalingMax int     `json:"scaling_max_instances"`
}

var ComplexityFactors = map[string]float64{
	"fedavg":   1.0,  // 기본 FedAvg
	"fedprox":  1.2,  // FedProx는 추가 계산 필요
	"scaffold": 1.3,  // SCAFFOLD는 더 복잡한 계산 필요
	"fedopt":   1.4,  // FedOpt는 최적화 단계가 추가됨
	"fedadam":  1.5,  // FedAdam은 적응형 최적화가 추가됨
}

// 기본 설정값
const (
	DefaultMinRAMGB       = 4
	DefaultMinCPUCores    = 2
	DefaultMinDiskGB      = 50
	DefaultMaxStoreDays   = 30
	DefaultGFLOPsPerCore  = 10.0 // 코어당 10 GFLOPS 처리 성능 가정
	DefaultMinInstances   = 1
	DefaultMaxScalingFactor = 0.01 // 클라이언트 100명당 추가 인스턴스 1개
	DefaultMinMaxInstances = 2      // 최소한 최대 인스턴스는 2개
)

// SpecCalculator는 Aggregator 사양 계산기 구조체
type SpecCalculator struct {
	Regions []RegionInfo
}

// NewSpecCalculator는 새로운 사양 계산기를 생성합니다
func NewSpecCalculator(regions []RegionInfo) *SpecCalculator {
	return &SpecCalculator{
		Regions: regions,
	}
}

// 사용 가능한 값 중에서 요구사항과 가장 가까운 값 선택
func findClosestValue(availableValues []int, requiredValue int) int {
	// 오름차순 정렬
	sort.Ints(availableValues)
	
	closestValue := availableValues[0]
	for _, value := range availableValues {
		if value >= requiredValue {
			closestValue = value
			break
		}
	}
	return closestValue
}

// CalculateOptimalSpecs는 주어진 설정에 따라 최적의 사양을 계산합니다
func (sc *SpecCalculator) CalculateOptimalSpecs(config AggregatorConfig) (*SpecRecommendation, error) {
	// 유효성 검사
	if config.NumClients <= 0 {
		return nil, fmt.Errorf("number of clients must be positive")
	}
	if config.ModelSizeMB <= 0 {
		return nil, fmt.Errorf("model size must be positive")
	}
	if config.UpdateInterval <= 0 {
		return nil, fmt.Errorf("update interval must be positive")
	}
	
	// Aggregation 방식에 따른 복잡도 계수 설정
	complexityFactor, ok := ComplexityFactors[config.AggregationType]
	if !ok {
		// 기본값 사용
		complexityFactor = 1.0
	}
	
	// RAM 계산 (클라이언트 수 * 모델 크기 * 복잡도 계수)
	// 모델은 float32로 저장된다고 가정할 경우 메모리에 2배로 로드됨
	requiredRAMMB := float64(config.NumClients) * config.ModelSizeMB * complexityFactor * 2.0
	requiredRAMGB := int(math.Ceil(requiredRAMMB / 1024.0))
	
	// 최소 RAM 보장
	if requiredRAMGB < DefaultMinRAMGB {
		requiredRAMGB = DefaultMinRAMGB
	}
	
	// CPU 계산 (클라이언트 수, FLOP 및 업데이트 주기 기반)
	// 총 FLOP / 초당 처리 가능한 FLOP (예: 10GFLOP/s/코어)
	processingPerformance := DefaultGFLOPsPerCore * 1e9  // GFLOP/s per core
	totalFLOPs := float64(config.NumClients) * config.FLOPsPerClient * complexityFactor
	requiredSeconds := totalFLOPs / processingPerformance
	minimumCores := int(math.Ceil(requiredSeconds / float64(config.UpdateInterval)))
	
	// 최소 CPU 코어 보장
	if minimumCores < DefaultMinCPUCores {
		minimumCores = DefaultMinCPUCores
	}
	
	// Disk 계산 (클라이언트 수 * 모델 크기 * 업데이트 주기에 따른 저장 필요량)
	// 업데이트 주기를 일 단위로 변환 (최대 30일 저장으로 가정)
	daysToStore := math.Min(DefaultMaxStoreDays, float64(config.UpdateInterval*config.NumClients) / (24*3600))
	if daysToStore < 1 {
		daysToStore = 1
	}
	requiredDiskGB := int(math.Ceil(float64(config.NumClients) * config.ModelSizeMB * daysToStore / 1024.0))
	
	// 최소 디스크 크기 보장
	if requiredDiskGB < DefaultMinDiskGB {
		requiredDiskGB = DefaultMinDiskGB
	}
	
	// 적절한 region 선택
	var selectedRegion RegionInfo
	
	if len(sc.Regions) == 0 {
		return nil, fmt.Errorf("no regions available")
	}
	
	// 지역 우선순위에 따라 선택
	if config.RegionPriority == "latency" {
		// 지연시간이 가장 낮은 리전 선택
		minLatency := math.MaxFloat64
		for _, region := range sc.Regions {
			if region.LatencyMS < minLatency {
				minLatency = region.LatencyMS
				selectedRegion = region
			}
		}
	} else {
		// 비용이 가장 낮은 리전 선택 (CPU, RAM, Disk 비용 계산)
		minCost := math.MaxFloat64
		for _, region := range sc.Regions {
			// 해당 리전에서 가능한 사양 찾기
			cpuCores := findClosestValue(region.AvailableCPUs, minimumCores)
			ramGB := findClosestValue(region.AvailableRAMGB, requiredRAMGB)
			diskGB := findClosestValue(region.AvailableDiskGB, requiredDiskGB)
			
			// 비용 계산 (단순화된 공식)
			cost := float64(cpuCores) * region.CostFactorCPU * 30 +
				float64(ramGB) * region.CostFactorRAM * 30 +
				float64(diskGB) * region.CostFactorDisk * 30
			
			if cost < minCost {
				minCost = cost
				selectedRegion = region
			}
		}
	}
	
	// 최종 사양 조정 (선택된 리전의 가용 사양에 맞춤)
	cpuCores := findClosestValue(selectedRegion.AvailableCPUs, minimumCores)
	ramGB := findClosestValue(selectedRegion.AvailableRAMGB, requiredRAMGB)
	diskGB := findClosestValue(selectedRegion.AvailableDiskGB, requiredDiskGB)
	
	// 비용 계산
	estimatedCost := float64(cpuCores) * selectedRegion.CostFactorCPU * 30 +
		float64(ramGB) * selectedRegion.CostFactorRAM * 30 +
		float64(diskGB) * selectedRegion.CostFactorDisk * 30

	// 오토스케일링 설정
	scalingMin := DefaultMinInstances
	scalingMax := int(math.Ceil(float64(config.NumClients) * DefaultMaxScalingFactor))
	if scalingMax < DefaultMinMaxInstances {
		scalingMax = DefaultMinMaxInstances // 최소 2개 인스턴스 보장
	}
	
	// 결과 생성
	result := &SpecRecommendation{
		CPU:        cpuCores,
		RAMInGB:    ramGB,
		DiskInGB:   diskGB,
		Region:     selectedRegion.Name,
		Provider:   selectedRegion.Provider,
		TotalCost:  estimatedCost,
		LatencyMS:  selectedRegion.LatencyMS,
		ScalingMin: scalingMin,
		ScalingMax: scalingMax,
	}
	
	return result, nil
}