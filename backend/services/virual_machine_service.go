package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
			Methods               []string `json:"methods"`
			ApplicationCredential *struct {
				ID     string `json:"id"`
				Secret string `json:"secret"`
			} `json:"application_credential,omitempty"`
		} `json:"identity"`
	} `json:"auth"`
}

// Flavor 상세 정보
type FlavorDetails struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	VCPUs int    `json:"vcpus"`
	RAM   int    `json:"ram"`  // MB 단위
	Disk  int    `json:"disk"` // GB 단위
}

// VM 인스턴스 정보
type VMInstance struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Flavor    FlavorDetails `json:"flavor"`
	Addresses map[string][]struct {
		Addr string `json:"addr"`
		Type string `json:"OS-EXT-IPS:type"`
	} `json:"addresses"`
	PowerState       int    `json:"OS-EXT-STS:power_state"`
	AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

// VM 목록 조회 응답
type VMListResponse struct {
	Servers []VMInstance `json:"servers"`
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
	Healthy      bool      `json:"healthy"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	CheckedAt    time.Time `json:"checked_at"`
	ResponseTime int64     `json:"response_time_ms"`
}

type OpenStackService struct {
	client            *http.Client
	prometheusService *PrometheusService
}

func NewOpenStackService(prometheusURL string) *OpenStackService {
	return &OpenStackService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		prometheusService: NewPrometheusService(prometheusURL),
	}
}

// OpenStack 인증 토큰 획득 -> TestConnection
func (s *OpenStackService) GetAuthToken(participant *models.Participant) (string, error) {
	authReq := AuthRequest{}

	// Application Credential 방식만 지원
	if participant.OpenStackApplicationCredentialID != "" && participant.OpenStackApplicationCredentialSecret != "" {
		// Application Credential 방식
		authReq.Auth.Identity.Methods = []string{"application_credential"}
		authReq.Auth.Identity.ApplicationCredential = &struct {
			ID     string `json:"id"`
			Secret string `json:"secret"`
		}{
			ID:     participant.OpenStackApplicationCredentialID,
			Secret: participant.OpenStackApplicationCredentialSecret,
		}
	} else {
		return "", fmt.Errorf("application Credential 인증 정보가 필요합니다")
	}

	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("인증 요청 생성 실패: %v", err)
	}

	url := fmt.Sprintf("%s/identity/v3/auth/tokens", participant.OpenStackEndpoint)
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

func (s *OpenStackService) GetAllVMInstances(participant *models.Participant) ([]VMInstance, error) {
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return nil, fmt.Errorf("인증 실패: %v", err)
	}

	url := fmt.Sprintf("%s/compute/v2.1/servers/detail", participant.OpenStackEndpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VM 목록 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VM 목록 조회 실패: HTTP %d, 응답: %s", resp.StatusCode, string(body))
	}

	// 먼저 기본 VM 정보를 파싱
	var basicResponse struct {
		Servers []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
			Flavor struct {
				ID string `json:"id"`
			} `json:"flavor"`
			Addresses map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			} `json:"addresses"`
			PowerState       int    `json:"OS-EXT-STS:power_state"`
			AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`
			Created          string `json:"created"`
			Updated          string `json:"updated"`
		} `json:"servers"`
	}

	if err := json.Unmarshal(body, &basicResponse); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v, 응답 내용: %s", err, string(body))
	}

	// 각 VM에 대해 flavor 상세 정보를 가져와서 완전한 VMInstance 생성
	var vmInstances []VMInstance
	for _, server := range basicResponse.Servers {
		flavorDetails, err := s.GetFlavorDetails(participant, token, server.Flavor.ID)
		if err != nil {
			// Flavor 정보를 가져오지 못한 경우 기본값으로 설정
			flavorDetails = &FlavorDetails{
				ID:    server.Flavor.ID,
				Name:  "Unknown",
				VCPUs: 0,
				RAM:   0,
				Disk:  0,
			}
		}

		vmInstance := VMInstance{
			ID:         server.ID,
			Name:       server.Name,
			Status:     server.Status,
			Flavor:     *flavorDetails,
			Addresses:  server.Addresses,
			PowerState: server.PowerState,
			Created:    server.Created,
			Updated:    server.Updated,
		}

		vmInstances = append(vmInstances, vmInstance)
	}

	return vmInstances, nil
}

// VM 인스턴스 정보 조회
func (s *OpenStackService) GetVMInstance(vm *models.VirtualMachine, participant *models.Participant, token string) (*VMInstance, error) {
	if vm.InstanceID == "" {
		return nil, fmt.Errorf("VM 인스턴스 ID가 설정되지 않았습니다")
	}

	url := fmt.Sprintf("%s/compute/v2.1/servers/%s", participant.OpenStackEndpoint, vm.InstanceID)
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

	var basicResponse struct {
		Server struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
			Flavor struct {
				ID string `json:"id"`
			} `json:"flavor"`
			Addresses map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			} `json:"addresses"`
			PowerState       int    `json:"OS-EXT-STS:power_state"`
			AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`
			Created          string `json:"created"`
			Updated          string `json:"updated"`
		} `json:"server"`
	}

	if err := json.Unmarshal(body, &basicResponse); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	// Flavor 상세 정보 조회
	flavorDetails, err := s.GetFlavorDetails(participant, token, basicResponse.Server.Flavor.ID)
	if err != nil {
		// Flavor 정보를 가져오지 못한 경우 기본값으로 설정
		flavorDetails = &FlavorDetails{
			ID:    basicResponse.Server.Flavor.ID,
			Name:  "Unknown",
			VCPUs: 0,
			RAM:   0,
			Disk:  0,
		}
	}

	vmInstance := &VMInstance{
		ID:         basicResponse.Server.ID,
		Name:       basicResponse.Server.Name,
		Status:     basicResponse.Server.Status,
		Flavor:     *flavorDetails,
		Addresses:  basicResponse.Server.Addresses,
		PowerState: basicResponse.Server.PowerState,
		Created:    basicResponse.Server.Created,
		Updated:    basicResponse.Server.Updated,
	}

	return vmInstance, nil
}

// VM 목록 조회
func (s *OpenStackService) ListVMInstances(participant *models.Participant, token string) ([]VMInstance, error) {
	url := fmt.Sprintf("%s/v2.1/servers/detail", participant.OpenStackEndpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VM 목록 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VM 목록 조회 실패: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var response VMListResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return response.Servers, nil
}

// VM 헬스체크 수행
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

// GetFlavorDetails는 특정 flavor의 상세 정보를 조회합니다
func (s *OpenStackService) GetFlavorDetails(participant *models.Participant, token string, flavorID string) (*FlavorDetails, error) {
	url := fmt.Sprintf("%s/compute/v2.1/flavors/%s", participant.OpenStackEndpoint, flavorID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flavor 정보 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flavor 정보 조회 실패: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var response struct {
		Flavor FlavorDetails `json:"flavor"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return &response.Flavor, nil
}

// GetVMRuntimeStatus는 실시간 VM 상태를 조회합니다 (DB에 저장하지 않음)
func (s *OpenStackService) GetVMRuntimeStatus(participant *models.Participant, instanceID string) (*models.VMRuntimeInfo, error) {
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return nil, fmt.Errorf("인증 실패: %v", err)
	}

	url := fmt.Sprintf("%s/compute/v2.1/servers/%s", participant.OpenStackEndpoint, instanceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VM 상태 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		Server struct {
			Status     string `json:"status"`
			PowerState int    `json:"OS-EXT-STS:power_state"`
		} `json:"server"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %v", err)
	}

	return &models.VMRuntimeInfo{
		InstanceID:  instanceID,
		Status:      response.Server.Status,
		PowerState:  response.Server.PowerState,
		LastChecked: time.Now(),
	}, nil
}

// GetVMMonitoringInfoWithParticipant는 participant의 OpenStack endpoint를 사용하여 모니터링 정보를 조회합니다
func (s *OpenStackService) GetVMMonitoringInfoWithParticipant(participant *models.Participant, instanceID string) (*models.VMMonitoringInfo, error) {
	if participant == nil {
		return nil, fmt.Errorf("participant 정보가 필요합니다")
	}

	if participant.OpenStackEndpoint == "" {
		return nil, fmt.Errorf("participant의 OpenStack endpoint가 설정되지 않았습니다")
	}

	// participant의 OpenStack endpoint를 사용하여 Prometheus URL 생성
	// OpenStack endpoint에서 포트 9090으로 Prometheus에 접근
	prometheusURL := fmt.Sprintf("%s:9090", participant.OpenStackEndpoint)

	// VM의 IP 주소 가져오기 (OpenStack API 호출)
	vmIP, err := s.getVMIPAddress(participant, instanceID)
	if err != nil {
		vmIP = instanceID // IP 조회 실패 시 인스턴스 ID 사용
	}

	// 해당 participant 전용 Prometheus 서비스 생성
	prometheusService := NewPrometheusService(prometheusURL)

	return prometheusService.GetVMMonitoringInfoWithIP(vmIP)
}

// getVMIPAddress는 VM의 IP 주소를 OpenStack에서 조회합니다
func (s *OpenStackService) getVMIPAddress(participant *models.Participant, instanceID string) (string, error) {
	token, err := s.GetAuthToken(participant)
	if err != nil {
		return "", fmt.Errorf("인증 실패: %v", err)
	}

	url := fmt.Sprintf("%s/compute/v2.1/servers/%s", participant.OpenStackEndpoint, instanceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("HTTP 요청 생성 실패: %v", err)
	}

	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("VM 정보 조회 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("VM 정보 조회 실패: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var response struct {
		Server struct {
			Addresses map[string][]struct {
				Addr string `json:"addr"`
				Type string `json:"OS-EXT-IPS:type"`
			} `json:"addresses"`
		} `json:"server"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("응답 파싱 실패: %v", err)
	}

	// IPv4 주소만 찾기
	for _, addresses := range response.Server.Addresses {
		for _, addr := range addresses {
			// IPv4 주소만 사용 (콜론이 없는 주소)
			if !strings.Contains(addr.Addr, ":") {
				return addr.Addr, nil
			}
		}
	}

	return "", fmt.Errorf("VM에 할당된 IP 주소가 없습니다")
}
