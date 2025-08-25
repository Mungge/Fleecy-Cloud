package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// VirtualMachineHandler는 가상머신 관련 API 핸들러입니다
type VirtualMachineHandler struct {
	vmRepo             *repository.VirtualMachineRepository
	participantRepo    *repository.ParticipantRepository
	openStackService   *services.OpenStackService
	vmSelectionService *services.VMSelectionService
}

// NewVirtualMachineHandler는 새 VirtualMachineHandler 인스턴스를 생성합니다
func NewVirtualMachineHandler(vmRepo *repository.VirtualMachineRepository, participantRepo *repository.ParticipantRepository) *VirtualMachineHandler {
	// 기본 Prometheus URL은 더미값으로 설정 (실제로는 participant별로 동적 생성)
	openStackService := services.NewOpenStackService("http://localhost:9090")
	vmSelectionService := services.NewVMSelectionService(openStackService)
	return &VirtualMachineHandler{
		vmRepo:             vmRepo,
		participantRepo:    participantRepo,
		openStackService:   openStackService,
		vmSelectionService: vmSelectionService,
	}
}

// GetVirtualMachines는 특정 참여자의 모든 가상머신을 조회합니다
func (h *VirtualMachineHandler) GetVirtualMachines(c *gin.Context) {
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

	vms, err := h.vmRepo.GetByParticipantID(participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "가상머신 목록 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vms})
}

// GetVirtualMachine은 특정 가상머신을 조회합니다
func (h *VirtualMachineHandler) GetVirtualMachine(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")
	vmID := c.Param("vmId")

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

	vm, err := h.vmRepo.GetByID(vmID)
	if err != nil || vm == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "가상머신을 찾을 수 없습니다"})
		return
	}

	// VM이 해당 참여자에 속하는지 확인
	if vm.ParticipantID != participantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vm})
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

	// 선택된 VM이 없는 경우
	if result.SelectedVM == nil {
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

// MonitorVirtualMachine은 가상머신을 모니터링합니다
func (h *VirtualMachineHandler) MonitorVirtualMachine(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")
	vmID := c.Param("vmId")

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	vm, err := h.vmRepo.GetByID(vmID)
	if err != nil || vm == nil || vm.ParticipantID != participantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "가상머신을 찾을 수 없습니다"})
		return
	}

	// OpenStack VM 모니터링
	monitoringInfo, err := h.openStackService.MonitorSpecificVM(participant, vm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("VM 모니터링에 실패했습니다: %v", err)})
		return
	}

	// VM 상태 업데이트
	vm.UpdatedAt = time.Now()
	h.vmRepo.Update(vm)

	c.JSON(http.StatusOK, gin.H{"data": monitoringInfo})
	return
}

// AssignTaskToVM은 특정 VM에 작업을 할당합니다
func (h *VirtualMachineHandler) AssignTaskToVM(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")
	vmID := c.Param("vmId")

	var request struct {
		TaskID string `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "작업 ID가 필요합니다"})
		return
	}

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	vm, err := h.vmRepo.GetByID(vmID)
	if err != nil || vm == nil || vm.ParticipantID != participantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "가상머신을 찾을 수 없습니다"})
		return
	}

	// VM 사용 가능 여부 확인
	if !vm.IsAvailable() {
		c.JSON(http.StatusConflict, gin.H{"error": "VM이 사용 중이거나 사용할 수 없는 상태입니다"})
		return
	}

	// 작업 할당
	if err := h.vmRepo.AssignTask(vmID, request.TaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "작업 할당에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "작업이 성공적으로 할당되었습니다"})
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

// GetAvailableVMs는 사용 가능한 VM 목록을 조회합니다
func (h *VirtualMachineHandler) GetAvailableVMs(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	vms, err := h.vmRepo.GetAvailableVMs(participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용 가능한 VM 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vms})
}

// GetBusyVMs는 작업 중인 VM 목록을 조회합니다
func (h *VirtualMachineHandler) GetBusyVMs(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")

	// 권한 확인
	participant, err := h.participantRepo.GetByID(participantID)
	if err != nil || participant == nil || participant.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "권한이 없습니다"})
		return
	}

	vms, err := h.vmRepo.GetBusyVMs(participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "작업 중인 VM 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vms})
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
