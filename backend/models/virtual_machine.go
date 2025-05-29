package models

import "time"

// VirtualMachine은 참여자가 관리하는 개별 VM을 나타냅니다
type VirtualMachine struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	ParticipantID   string    `json:"participant_id" gorm:"not null;index"`
	Name            string    `json:"name" gorm:"not null"`
	InstanceID      string    `json:"instance_id" gorm:"not null;uniqueIndex"` // 클라우드 제공자의 VM 인스턴스 ID
	Status          string    `json:"status" gorm:"default:unknown"`          // ACTIVE, STOPPED, ERROR, BUILDING, etc.
	IPAddress       string    `json:"ip_address,omitempty"`
	PrivateIP       string    `json:"private_ip,omitempty"`
	FlavorID        string    `json:"flavor_id,omitempty"`                    // VM 사양 ID
	ImageID         string    `json:"image_id,omitempty"`                     // OS 이미지 ID
	
	// 리소스 사양
	VCPUs           int       `json:"vcpus" gorm:"default:0"`
	RAM             int       `json:"ram" gorm:"default:0"`                   // MB 단위
	Disk            int       `json:"disk" gorm:"default:0"`                  // GB 단위
	
	// 모니터링 관련 필드
	LastHealthCheck  time.Time `json:"last_health_check,omitempty"`
	CPUUsage         float64   `json:"cpu_usage" gorm:"default:0"`            // CPU 사용률 (%)
	MemoryUsage      float64   `json:"memory_usage" gorm:"default:0"`         // 메모리 사용률 (%)
	DiskUsage        float64   `json:"disk_usage" gorm:"default:0"`           // 디스크 사용률 (%)
	NetworkInBytes   int64     `json:"network_in_bytes" gorm:"default:0"`     // 네트워크 입력 바이트
	NetworkOutBytes  int64     `json:"network_out_bytes" gorm:"default:0"`    // 네트워크 출력 바이트
	
	// 연합학습 관련 필드
	CurrentTaskID       string    `json:"current_task_id,omitempty"`          // 현재 수행 중인 작업 ID
	TaskAssignedAt      time.Time `json:"task_assigned_at,omitempty"`         // 작업 할당 시간
	LastTaskCompletedAt time.Time `json:"last_task_completed_at,omitempty"`   // 마지막 작업 완료 시간
	TotalTasksCompleted int       `json:"total_tasks_completed" gorm:"default:0"` // 총 완료한 작업 수
	
	// 추가 메타데이터
	Metadata        string    `json:"metadata,omitempty"`                     // JSON 형태의 추가 정보
	
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 관계 설정
	Participant     Participant `json:"participant,omitempty" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (VirtualMachine) TableName() string {
	return "virtual_machines"
}

// IsAvailable은 VM이 새로운 작업을 받을 수 있는지 확인합니다
func (vm *VirtualMachine) IsAvailable() bool {
	return vm.Status == "ACTIVE" && vm.CurrentTaskID == ""
}

// IsBusy는 VM이 현재 작업 중인지 확인합니다
func (vm *VirtualMachine) IsBusy() bool {
	return vm.CurrentTaskID != ""
}
