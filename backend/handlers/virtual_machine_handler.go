package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/gin-gonic/gin"
)

type VirtualMachineHandler struct {
	participantRepo    *repository.ParticipantRepository
	openStackService   *services.OpenStackService
	vmSelectionService *services.VMSelectionService
}

func NewVirtualMachineHandler(participantRepo *repository.ParticipantRepository) *VirtualMachineHandler {
	openStackService := services.NewOpenStackService("http://localhost:9090")
	vmSelectionService := services.NewVMSelectionService(openStackService)
	return &VirtualMachineHandler{
		participantRepo:    participantRepo,
		openStackService:   openStackService,
		vmSelectionService: vmSelectionService,
	}
}

// SelectOptimalVM - 깔끔하게 정리
func (h *VirtualMachineHandler) SelectOptimalVM(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	var criteria services.VMSelectionCriteria
	if err := c.ShouldBindJSON(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식", "details": err.Error()})
		return
	}

	// 실제 VM 조회 시도
	vmInstances, err := h.openStackService.GetAllVMInstances(participant)

	// 실제 VM이 없으면 Mock 데이터로 처리
	if err != nil || len(vmInstances) == 0 {
		mockVMs := generateMockVMInstances(participant.ID)
		criteriaMock := createMockCriteria(500)
		result, err := h.vmSelectionService.SelectOptimalVMFromMockData(participant, criteriaMock, mockVMs)
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Mock VM 선택 실패", "details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "최적의 VM이 선택되었습니다 (Mock 데이터)",
			"data":    result,
			"is_mock": true,
		})
		return
	}

	// 실제 VM 처리
	result, err := h.vmSelectionService.SelectOptimalVM(participant, criteria)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "VM 선택 실패", "details": err.Error(),
		})
		return
	}

	if result.SelectedVM == nil || result.SelectedVM.InstanceID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "조건을 만족하는 VM 없음",
			"reason": result.SelectionReason,
			"candidate_count": result.CandidateCount,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "최적의 VM이 선택되었습니다",
		"data":    result,
		"is_mock": false,
	})
}

func createMockCriteria(modelSizeMB int) services.VMSelectionCriteria {
    return services.VMSelectionCriteria{
        MinVCPUs:         1,
        MinRAM:           512,
        MinDisk:          5,
        RequiredStatus:   "ACTIVE",
        MaxCPUUsage:      70.0,
        MaxMemoryUsage:   80.0,
        ModelSizeMB:      modelSizeMB,
    }
}

func (h *VirtualMachineHandler) GetVMStats(c *gin.Context) {
	// 권한 확인 로직...
	participant, err := h.getParticipantWithAuth(c)
	if err != nil {
		return // 에러는 getParticipantWithAuth에서 처리
	}

	vmInstances, err := h.openStackService.GetAllVMInstances(participant)
	isMockData := false

	if err != nil || len(vmInstances) == 0 {
		vmInstances = generateMockVMInstances(participant.ID)
		isMockData = true
	}

	stats := calculateVMStats(vmInstances, isMockData)
	message := "실제 OpenStack VM 통계"
	if isMockData {
		message = "Mock 데이터 통계"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": message,
		"is_mock": isMockData,
	})
}

func (h *VirtualMachineHandler) GetVMRequests(c *gin.Context) {
	participant, err := h.getParticipantWithAuth(c)
	if err != nil {
		return
	}

	vmInstances, err := h.openStackService.GetAllVMInstances(participant)
	isMockData := false

	if err != nil || len(vmInstances) == 0 {
		vmInstances = generateMockVMInstances(participant.ID)
		isMockData = true
	}

	message := "VM 목록 조회 완료"
	if isMockData {
		message = "Mock 데이터 VM 목록"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    vmInstances,
		"count":   len(vmInstances),
		"is_mock": isMockData,
	})
}

// 나머지 메서드들...
func (h *VirtualMachineHandler) GetVMUtilizations(c *gin.Context) {
	participant, err := h.getParticipantWithAuth(c)
	if err != nil {
		return
	}

	utilizations, err := h.vmSelectionService.GetVMUtilizations(participant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "VM 사용률 조회 실패", "details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "VM 사용률 정보 조회 성공",
		"data":    utilizations,
		"count":   len(utilizations),
	})
}

func (h *VirtualMachineHandler) ResetVMSelectionRoundRobin(c *gin.Context) {
	participant, err := h.getParticipantWithAuth(c)
	if err != nil {
		return
	}

	h.vmSelectionService.ResetRoundRobinIndex(participant.ID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "라운드로빈 인덱스 초기화 완료",
	})
}

// 헬퍼 함수들
func (h *VirtualMachineHandler) getParticipantWithAuth(c *gin.Context) (*models.Participant, error) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return nil, err
	}
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return nil, fmt.Errorf("권한 없음")
	}
	return participant, nil
}

func calculateVMStats(vmInstances []services.VMInstance, isMockData bool) map[string]interface{} {
	stats := map[string]interface{}{
		"total": len(vmInstances), "active": 0, "available": 0, 
		"busy": 0, "error": 0, "building": 0, "shutoff": 0,
		"is_mock": isMockData,
	}

	for _, vm := range vmInstances {
		switch vm.Status {
		case "ACTIVE":
			stats["active"] = stats["active"].(int) + 1
			if isMockData && strings.Contains(vm.ID, "mock-vm-2") {
				stats["busy"] = stats["busy"].(int) + 1
			} else {
				stats["available"] = stats["available"].(int) + 1
			}
		case "ERROR":
			stats["error"] = stats["error"].(int) + 1
		case "BUILD", "BUILDING":
			stats["building"] = stats["building"].(int) + 1
		case "SHUTOFF":
			stats["shutoff"] = stats["shutoff"].(int) + 1
		}
	}
	return stats
}

func generateMockVMInstances(participantID string) []services.VMInstance {
	return []services.VMInstance{
		{
			ID: fmt.Sprintf("mock-vm-1-%s", participantID),
			Name: "저사양-테스트VM-1", Status: "ACTIVE", PowerState: 1,
			Flavor: services.FlavorDetails{ID: "flavor-small", Name: "small", VCPUs: 1, RAM: 2048, Disk: 20},
			Addresses: map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			}{"private": {{Addr: "192.168.1.10", Type: "fixed"}}},
			AvailabilityZone: "nova",
			Created: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Updated: time.Now().Format(time.RFC3339),
		},
		{
			ID: fmt.Sprintf("mock-vm-2-%s", participantID),
			Name: "중사양-테스트VM-2", Status: "ACTIVE", PowerState: 1,
			Flavor: services.FlavorDetails{ID: "flavor-medium", Name: "medium", VCPUs: 2, RAM: 4096, Disk: 50},
			Addresses: map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			}{"private": {{Addr: "192.168.1.11", Type: "fixed"}}},
			AvailabilityZone: "nova",
			Created: time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			Updated: time.Now().Format(time.RFC3339),
		},
		{
			ID: fmt.Sprintf("mock-vm-3-%s", participantID),
			Name: "고사양-테스트VM-3", Status: "ACTIVE", PowerState: 1,
			Flavor: services.FlavorDetails{ID: "flavor-large", Name: "large", VCPUs: 4, RAM: 8192, Disk: 100},
			Addresses: map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			}{"private": {{Addr: "192.168.1.12", Type: "fixed"}}},
			AvailabilityZone: "nova",
			Created: time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
			Updated: time.Now().Format(time.RFC3339),
		},
		{
			ID: fmt.Sprintf("mock-vm-4-%s", participantID),
			Name: "오프라인-테스트VM-4", Status: "SHUTOFF", PowerState: 4,
			Flavor: services.FlavorDetails{ID: "flavor-medium", Name: "medium", VCPUs: 2, RAM: 4096, Disk: 40},
			AvailabilityZone: "nova",
			Created: time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			Updated: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			ID: fmt.Sprintf("mock-vm-5-%s", participantID),
			Name: "에러-테스트VM-5", Status: "ERROR", PowerState: 0,
			Flavor: services.FlavorDetails{ID: "flavor-small", Name: "small", VCPUs: 1, RAM: 1024, Disk: 10},
			AvailabilityZone: "nova",
			Created: time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			Updated: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}
}