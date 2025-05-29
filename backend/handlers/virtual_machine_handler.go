package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
)

// VirtualMachineHandler는 가상머신 관련 API 핸들러입니다
type VirtualMachineHandler struct {
	vmRepo              *repository.VirtualMachineRepository
	participantRepo     *repository.ParticipantRepository
	openStackService    *services.OpenStackService
}

// NewVirtualMachineHandler는 새 VirtualMachineHandler 인스턴스를 생성합니다
func NewVirtualMachineHandler(vmRepo *repository.VirtualMachineRepository, participantRepo *repository.ParticipantRepository) *VirtualMachineHandler {
	return &VirtualMachineHandler{
		vmRepo:              vmRepo,
		participantRepo:     participantRepo,
		openStackService:    services.NewOpenStackService(),
	}
}

// CreateVirtualMachine은 새 가상머신을 생성하는 핸들러입니다
func (h *VirtualMachineHandler) CreateVirtualMachine(c *gin.Context) {
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

	var request struct {
		Name       string `json:"name" binding:"required"`
		InstanceID string `json:"instance_id" binding:"required"`
		FlavorID   string `json:"flavor_id"`
		ImageID    string `json:"image_id"`
		VCPUs      int    `json:"vcpus"`
		RAM        int    `json:"ram"`
		Disk       int    `json:"disk"`
		Metadata   string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	vm := &models.VirtualMachine{
		ID:            uuid.New().String(),
		ParticipantID: participantID,
		Name:          request.Name,
		InstanceID:    request.InstanceID,
		Status:        "BUILDING",
		FlavorID:      request.FlavorID,
		ImageID:       request.ImageID,
		VCPUs:         request.VCPUs,
		RAM:           request.RAM,
		Disk:          request.Disk,
		Metadata:      request.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.vmRepo.Create(vm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "가상머신 생성에 실패했습니다"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": vm})
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
	vm.LastHealthCheck = time.Now()
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

// CompleteVMTask는 VM의 작업을 완료 처리합니다
func (h *VirtualMachineHandler) CompleteVMTask(c *gin.Context) {
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

	// 작업 완료 처리
	if err := h.vmRepo.CompleteTask(vmID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "작업 완료 처리에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "작업이 완료되었습니다"})
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

	stats, err := h.vmRepo.GetVMStats(participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "통계 조회에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// DeleteVirtualMachine은 가상머신을 삭제합니다
func (h *VirtualMachineHandler) DeleteVirtualMachine(c *gin.Context) {
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

	// 작업 중인 VM은 삭제 불가
	if vm.IsBusy() {
		c.JSON(http.StatusConflict, gin.H{"error": "작업 중인 가상머신은 삭제할 수 없습니다"})
		return
	}

	if err := h.vmRepo.Delete(vmID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "가상머신 삭제에 실패했습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "가상머신이 삭제되었습니다"})
}

// VMPowerAction은 VM의 전원을 제어합니다
func (h *VirtualMachineHandler) VMPowerAction(c *gin.Context) {
	userID := utils.GetUserIDFromMiddleware(c)

	participantID := c.Param("id")
	vmID := c.Param("vmId")

	var request struct {
		Action string `json:"action" binding:"required,oneof=start stop reboot"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다. action은 start, stop, reboot 중 하나여야 합니다"})
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


	// VM 상태 업데이트
	switch request.Action {
	case "start":
		vm.Status = "BUILDING"
	case "stop":
		vm.Status = "SHUTOFF"
	case "reboot":
		vm.Status = "REBOOT"
	}
	vm.UpdatedAt = time.Now()
	h.vmRepo.Update(vm)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("VM %s 요청이 성공적으로 전송되었습니다", request.Action)})
	return
}

// HealthCheckVM은 VM의 헬스체크를 수행합니다
func (h *VirtualMachineHandler) HealthCheckVM(c *gin.Context) {
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

	// OpenStack VM 헬스체크
	
	healthResult, err := h.openStackService.HealthCheckSpecificVM(participant, vm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("VM 헬스체크에 실패했습니다: %v", err)})
		return
	}

	// VM 상태 업데이트
	if healthResult.Healthy {
		vm.Status = "ACTIVE"
	} else {
		vm.Status = "ERROR"
	}
	vm.LastHealthCheck = time.Now()
	vm.UpdatedAt = time.Now()
	h.vmRepo.Update(vm)

	c.JSON(http.StatusOK, gin.H{"data": healthResult})
	return
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
