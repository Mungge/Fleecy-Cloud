package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
)

// OpenStack 인증 토큰 응답
type AuthTokenResponse struct {
	Token struct {
		ID        string    `json:"id"`
		ExpiresAt time.Time `json:"expires_at"`
		Project   struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
	} `json:"token"`
}

// OpenStack 인증 요청
type AuthRequest struct {
	Auth struct {
		Identity struct {
			Methods  []string `json:"methods"`
			Password struct {
				User struct {
					Name     string `json:"name"`
					Password string `json:"password"`
					Domain   struct {
						Name string `json:"name"`
					} `json:"domain"`
				} `json:"user"`
			} `json:"password"`
		} `json:"identity"`
		Scope struct {
			Project struct {
				Name   string `json:"name"`
				Domain struct {
					Name string `json:"name"`
				} `json:"domain"`
			} `json:"project"`
		} `json:"scope"`
	} `json:"auth"`
}

// VM 인스턴스 정보
type VMInstance struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Flavor  struct {
		ID string `json:"id"`
	} `json:"flavor"`
	Image struct {
		ID string `json:"id"`
	} `json:"image"`
	Addresses map[string][]struct {
		Addr    string `json:"addr"`
		Version int    `json:"version"`
		Type    string `json:"type"`
	} `json:"addresses"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// VM 모니터링 정보
type VMMonitoringInfo struct {
	InstanceID      string    `json:"instance_id"`
	Status          string    `json:"status"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkInBytes  int64     `json:"network_in_bytes"`
	NetworkOutBytes int64     `json:"network_out_bytes"`
	LastUpdated     time.Time `json:"last_updated"`
}

// VM 헬스체크 결과
type VMHealthCheckResult struct {
	Healthy     bool      `json:"healthy"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CheckedAt   time.Time `json:"checked_at"`
	ResponseTime int64    `json:"response_time_ms"`
}

type OpenStackService struct {
	client *http.Client
}

func NewOpenStackService() *OpenStackService {
	return &OpenStackService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenStack 인증 토큰 획득
func (s *OpenStackService) GetAuthToken(participant *models.Participant) (string, error) {
	authReq := AuthRequest{}
	authReq.Auth.Identity.Methods = []string{"password"}
	authReq.Auth.Identity.Password.User.Name = participant.OpenStackUsername
	authReq.Auth.Identity.Password.User.Password = participant.OpenStackPassword
	authReq.Auth.Identity.Password.User.Domain.Name = participant.OpenStackDomainName
	authReq.Auth.Scope.Project.Name = participant.OpenStackProjectName
	authReq.Auth.Scope.Project.Domain.Name = participant.OpenStackDomainName

	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("인증 요청 생성 실패: %v", err)
	}

	url := fmt.Sprintf("%s/v3/auth/tokens", participant.OpenStackEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("인증 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("인증 실패: HTTP %d", resp.StatusCode)
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("인증 토큰을 받지 못했습니다")
	}

	return token, nil
}

// VM 인스턴스 정보 조회 (VirtualMachine 인스턴스 기반)
func (s *OpenStackService) GetVMInstance(vm *models.VirtualMachine, participant *models.Participant, token string) (*VMInstance, error) {
	if vm.InstanceID == "" {
		return nil, fmt.Errorf("VM 인스턴스 ID가 설정되지 않았습니다")
	}

	url := fmt.Sprintf("%s/v2.1/servers/%s", participant.OpenStackEndpoint, vm.InstanceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VM 정보 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VM 정보 조회 실패: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var response struct {
		Server VMInstance `json:"server"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return &response.Server, nil
}

// VM 모니터링 정보 수집 (특정 VirtualMachine 인스턴스 기반)
func (s *OpenStackService) MonitorSpecificVM(participant *models.Participant, vm *models.VirtualMachine) (*VMMonitoringInfo, error) {
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return nil, fmt.Errorf("인증 실패: %v", err)
	}

	instance, err := s.GetVMInstance(vm, participant, token)
	if err != nil {
		return nil, fmt.Errorf("VM 인스턴스 조회 실패: %v", err)
	}

	// 실제 환경에서는 OpenStack의 telemetry 서비스(Ceilometer)나 
	// Prometheus 메트릭을 통해 실제 모니터링 데이터를 수집해야 합니다.
	// 여기서는 시뮬레이션 데이터를 반환합니다.
	monitoringInfo := &VMMonitoringInfo{
		InstanceID:      instance.ID,
		Status:          instance.Status,
		CPUUsage:        75.5,  // 실제로는 telemetry API에서 가져옴
		MemoryUsage:     82.3,  // 실제로는 telemetry API에서 가져옴
		DiskUsage:       45.8,  // 실제로는 telemetry API에서 가져옴
		NetworkInBytes:  1024000,
		NetworkOutBytes: 2048000,
		LastUpdated:     time.Now(),
	}

	return monitoringInfo, nil
}

// VM 모니터링 정보 수집 (기존 호환성을 위해 유지)
func (s *OpenStackService) MonitorVM(vm *models.VirtualMachine, participant *models.Participant) (*VMMonitoringInfo, error) {
	return s.MonitorSpecificVM(participant, vm)
}

// VM 헬스체크 수행 (특정 VirtualMachine 인스턴스 기반)
func (s *OpenStackService) HealthCheckSpecificVM(participant *models.Participant, vm *models.VirtualMachine) (*VMHealthCheckResult, error) {
	startTime := time.Now()
	
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return &VMHealthCheckResult{
			Healthy:      false,
			Status:       "ERROR",
			Message:      fmt.Sprintf("인증 실패: %v", err),
			CheckedAt:    time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	instance, err := s.GetVMInstance(vm, participant, token)
	if err != nil {
		return &VMHealthCheckResult{
			Healthy:      false,
			Status:       "ERROR",
			Message:      fmt.Sprintf("VM 조회 실패: %v", err),
			CheckedAt:    time.Now(),
			ResponseTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	healthy := instance.Status == "ACTIVE"
	status := instance.Status
	message := "VM이 정상적으로 동작 중입니다"
	
	if !healthy {
		message = fmt.Sprintf("VM 상태가 비정상입니다: %s", instance.Status)
	}

	return &VMHealthCheckResult{
		Healthy:      healthy,
		Status:       status,
		Message:      message,
		CheckedAt:    time.Now(),
		ResponseTime: time.Since(startTime).Milliseconds(),
	}, nil
}

// VM 헬스체크 수행 (기존 호환성을 위해 유지 - 더 이상 사용하지 않음)
func (s *OpenStackService) HealthCheckVM(participant *models.Participant) (*VMHealthCheckResult, error) {
	return &VMHealthCheckResult{
		Healthy:      false,
		Status:       "ERROR",
		Message:      "이 메서드는 더 이상 지원되지 않습니다. HealthCheckSpecificVM을 사용하세요.",
		CheckedAt:    time.Now(),
		ResponseTime: 0,
	}, fmt.Errorf("deprecated method")
}

// VM 전원 제어 (시작/중지/재부팅) - 특정 VirtualMachine 인스턴스 기반
func (s *OpenStackService) VMPowerActionSpecific(participant *models.Participant, vm *models.VirtualMachine, action string) error {
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return fmt.Errorf("인증 실패: %v", err)
	}

	url := fmt.Sprintf("%s/v2.1/servers/%s/action", participant.OpenStackEndpoint, vm.InstanceID)
	
	var actionData map[string]interface{}
	switch action {
	case "start":
		actionData = map[string]interface{}{"os-start": nil}
	case "stop":
		actionData = map[string]interface{}{"os-stop": nil}
	case "reboot":
		actionData = map[string]interface{}{
			"reboot": map[string]string{"type": "SOFT"},
		}
	default:
		return fmt.Errorf("지원하지 않는 액션입니다: %s", action)
	}

	jsonData, err := json.Marshal(actionData)
	if err != nil {
		return fmt.Errorf("액션 데이터 생성 실패: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("VM 액션 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("VM 액션 실패: HTTP %d", resp.StatusCode)
	}

	return nil
}

// VM 전원 제어 (기존 호환성을 위해 유지 - 더 이상 사용하지 않음)
func (s *OpenStackService) VMPowerAction(participant *models.Participant, action string) error {
	return fmt.Errorf("이 메서드는 더 이상 지원되지 않습니다. VMPowerActionSpecific을 사용하세요")
}

// 연합학습 작업 할당 (특정 VirtualMachine 인스턴스 기반)
func (s *OpenStackService) AssignFederatedLearningTaskSpecific(participant *models.Participant, vm *models.VirtualMachine, taskID string) error {
	// 현재 VM 상태 확인
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return fmt.Errorf("인증 실패: %v", err)
	}

	instance, err := s.GetVMInstance(vm, participant, token)
	if err != nil {
		return fmt.Errorf("VM 상태 확인 실패: %v", err)
	}

	if instance.Status != "ACTIVE" {
		return fmt.Errorf("VM이 활성 상태가 아닙니다: %s", instance.Status)
	}

	// 실제 환경에서는 VM에 SSH 연결하거나 에이전트를 통해 
	// 연합학습 작업을 할당하고 실행합니다.
	// 여기서는 시뮬레이션합니다.
	
	return nil
}

// 연합학습 작업 할당 (기존 호환성을 위해 유지 - 더 이상 사용하지 않음)
func (s *OpenStackService) AssignFederatedLearningTask(participant *models.Participant, taskID string) error {
	return fmt.Errorf("이 메서드는 더 이상 지원되지 않습니다. AssignFederatedLearningTaskSpecific을 사용하세요")
}
