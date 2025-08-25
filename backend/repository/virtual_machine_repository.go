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

// AssignTask는 가상머신에 작업을 할당합니다
func (r *VirtualMachineRepository) AssignTask(id string, taskID string) error {
	updates := map[string]interface{}{
		"current_task_id":  taskID,
		"task_assigned_at": "NOW()",
	}
	return r.db.Model(&models.VirtualMachine{}).Where("id = ?", id).Updates(updates).Error
}
