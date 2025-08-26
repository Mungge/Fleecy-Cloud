package services

import (
	"encoding/json"
	"time"
)

// VMRuntimeInfo는 실시간으로 변하는 VM 상태 정보 (DB에 저장하지 않음)
type VMRuntimeInfo struct {
	InstanceID  string    `json:"instance_id"`
	Status      string    `json:"status"`      // ACTIVE, SHUTOFF, ERROR 등
	PowerState  int       `json:"power_state"` // 1: Running, 4: Shutdown
	LastChecked time.Time `json:"last_checked"`
}

// VMMonitoringInfo는 모니터링 데이터 (캐시/별도 시스템에서 관리)
type VMMonitoringInfo struct {
	InstanceID      string    `json:"instance_id"`
	CPUUsage        float64   `json:"cpu_usage"`         // CPU 사용률 (%)
	MemoryUsage     float64   `json:"memory_usage"`      // 메모리 사용률 (%)
	DiskUsage       float64   `json:"disk_usage"`        // 디스크 사용률 (%)
	NetworkInBytes  int64     `json:"network_in_bytes"`  // 네트워크 입력 바이트
	NetworkOutBytes int64     `json:"network_out_bytes"` // 네트워크 출력 바이트
	LastUpdated     time.Time `json:"last_updated"`
}

// VirtualMachine은 OpenStack에서 조회되는 VM 정보를 나타냅니다 (DB 저장하지 않음)
type VirtualMachine struct {
	ID            string `json:"id"`
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
	InstanceID    string `json:"instance_id"` // 클라우드 제공자의 VM 인스턴스 ID
	Status        string `json:"status"`      // ACTIVE, STOPPED, ERROR, BUILDING, etc.
	IPAddress     string `json:"ip_address,omitempty"`
	PrivateIP     string `json:"private_ip,omitempty"`
	FlavorID      string `json:"flavor_id,omitempty"` // VM 사양 ID
	ImageID       string `json:"image_id,omitempty"`  // OS 이미지 ID

	// 리소스 사양
	FlavorName string `json:"flavor_name,omitempty"` // VM 사양 이름
	VCPUs      int    `json:"vcpus"`
	RAM        int    `json:"ram"`  // MB 단위
	Disk       int    `json:"disk"` // GB 단위

	// 네트워크 정보
	IPAddresses      string `json:"ip_addresses,omitempty"`      // JSON 형태로 저장
	AvailabilityZone string `json:"availability_zone,omitempty"` // OpenStack AZ

	// 추가 메타데이터
	Metadata string `json:"metadata,omitempty"` // JSON 형태의 추가 정보

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// IsAvailable은 VM이 새로운 작업을 받을 수 있는지 확인합니다
func (vm *VirtualMachine) IsAvailable() bool {
	return vm.Status == "ACTIVE"
}

// GetIPAddressesMap은 저장된 IP 주소들을 맵 형태로 반환합니다
func (vm *VirtualMachine) GetIPAddressesMap() map[string]interface{} {
	if vm.IPAddresses == "" {
		return make(map[string]interface{})
	}

	var addresses map[string]interface{}
	// JSON 파싱 에러가 발생하면 빈 맵 반환
	if err := json.Unmarshal([]byte(vm.IPAddresses), &addresses); err != nil {
		return make(map[string]interface{})
	}

	return addresses
}

// SetIPAddressesFromMap은 맵 형태의 IP 주소들을 JSON으로 저장합니다
func (vm *VirtualMachine) SetIPAddressesFromMap(addresses map[string]interface{}) error {
	data, err := json.Marshal(addresses)
	if err != nil {
		return err
	}
	vm.IPAddresses = string(data)
	return nil
}

// VMCompleteInfo는 DB 정보 + 실시간 정보를 조합한 완전한 VM 정보
type VMCompleteInfo struct {
	VirtualMachine VirtualMachine    `json:"virtual_machine"`
	RuntimeInfo    *VMRuntimeInfo    `json:"runtime_info,omitempty"`
	MonitoringInfo *VMMonitoringInfo `json:"monitoring_info,omitempty"`
}
