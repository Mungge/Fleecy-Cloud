package services

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
)

// VM 선택 기준
type VMSelectionCriteria struct {
	MinVCPUs       int     `json:"min_vcpus"`
	MinRAM         int     `json:"min_ram"`          // MB 단위
	MinDisk        int     `json:"min_disk"`         // GB 단위
	MaxCPUUsage    float64 `json:"max_cpu_usage"`    // 최대 CPU 사용률 (%)
	MaxMemoryUsage float64 `json:"max_memory_usage"` // 최대 메모리 사용률 (%)
	RequiredStatus string  `json:"required_status"`  // 필요한 VM 상태 (기본: "ACTIVE")
}

// VM 선택 결과
type VMSelectionResult struct {
	SelectedVM      *VirtualMachine `json:"selected_vm"`
	SelectionReason string          `json:"selection_reason"`
	CandidateCount  int             `json:"candidate_count"`
}

// VM 사용률 정보 (선택 알고리즘용) - 기존 모델을 활용
type VMUtilization struct {
	VM               VirtualMachine   `json:"vm"`
	MonitoringInfo   VMMonitoringInfo `json:"monitoring_info"`
	RuntimeInfo      VMRuntimeInfo    `json:"runtime_info"`
	UtilizationScore float64          `json:"utilization_score"` // 종합 사용률 점수
	IsHealthy        bool             `json:"is_healthy"`
}

type VMSelectionService struct {
	openStackService  *OpenStackService
	roundRobinMutex   sync.Mutex
	lastSelectedIndex map[string]int // ParticipantID별 마지막 선택된 인덱스
}

func NewVMSelectionService(openStackService *OpenStackService) *VMSelectionService {
	return &VMSelectionService{
		openStackService:  openStackService,
		lastSelectedIndex: make(map[string]int),
	}
}

// SelectOptimalVM은 사용률과 라운드로빈을 고려하여 최적의 VM을 선택합니다
func (s *VMSelectionService) SelectOptimalVM(participant *models.Participant, criteria VMSelectionCriteria) (*VMSelectionResult, error) {
	// 기본값 설정
	if criteria.RequiredStatus == "" {
		criteria.RequiredStatus = "ACTIVE"
	}
	if criteria.MaxCPUUsage == 0 {
		criteria.MaxCPUUsage = 80.0 // 기본 80%
	}
	if criteria.MaxMemoryUsage == 0 {
		criteria.MaxMemoryUsage = 80.0 // 기본 80%
	}

	fmt.Printf("=== VM 선택 시작 ===\n")
	fmt.Printf("조건: vCPU>=%d, RAM>=%d, Disk>=%d, 상태=%s, MaxCPU=%.1f%%, MaxMemory=%.1f%%\n", 
		criteria.MinVCPUs, criteria.MinRAM, criteria.MinDisk, criteria.RequiredStatus, criteria.MaxCPUUsage, criteria.MaxMemoryUsage)

	openStackVMs, err := s.openStackService.GetAllVMInstances(participant)
	if err != nil {
		return nil, fmt.Errorf("VM 목록 조회 실패: %v", err)
	}

	fmt.Printf("전체 VM 개수: %d\n", len(openStackVMs))

	// 2. DB 형태로 변환 및 기본 필터링
	var candidateVMs []VirtualMachine
	for i, osVM := range openStackVMs {
		fmt.Printf("[%d/%d] VM 체크: %s (상태:%s, vCPU:%d, RAM:%dMB, Disk:%dGB)\n", 
			i+1, len(openStackVMs), osVM.Name, osVM.Status, osVM.Flavor.VCPUs, osVM.Flavor.RAM, osVM.Flavor.Disk)

		// 기본 조건 확인
		if osVM.Status != criteria.RequiredStatus {
			fmt.Printf("  → 상태 불일치로 제외: %s != %s\n", osVM.Status, criteria.RequiredStatus)
			continue
		}
		if osVM.Flavor.VCPUs < criteria.MinVCPUs {
			fmt.Printf("  → vCPU 부족으로 제외: %d < %d\n", osVM.Flavor.VCPUs, criteria.MinVCPUs)
			continue
		}
		if osVM.Flavor.RAM < criteria.MinRAM {
			fmt.Printf("  → RAM 부족으로 제외: %dMB < %dMB\n", osVM.Flavor.RAM, criteria.MinRAM)
			continue
		}
		if osVM.Flavor.Disk < criteria.MinDisk {
			fmt.Printf("  → Disk 부족으로 제외: %dGB < %dGB\n", osVM.Flavor.Disk, criteria.MinDisk)
			continue
		}

		// IP 주소 직렬화
		ipAddressesJSON, _ := json.Marshal(osVM.Addresses)

		vm := VirtualMachine{
			InstanceID:    osVM.ID,
			Name:          osVM.Name,
			ParticipantID: participant.ID,
			Status:        osVM.Status,
			FlavorID:      osVM.Flavor.ID,
			FlavorName:    osVM.Flavor.Name,
			VCPUs:         osVM.Flavor.VCPUs,
			RAM:           osVM.Flavor.RAM,
			Disk:          osVM.Flavor.Disk,
			IPAddresses:   string(ipAddressesJSON),
		}

		candidateVMs = append(candidateVMs, vm)
		fmt.Printf("  → 기본 조건 통과\n")
	}

	if len(candidateVMs) == 0 {
		fmt.Printf("기본 필터링 후 후보: 0개 - 조건을 만족하는 VM이 없음\n")
		return &VMSelectionResult{
			SelectedVM:      nil,
			SelectionReason: "조건을 만족하는 VM을 찾을 수 없습니다",
			CandidateCount:  0,
		}, nil
	}

	fmt.Printf("기본 필터링 후 후보: %d개\n", len(candidateVMs))

	// 3. 각 VM의 사용률 정보 수집
	var vmUtilizations []VMUtilization
	for i, vm := range candidateVMs {
		fmt.Printf("[%d/%d] 사용률 체크: %s\n", i+1, len(candidateVMs), vm.Name)
		
		utilization, err := s.getVMUtilization(participant, &vm)
		if err != nil {
			fmt.Printf("  → 사용률 조회 실패로 제외: %v\n", err)
			continue
		}

		fmt.Printf("  → CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%, 종합점수: %.1f, 건강상태: %t\n", 
			utilization.MonitoringInfo.CPUUsage, 
			utilization.MonitoringInfo.MemoryUsage,
			utilization.MonitoringInfo.DiskUsage,
			utilization.UtilizationScore,
			utilization.IsHealthy)

		// 사용률 조건 확인
		if utilization.MonitoringInfo.CPUUsage > criteria.MaxCPUUsage {
			fmt.Printf("  → CPU 사용률 초과로 제외: %.1f%% > %.1f%%\n", 
				utilization.MonitoringInfo.CPUUsage, criteria.MaxCPUUsage)
			continue
		}
		
		if utilization.MonitoringInfo.MemoryUsage > criteria.MaxMemoryUsage {
			fmt.Printf("  → Memory 사용률 초과로 제외: %.1f%% > %.1f%%\n", 
				utilization.MonitoringInfo.MemoryUsage, criteria.MaxMemoryUsage)
			continue
		}

		vmUtilizations = append(vmUtilizations, *utilization)
		fmt.Printf("  → 사용률 조건 통과\n")
	}

	if len(vmUtilizations) == 0 {
		fmt.Printf("사용률 필터링 후 후보: 0개 - 사용률 조건을 만족하는 VM이 없음\n")
		return &VMSelectionResult{
			SelectedVM:      nil,
			SelectionReason: "사용률 조건을 만족하는 VM을 찾을 수 없습니다",
			CandidateCount:  len(candidateVMs),
		}, nil
	}

	fmt.Printf("사용률 필터링 후 후보: %d개\n", len(vmUtilizations))

	// 최종 후보 VM들 요약 출력
	fmt.Printf("최종 후보 VM들:\n")
	for i, util := range vmUtilizations {
		fmt.Printf("  [%d] %s: CPU=%.1f%%, Memory=%.1f%%, Score=%.1f, Healthy=%t\n", 
			i+1, util.VM.Name, util.MonitoringInfo.CPUUsage, 
			util.MonitoringInfo.MemoryUsage, util.UtilizationScore, util.IsHealthy)
	}

	// 4. VM 선택 (사용률 + 라운드로빈)
	selectedVM, reason := s.selectVMWithUtilizationAndRoundRobin(participant.ID, vmUtilizations)

	fmt.Printf("최종 선택: %s\n", selectedVM.VM.Name)
	fmt.Printf("선택 이유: %s\n", reason)
	fmt.Printf("=== VM 선택 완료 ===\n\n")

	return &VMSelectionResult{
		SelectedVM:      &selectedVM.VM,
		SelectionReason: reason,
		CandidateCount:  len(vmUtilizations),
	}, nil
}

// getVMUtilization은 VM의 현재 사용률 정보를 조회합니다
func (s *VMSelectionService) getVMUtilization(participant *models.Participant, vm *VirtualMachine) (*VMUtilization, error) {
	// 모니터링 정보 조회 - participant의 OpenStack endpoint 사용
	monitoringInfo, err := s.openStackService.GetVMMonitoringInfoWithParticipant(participant, vm.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("VM 모니터링 정보 조회 실패: %v", err)
	}

	// 실시간 상태 정보 조회
	runtimeInfo, err := s.openStackService.GetVMRuntimeStatus(participant, vm.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("VM 런타임 상태 조회 실패: %v", err)
	}

	// 헬스체크 수행
	healthCheck, err := s.openStackService.HealthCheckSpecificVM(participant, vm)
	if err != nil {
		return nil, fmt.Errorf("VM 헬스체크 실패: %v", err)
	}

	// 종합 사용률 점수 계산 (CPU 60%, Memory 30%, Disk 10%)
	utilizationScore := (monitoringInfo.CPUUsage * 0.6) +
		(monitoringInfo.MemoryUsage * 0.3) +
		(monitoringInfo.DiskUsage * 0.1)

	return &VMUtilization{
		VM:               *vm,
		MonitoringInfo:   *monitoringInfo,
		RuntimeInfo:      *runtimeInfo,
		UtilizationScore: utilizationScore,
		IsHealthy:        healthCheck.Healthy,
	}, nil
}

// selectVMWithUtilizationAndRoundRobin은 사용률과 라운드로빈을 조합하여 VM을 선택합니다
func (s *VMSelectionService) selectVMWithUtilizationAndRoundRobin(participantID string, vmUtilizations []VMUtilization) (*VMUtilization, string) {
	// 건강한 VM만 필터링
	var healthyVMs []VMUtilization
	for _, vm := range vmUtilizations {
		if vm.IsHealthy {
			healthyVMs = append(healthyVMs, vm)
		}
	}

	if len(healthyVMs) == 0 {
		// 건강한 VM이 없으면 가장 사용률이 낮은 VM 선택
		sort.Slice(vmUtilizations, func(i, j int) bool {
			return vmUtilizations[i].UtilizationScore < vmUtilizations[j].UtilizationScore
		})
		return &vmUtilizations[0], "건강한 VM이 없어 사용률이 가장 낮은 VM 선택"
	}

	// 사용률 기준으로 정렬 (낮은 순서)
	sort.Slice(healthyVMs, func(i, j int) bool {
		return healthyVMs[i].UtilizationScore < healthyVMs[j].UtilizationScore
	})

	// 사용률이 비슷한 VM들 중에서 라운드로빈 선택
	// 사용률 차이가 10% 이내인 VM들을 동일 그룹으로 간주
	const utilizationThreshold = 10.0
	lowestUtilization := healthyVMs[0].UtilizationScore

	var similarUtilizationVMs []VMUtilization
	for _, vm := range healthyVMs {
		if vm.UtilizationScore-lowestUtilization <= utilizationThreshold {
			similarUtilizationVMs = append(similarUtilizationVMs, vm)
		} else {
			break
		}
	}

	// 라운드로빈 선택
	s.roundRobinMutex.Lock()
	defer s.roundRobinMutex.Unlock()

	lastIndex, exists := s.lastSelectedIndex[participantID]
	if !exists {
		lastIndex = -1
	}

	nextIndex := (lastIndex + 1) % len(similarUtilizationVMs)
	s.lastSelectedIndex[participantID] = nextIndex

	selectedVM := &similarUtilizationVMs[nextIndex]

	reason := fmt.Sprintf("사용률 %.1f%% (CPU: %.1f%%, Memory: %.1f%%) - 라운드로빈으로 선택",
		selectedVM.UtilizationScore,
		selectedVM.MonitoringInfo.CPUUsage,
		selectedVM.MonitoringInfo.MemoryUsage)

	return selectedVM, reason
}

// GetVMUtilizations은 참가자의 모든 VM 사용률 정보를 반환합니다 (모니터링용)
func (s *VMSelectionService) GetVMUtilizations(participant *models.Participant) ([]VMUtilization, error) {
	// OpenStack에서 직접 VM 목록 조회 (GetAllVMInstances 사용)
	openStackVMs, err := s.openStackService.GetAllVMInstances(participant)
	if err != nil {
		return nil, fmt.Errorf("VM 목록 조회 실패: %v", err)
	}

	var utilizations []VMUtilization
	for _, osVM := range openStackVMs {
		// IP 주소 직렬화
		ipAddressesJSON, _ := json.Marshal(osVM.Addresses)

		vm := VirtualMachine{
			InstanceID:    osVM.ID,
			Name:          osVM.Name,
			ParticipantID: participant.ID,
			Status:        osVM.Status,
			FlavorID:      osVM.Flavor.ID,
			FlavorName:    osVM.Flavor.Name,
			VCPUs:         osVM.Flavor.VCPUs,
			RAM:           osVM.Flavor.RAM,
			Disk:          osVM.Flavor.Disk,
			IPAddresses:   string(ipAddressesJSON),
		}

		utilization, err := s.getVMUtilization(participant, &vm)
		if err != nil {
			// 에러가 발생한 경우 기본값으로 설정
			utilization = &VMUtilization{
				VM: vm,
				MonitoringInfo: VMMonitoringInfo{
					InstanceID:      vm.InstanceID,
					CPUUsage:        0,
					MemoryUsage:     0,
					DiskUsage:       0,
					NetworkInBytes:  0,
					NetworkOutBytes: 0,
					LastUpdated:     time.Now(),
				},
				RuntimeInfo: VMRuntimeInfo{
					InstanceID:  vm.InstanceID,
					Status:      vm.Status,
					PowerState:  0,
					LastChecked: time.Now(),
				},
				UtilizationScore: 0,
				IsHealthy:        false,
			}
		}

		utilizations = append(utilizations, *utilization)
	}

	return utilizations, nil
}

// ResetRoundRobinIndex는 특정 참가자의 라운드로빈 인덱스를 초기화합니다
func (s *VMSelectionService) ResetRoundRobinIndex(participantID string) {
	s.roundRobinMutex.Lock()
	defer s.roundRobinMutex.Unlock()

	delete(s.lastSelectedIndex, participantID)
}
