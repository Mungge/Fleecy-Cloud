package models

import "time"

type Participant struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"` // 클라우드 참여자(회사) 이름
	Status    string    `json:"status" gorm:"default:inactive"` // "active", "inactive", "busy", "error", "pending"
	Metadata  string    `json:"metadata,omitempty"` // JSON 형태의 추가 메타데이터
	
	// OpenStack 클라우드 관련 필드
	OpenStackEndpoint    string `json:"openstack_endpoint,omitempty"`     // OpenStack 인증 엔드포인트
	OpenStackUsername    string `json:"openstack_username,omitempty"`     // OpenStack 사용자명
	OpenStackPassword    string `json:"openstack_password,omitempty"`     // OpenStack 비밀번호 (암호화 저장)
	OpenStackProjectName string `json:"openstack_project_name,omitempty"` // OpenStack 프로젝트명
	OpenStackDomainName  string `json:"openstack_domain_name,omitempty"`  // OpenStack 도메인명
	OpenStackRegion      string `json:"openstack_region,omitempty"`       // OpenStack 리전
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 관계 설정
	User                 User                   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FederatedLearnings   []FederatedLearning    `json:"federated_learnings,omitempty" gorm:"many2many:participant_federated_learnings;"`
	VirtualMachines      []VirtualMachine       `json:"virtual_machines,omitempty" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Participant) TableName() string {
	return "participants"
}

// GetAvailableVMs는 사용 가능한 VM 목록을 반환합니다
func (p *Participant) GetAvailableVMs() []VirtualMachine {
	var availableVMs []VirtualMachine
	for _, vm := range p.VirtualMachines {
		if vm.IsAvailable() {
			availableVMs = append(availableVMs, vm)
		}
	}
	return availableVMs
}

// GetBusyVMs는 작업 중인 VM 목록을 반환합니다
func (p *Participant) GetBusyVMs() []VirtualMachine {
	var busyVMs []VirtualMachine
	for _, vm := range p.VirtualMachines {
		if vm.IsBusy() {
			busyVMs = append(busyVMs, vm)
		}
	}
	return busyVMs
}

// GetTotalCapacity는 전체 VM 수용량을 계산합니다
func (p *Participant) GetTotalCapacity() int {
	return len(p.VirtualMachines)
}

// GetAvailableCapacity는 사용 가능한 VM 수를 반환합니다
func (p *Participant) GetAvailableCapacity() int {
	return len(p.GetAvailableVMs())
}
