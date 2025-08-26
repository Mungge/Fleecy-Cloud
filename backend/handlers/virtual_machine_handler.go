package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// VirtualMachineHandler는 가상머신 관련 API 핸들러입니다
type VirtualMachineHandler struct {
	participantRepo    *repository.ParticipantRepository
	openStackService   *services.OpenStackService
	vmSelectionService *services.VMSelectionService
}

// NewVirtualMachineHandler는 새 VirtualMachineHandler 인스턴스를 생성합니다
func NewVirtualMachineHandler(participantRepo *repository.ParticipantRepository) *VirtualMachineHandler {
	// 기본 Prometheus URL은 더미값으로 설정 (실제로는 participant별로 동적 생성)
	openStackService := services.NewOpenStackService("http://localhost:9090")
	vmSelectionService := services.NewVMSelectionService(openStackService)
	return &VirtualMachineHandler{
		participantRepo:    participantRepo,
		openStackService:   openStackService,
		vmSelectionService: vmSelectionService,
	}
}

// SelectOptimalVM은 연합학습에 최적의 VM을 선택합니다 (새로 추가)
func (h *VirtualMachineHandler) SelectOptimalVM(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	// 참여자 소유자 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	// 요청 본문에서 선택 기준 파싱
	var criteria services.VMSelectionCriteria
	if err := c.ShouldBindJSON(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식", "details": err.Error()})
		return
	}

	// VM 선택 실행
	result, err := h.vmSelectionService.SelectOptimalVM(participant, criteria)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "VM 선택 중 오류 발생",
			"details": err.Error(),
		})
		return
	}

	// 선택된 VM이 없는 경우 (빈 구조체인지 확인)
	if result.SelectedVM.InstanceID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":           "조건을 만족하는 VM을 찾을 수 없습니다",
			"reason":          result.SelectionReason,
			"candidate_count": result.CandidateCount,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "최적의 VM이 선택되었습니다",
		"data":    result,
	})
}

// GetVMUtilizations는 참가자의 모든 VM 사용률 정보를 조회합니다
func (h *VirtualMachineHandler) GetVMUtilizations(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	// 참여자 소유자 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	// VM 사용률 정보 조회
	utilizations, err := h.vmSelectionService.GetVMUtilizations(participant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "VM 사용률 조회 중 오류 발생",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "VM 사용률 정보를 성공적으로 조회했습니다",
		"data":    utilizations,
		"count":   len(utilizations),
	})
}

// ResetVMSelectionRoundRobin은 VM 선택 라운드로빈을 초기화합니다 (새로 추가)
func (h *VirtualMachineHandler) ResetVMSelectionRoundRobin(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	// 참여자 소유자 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "참여자를 찾을 수 없습니다"})
		return
	}
	if participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	h.vmSelectionService.ResetRoundRobinIndex(participantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "라운드로빈 인덱스가 초기화되었습니다",
	})
}

// GetVMStats는 참여자의 VM 통계를 조회합니다
func (h *VirtualMachineHandler) GetVMStats(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	// OpenStack에서 실시간 VM 목록을 가져와서 통계 계산
	vmInstances, err := h.openStackService.GetAllVMInstances(participant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("VM 목록 조회 실패: %v", err)})
		return
	}

	// 실시간 통계 계산
	stats := map[string]interface{}{
		"total":     len(vmInstances),
		"active":    0,
		"available": 0,
		"busy":      0,
		"error":     0,
		"building":  0,
		"shutoff":   0,
	}

	for _, vm := range vmInstances {
		switch vm.Status {
		case "ACTIVE":
			stats["active"] = stats["active"].(int) + 1
			// ACTIVE 상태인 VM은 기본적으로 사용 가능으로 간주
			// 실제로는 현재 작업 중인지 확인해야 하지만, 여기서는 단순화
			stats["available"] = stats["available"].(int) + 1
		case "ERROR":
			stats["error"] = stats["error"].(int) + 1
		case "BUILD", "BUILDING":
			stats["building"] = stats["building"].(int) + 1
		case "SHUTOFF":
			stats["shutoff"] = stats["shutoff"].(int) + 1
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetOpenStackVMs는 OpenStack에서 직접 VM 목록을 조회합니다
func (h *VirtualMachineHandler) GetVMRequests(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)
	participantID := c.Param("id")

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	// OpenStack에서 직접 VM 목록 조회
	vmInstances, err := h.openStackService.GetAllVMInstances(participant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("OpenStack VM 목록 조회 실패: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "VM 목록 조회가 완료되었습니다",
		"data":    vmInstances,
		"count":   len(vmInstances),
	})
}
