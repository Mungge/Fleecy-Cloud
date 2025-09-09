package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
)

// VM 선택 기준
type VMSelectionCriteria struct {
	MinVCPUs         int
	MinRAM           int    // MB
	MinDisk          int    // GB
	RequiredStatus   string
	MaxCPUUsage      float64
	MaxMemoryUsage   float64
	ModelSizeMB      int    
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

// SelectOptimalVM은 실제 VM 데이터에서 최적의 VM을 선택합니다
func (s *VMSelectionService) SelectOptimalVM(participant *models.Participant, criteria VMSelectionCriteria) (*VMSelectionResult, error) {
    // OpenStack VM 목록을 VirtualMachine으로 변환
    openStackVMs, err := s.openStackService.GetAllVMInstances(participant)
    if err != nil {
        return nil, fmt.Errorf("VM 목록 조회 실패: %v", err)
    }

    var virtualMachines []VirtualMachine
    for _, osVM := range openStackVMs {
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
            IPAddresses:   "",
        }
        virtualMachines = append(virtualMachines, vm)
    }
    
    return s.selectOptimalVMCore(participant, criteria, virtualMachines, false)
}

// selectVMWithPriorityScore는 우선순위 점수 기반으로 VM을 선택합니다
func (s *VMSelectionService) selectVMWithPriorityScore(vmUtilizations []VMUtilization, modelSizeMB int) (*VMUtilization, string) {
	bestVM := &vmUtilizations[0]
	bestScore := s.calculatePriorityScore(&vmUtilizations[0])
	
	fmt.Printf("점수 계산:\n")
	for i := range vmUtilizations {
		score := s.calculatePriorityScore(&vmUtilizations[i])
		fmt.Printf("  %s: %.1f점\n", vmUtilizations[i].VM.Name, score)
		
		if score > bestScore {
			bestVM = &vmUtilizations[i]
			bestScore = score
		}
	}
	
	reason := fmt.Sprintf("우선순위 점수 기반 선택 (점수: %.1f, 모델크기: %dMB)", bestScore, modelSizeMB)
	return bestVM, reason
}

// calculatePriorityScore는 Python 시뮬레이션과 동일한 점수 계산 로직
func (s *VMSelectionService) calculatePriorityScore(vmUtil *VMUtilization) float64 {
	// 스펙 점수 (Python과 동일: CPU×10 + RAM(GB)×5 + Disk×2)
	specScore := float64(vmUtil.VM.VCPUs) * 10.0 + 
				float64(vmUtil.VM.RAM) / 1024.0 * 5.0 + 
				float64(vmUtil.VM.Disk) * 2.0
	
	// 사용률 페널티 (Python과 동일)
	usagePenalty := (vmUtil.MonitoringInfo.CPUUsage + 
					vmUtil.MonitoringInfo.MemoryUsage + 
					vmUtil.MonitoringInfo.DiskUsage) / 100.0 * 100.0
	
	return specScore - usagePenalty
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

func (s *VMSelectionService) SelectOptimalVMFromMockData(participant *models.Participant, criteria VMSelectionCriteria, mockVMs []VMInstance) (*VMSelectionResult, error) {
    // Mock VMInstance를 VirtualMachine으로 변환
    var virtualMachines []VirtualMachine
    for _, mockVM := range mockVMs {
        vm := VirtualMachine{
            InstanceID:    mockVM.ID,
            Name:          mockVM.Name,
            ParticipantID: participant.ID,
            Status:        mockVM.Status,
            FlavorID:      mockVM.Flavor.ID,
            FlavorName:    mockVM.Flavor.Name,
            VCPUs:         mockVM.Flavor.VCPUs,
            RAM:           mockVM.Flavor.RAM,
            Disk:          mockVM.Flavor.Disk,
            IPAddresses:   "",
        }
        virtualMachines = append(virtualMachines, vm)
    }
    
    // 기존 SelectOptimalVM 로직을 Mock 모니터링으로 실행
    return s.selectOptimalVMCore(participant, criteria, virtualMachines, true)
}

// selectOptimalVMCore는 VM 선택의 핵심 로직을 구현합니다
func (s *VMSelectionService) selectOptimalVMCore(participant *models.Participant, criteria VMSelectionCriteria, vms []VirtualMachine, isMock bool) (*VMSelectionResult, error) {
    // 기본값 설정
    if criteria.RequiredStatus == "" {
        criteria.RequiredStatus = "ACTIVE"
    }
    if criteria.MaxCPUUsage == 0 {
        criteria.MaxCPUUsage = 70.0
    }
    if criteria.ModelSizeMB == 0 {
        criteria.ModelSizeMB = 500
    }
    if criteria.MinVCPUs == 0 {
        criteria.MinVCPUs = 1
    }
    if criteria.MinRAM == 0 {
        criteria.MinRAM = 512
    }
    if criteria.MinDisk == 0 {
        criteria.MinDisk = 5
    }

    // dataType := "실제 데이터"
    // if isMock {
    //     dataType = "Mock 데이터"
    // }

    fmt.Printf("=== VM 선택 시작 ===\n")
    fmt.Printf("모델 크기: %dMB\n", criteria.ModelSizeMB)
    fmt.Printf("조건: vCPU>=%d, RAM>=%d, Disk>=%d, 상태=%s, MaxCPU=%.1f%%\n", 
        criteria.MinVCPUs, criteria.MinRAM, criteria.MinDisk, criteria.RequiredStatus, criteria.MaxCPUUsage)

    fmt.Printf("전체 VM 개수: %d\n", len(vms))

    // 1. 기본 필터링 및 모델 크기 기반 동적 체크
    var candidateVMs []VirtualMachine
    for i, vm := range vms {
        fmt.Printf("[%d/%d] VM 체크: %s (상태:%s, vCPU:%d, RAM:%dMB, Disk:%dGB)\n", 
            i+1, len(vms), vm.Name, vm.Status, vm.VCPUs, vm.RAM, vm.Disk)

        // 기본 조건 확인
        if vm.Status != criteria.RequiredStatus {
            fmt.Printf("  → 상태 불일치로 제외: %s != %s\n", vm.Status, criteria.RequiredStatus)
            continue
        }
        if vm.VCPUs < criteria.MinVCPUs {
            fmt.Printf("  → vCPU 부족으로 제외: %d < %d\n", vm.VCPUs, criteria.MinVCPUs)
            continue
        }
        if vm.RAM < criteria.MinRAM {
            fmt.Printf("  → RAM 부족으로 제외: %dMB < %dMB\n", vm.RAM, criteria.MinRAM)
            continue
        }
        if vm.Disk < criteria.MinDisk {
            fmt.Printf("  → Disk 부족으로 제외: %dGB < %dGB\n", vm.Disk, criteria.MinDisk)
            continue
        }

        // 모델 크기 기반 동적 리소스 체크
        requiredMemoryMB := int((float64(criteria.ModelSizeMB) * 2.0)) + 512
        requiredDiskGB := int((float64(criteria.ModelSizeMB) / 1024.0 * 3.0)) + 1
        
        if vm.RAM < requiredMemoryMB {
            fmt.Printf("  → 모델 크기 대비 RAM 부족으로 제외: %dMB < %dMB (모델:%dMB)\n", 
                vm.RAM, requiredMemoryMB, criteria.ModelSizeMB)
            continue
        }
        if vm.Disk < requiredDiskGB {
            fmt.Printf("  → 모델 크기 대비 Disk 부족으로 제외: %dGB < %dGB (모델:%dMB)\n", 
                vm.Disk, requiredDiskGB, criteria.ModelSizeMB)
            continue
        }

        candidateVMs = append(candidateVMs, vm)
        fmt.Printf("  → 모든 조건 통과\n")
    }

    if len(candidateVMs) == 0 {
        fmt.Printf("기본 필터링 후 후보: 0개 - 조건을 만족하는 VM이 없음\n")
        return &VMSelectionResult{
            SelectedVM:      nil,
            SelectionReason: fmt.Sprintf("조건을 만족하는 VM을 찾을 수 없습니다."),
            CandidateCount:  0,
        }, nil
    }

    fmt.Printf("기본 필터링 후 후보: %d개\n", len(candidateVMs))

    // 2. 사용률 정보 수집 및 필터링
    var vmUtilizations []VMUtilization
    for i, vm := range candidateVMs {
        fmt.Printf("[%d/%d] 사용률 체크: %s\n", i+1, len(candidateVMs), vm.Name)
        
        var utilization *VMUtilization
        var err error
        
        if isMock {
            // Mock 데이터 사용
            utilization = s.createMockUtilization(vm)
        } else {
            // 실제 모니터링 데이터 사용
            utilization, err = s.getVMUtilization(participant, &vm)
            if err != nil {
                fmt.Printf("  → 사용률 조회 실패로 제외: %v\n", err)
                continue
            }
        }

        fmt.Printf("  → CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%, 건강상태: %t\n", 
            utilization.MonitoringInfo.CPUUsage, 
            utilization.MonitoringInfo.MemoryUsage,
            utilization.MonitoringInfo.DiskUsage,
            utilization.IsHealthy)

        // CPU 사용률 조건 확인
        if utilization.MonitoringInfo.CPUUsage > criteria.MaxCPUUsage {
            fmt.Printf("  → CPU 사용률 초과로 제외: %.1f%% > %.1f%%\n", 
                utilization.MonitoringInfo.CPUUsage, criteria.MaxCPUUsage)
            continue
        }
        
        // 메모리 사용률 동적 계산
        availableMemoryMB := float64(vm.RAM) * (1.0 - utilization.MonitoringInfo.MemoryUsage/100.0)
        requiredMemoryMB := float64(criteria.ModelSizeMB) * 2.0 + 512.0
        
        if availableMemoryMB < requiredMemoryMB {
            fmt.Printf("  → 사용 가능한 메모리 부족으로 제외: %.1fMB < %.1fMB\n", 
                availableMemoryMB, requiredMemoryMB)
            continue
        }

        // 디스크 사용률 동적 계산
        availableDiskGB := float64(vm.Disk) * (1.0 - utilization.MonitoringInfo.DiskUsage/100.0)
        requiredDiskGB := float64(criteria.ModelSizeMB) / 1024.0 * 3.0 + 1.0
        
        if availableDiskGB < requiredDiskGB {
            fmt.Printf("  → 사용 가능한 디스크 부족으로 제외: %.1fGB < %.1fGB\n", 
                availableDiskGB, requiredDiskGB)
            continue
        }

        vmUtilizations = append(vmUtilizations, *utilization)
        fmt.Printf("  → 모든 사용률 조건 통과\n")
    }

    if len(vmUtilizations) == 0 {
        fmt.Printf("사용률 필터링 후 후보: 0개 - 사용률 조건을 만족하는 VM이 없음\n")
        return &VMSelectionResult{
            SelectedVM:      nil,
            SelectionReason: fmt.Sprintf("사용률 조건을 만족하는 VM을 찾을 수 없습니다"),
            CandidateCount:  len(candidateVMs),
        }, nil
    }

    fmt.Printf("사용률 필터링 후 후보: %d개\n", len(vmUtilizations))

    // 3. 우선순위 점수 기반 선택
    selectedUtilization, reason := s.selectVMWithPriorityScore(vmUtilizations, criteria.ModelSizeMB)

    finalReason := reason
    if isMock {
        finalReason = reason
    }

    fmt.Printf("최종 선택: %s\n", selectedUtilization.VM.Name)
    fmt.Printf("선택 이유: %s\n", finalReason)
    fmt.Printf("=== VM 선택 완료 ===\n\n")

    return &VMSelectionResult{
        SelectedVM:      &selectedUtilization.VM,
        SelectionReason: finalReason,
        CandidateCount:  len(vmUtilizations),
    }, nil
}

// createMockUtilization은 Mock VM의 사용률 정보를 생성합니다
func (s *VMSelectionService) createMockUtilization(vm VirtualMachine) *VMUtilization {
    var cpuUsage, memoryUsage, diskUsage float64
    
    switch {
    case strings.Contains(vm.InstanceID, "mock-vm-1"):
        cpuUsage, memoryUsage, diskUsage = 25.5, 40.2, 15.8
    case strings.Contains(vm.InstanceID, "mock-vm-2"):
        cpuUsage, memoryUsage, diskUsage = 75.8, 60.3, 30.5
    case strings.Contains(vm.InstanceID, "mock-vm-3"):
        cpuUsage, memoryUsage, diskUsage = 35.2, 25.1, 12.7
    case strings.Contains(vm.InstanceID, "mock-vm-4"):
        cpuUsage, memoryUsage, diskUsage = 0, 0, 0
    case strings.Contains(vm.InstanceID, "mock-vm-5"):
        cpuUsage, memoryUsage, diskUsage = 0, 0, 0
    default:
        cpuUsage, memoryUsage, diskUsage = 50.0, 50.0, 50.0
    }
    
    monitoringInfo := VMMonitoringInfo{
        InstanceID:       vm.InstanceID,
        CPUUsage:         cpuUsage,
        MemoryUsage:      memoryUsage,
        DiskUsage:        diskUsage,
        NetworkInBytes:   1024 * 1024,
        NetworkOutBytes:  512 * 1024,
        LastUpdated:      time.Now(),
    }

    return &VMUtilization{
        VM:               vm,
        MonitoringInfo:   monitoringInfo,
        IsHealthy:        cpuUsage < 80 && memoryUsage < 80,
        UtilizationScore: (cpuUsage * 0.6) + (memoryUsage * 0.3) + (diskUsage * 0.1),
    }
}