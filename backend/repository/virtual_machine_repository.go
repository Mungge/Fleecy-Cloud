package repository

import (
	"github.com/Mungge/Fleecy-Cloud/models"
	"gorm.io/gorm"
)

type VirtualMachineRepository struct {
	db *gorm.DB
}

func NewVirtualMachineRepository(db *gorm.DB) *VirtualMachineRepository {
	return &VirtualMachineRepository{db: db}
}

// Create은 새로운 가상머신을 생성합니다
func (r *VirtualMachineRepository) Create(vm *models.VirtualMachine) error {
	return r.db.Create(vm).Error
}

// GetByID는 ID로 가상머신을 조회합니다
func (r *VirtualMachineRepository) GetByID(id string) (*models.VirtualMachine, error) {
	var vm models.VirtualMachine
	err := r.db.Where("id = ?", id).First(&vm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vm, nil
}

// GetByParticipantID는 특정 참여자의 모든 가상머신을 조회합니다
func (r *VirtualMachineRepository) GetByParticipantID(participantID string) ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Where("participant_id = ?", participantID).Find(&vms).Error
	return vms, err
}

// GetByInstanceID는 클라우드 인스턴스 ID로 가상머신을 조회합니다
func (r *VirtualMachineRepository) GetByInstanceID(instanceID string) (*models.VirtualMachine, error) {
	var vm models.VirtualMachine
	err := r.db.Where("instance_id = ?", instanceID).First(&vm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vm, nil
}

// GetAvailableVMs는 사용 가능한 가상머신 목록을 조회합니다
func (r *VirtualMachineRepository) GetAvailableVMs(participantID string) ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Where("participant_id = ? AND status = ? AND (current_task_id = ? OR current_task_id IS NULL)", 
		participantID, "ACTIVE", "").Find(&vms).Error
	return vms, err
}

// GetBusyVMs는 작업 중인 가상머신 목록을 조회합니다
func (r *VirtualMachineRepository) GetBusyVMs(participantID string) ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Where("participant_id = ? AND current_task_id != ? AND current_task_id IS NOT NULL", 
		participantID, "").Find(&vms).Error
	return vms, err
}

// Update는 가상머신 정보를 업데이트합니다
func (r *VirtualMachineRepository) Update(vm *models.VirtualMachine) error {
	return r.db.Save(vm).Error
}

// UpdateStatus는 가상머신 상태를 업데이트합니다
func (r *VirtualMachineRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.VirtualMachine{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateMetrics는 가상머신 메트릭을 업데이트합니다
func (r *VirtualMachineRepository) UpdateMetrics(id string, metrics map[string]interface{}) error {
	return r.db.Model(&models.VirtualMachine{}).Where("id = ?", id).Updates(metrics).Error
}

// AssignTask는 가상머신에 작업을 할당합니다
func (r *VirtualMachineRepository) AssignTask(id string, taskID string) error {
	updates := map[string]interface{}{
		"current_task_id":  taskID,
		"task_assigned_at": "NOW()",
	}
	return r.db.Model(&models.VirtualMachine{}).Where("id = ?", id).Updates(updates).Error
}

// CompleteTask는 가상머신의 작업을 완료 처리합니다
func (r *VirtualMachineRepository) CompleteTask(id string) error {
	updates := map[string]interface{}{
		"current_task_id":         "",
		"last_task_completed_at":  "NOW()",
		"total_tasks_completed":   gorm.Expr("total_tasks_completed + 1"),
	}
	return r.db.Model(&models.VirtualMachine{}).Where("id = ?", id).Updates(updates).Error
}

// Delete는 가상머신을 삭제합니다
func (r *VirtualMachineRepository) Delete(id string) error {
	return r.db.Delete(&models.VirtualMachine{}, "id = ?", id).Error
}

// GetVMStats는 가상머신 통계를 조회합니다
func (r *VirtualMachineRepository) GetVMStats(participantID string) (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `json:"total"`
		Active    int64 `json:"active"`
		Busy      int64 `json:"busy"`
		Available int64 `json:"available"`
		Error     int64 `json:"error"`
	}

	// 전체 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("participant_id = ?", participantID).Count(&stats.Total)
	
	// 활성 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("participant_id = ? AND status = ?", participantID, "ACTIVE").Count(&stats.Active)
	
	// 작업 중인 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("participant_id = ? AND current_task_id != ? AND current_task_id IS NOT NULL", participantID, "").Count(&stats.Busy)
	
	// 사용 가능한 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("participant_id = ? AND status = ? AND (current_task_id = ? OR current_task_id IS NULL)", participantID, "ACTIVE", "").Count(&stats.Available)
	
	// 오류 상태 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("participant_id = ? AND status = ?", participantID, "ERROR").Count(&stats.Error)

	return map[string]interface{}{
		"total":     stats.Total,
		"active":    stats.Active,
		"busy":      stats.Busy,
		"available": stats.Available,
		"error":     stats.Error,
	}, nil
}

// GetGlobalVMStats는 시스템 전체 가상머신 통계를 조회합니다
func (r *VirtualMachineRepository) GetGlobalVMStats() (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `json:"total"`
		Active    int64 `json:"active"`
		Busy      int64 `json:"busy"`
		Available int64 `json:"available"`
		Error     int64 `json:"error"`
		Building  int64 `json:"building"`
	}

	// 전체 VM 수
	r.db.Model(&models.VirtualMachine{}).Count(&stats.Total)
	
	// 활성 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("status = ?", "ACTIVE").Count(&stats.Active)
	
	// 작업 중인 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("current_task_id != ? AND current_task_id IS NOT NULL", "").Count(&stats.Busy)
	
	// 사용 가능한 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("status = ? AND (current_task_id = ? OR current_task_id IS NULL)", "ACTIVE", "").Count(&stats.Available)
	
	// 오류 상태 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("status = ?", "ERROR").Count(&stats.Error)
	
	// 빌딩 상태 VM 수
	r.db.Model(&models.VirtualMachine{}).Where("status = ?", "BUILDING").Count(&stats.Building)

	return map[string]interface{}{
		"total":     stats.Total,
		"active":    stats.Active,
		"busy":      stats.Busy,
		"available": stats.Available,
		"error":     stats.Error,
		"building":  stats.Building,
	}, nil
}

// GetGlobalAvailableVMs는 시스템 전체에서 사용 가능한 가상머신 목록을 조회합니다
func (r *VirtualMachineRepository) GetGlobalAvailableVMs() ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Preload("Participant").Where("status = ? AND (current_task_id = ? OR current_task_id IS NULL)", 
		"ACTIVE", "").Find(&vms).Error
	return vms, err
}

// GetGlobalBusyVMs는 시스템 전체에서 작업 중인 가상머신 목록을 조회합니다
func (r *VirtualMachineRepository) GetGlobalBusyVMs() ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Preload("Participant").Where("current_task_id != ? AND current_task_id IS NOT NULL", 
		"").Find(&vms).Error
	return vms, err
}

// GetVMsByStatus는 특정 상태의 시스템 전체 가상머신 목록을 조회합니다
func (r *VirtualMachineRepository) GetVMsByStatus(status string) ([]models.VirtualMachine, error) {
	var vms []models.VirtualMachine
	err := r.db.Preload("Participant").Where("status = ?", status).Find(&vms).Error
	return vms, err
}

// GetVMStatsGroupedByParticipant는 참여자별로 그룹화된 VM 통계를 조회합니다
func (r *VirtualMachineRepository) GetVMStatsGroupedByParticipant() ([]map[string]interface{}, error) {
	var results []struct {
		ParticipantID   string `json:"participant_id"`
		ParticipantName string `json:"participant_name"`
		Total           int64  `json:"total"`
		Active          int64  `json:"active"`
		Busy            int64  `json:"busy"`
		Available       int64  `json:"available"`
		Error           int64  `json:"error"`
	}

	query := `
		SELECT 
			p.id as participant_id,
			p.name as participant_name,
			COUNT(vm.id) as total,
			SUM(CASE WHEN vm.status = 'ACTIVE' THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN vm.current_task_id != '' AND vm.current_task_id IS NOT NULL THEN 1 ELSE 0 END) as busy,
			SUM(CASE WHEN vm.status = 'ACTIVE' AND (vm.current_task_id = '' OR vm.current_task_id IS NULL) THEN 1 ELSE 0 END) as available,
			SUM(CASE WHEN vm.status = 'ERROR' THEN 1 ELSE 0 END) as error
		FROM participants p
		LEFT JOIN virtual_machines vm ON p.id = vm.participant_id
		GROUP BY p.id, p.name
		HAVING COUNT(vm.id) > 0
	`

	err := r.db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 결과를 map 배열로 변환
	var stats []map[string]interface{}
	for _, result := range results {
		stats = append(stats, map[string]interface{}{
			"participant_id":   result.ParticipantID,
			"participant_name": result.ParticipantName,
			"total":            result.Total,
			"active":           result.Active,
			"busy":             result.Busy,
			"available":        result.Available,
			"error":            result.Error,
		})
	}

	return stats, nil
}
