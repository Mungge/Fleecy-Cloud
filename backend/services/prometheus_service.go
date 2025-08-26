package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PrometheusService는 Prometheus API를 통해 메트릭 데이터를 수집하는 서비스입니다
type PrometheusService struct {
	baseURL    string
	httpClient *http.Client
}

// PrometheusQueryResult는 Prometheus API 응답 구조체입니다
type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// NewPrometheusService는 새로운 PrometheusService 인스턴스를 생성합니다
func NewPrometheusService(prometheusURL string) *PrometheusService {
	return &PrometheusService{
		baseURL: strings.TrimSuffix(prometheusURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// executeQuery는 Prometheus 쿼리를 실행하고 결과를 반환합니다
func (p *PrometheusService) executeQuery(query string) (*PrometheusQueryResult, error) {
	// URL 인코딩
	params := url.Values{}
	params.Add("query", query)

	queryURL := fmt.Sprintf("%s/api/v1/query?%s", p.baseURL, params.Encode())

	resp, err := p.httpClient.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("prometheus 쿼리 실행 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus API 오류 (상태코드: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %v", err)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %v", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus 쿼리 실패: %s", result.Status)
	}

	return &result, nil
}

// parseFloatValue는 Prometheus 응답에서 float 값을 추출합니다
func (p *PrometheusService) parseFloatValue(result *PrometheusQueryResult) (float64, error) {
	if len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("결과가 없습니다")
	}

	if len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("값이 없습니다")
	}

	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("값이 문자열이 아닙니다")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("float 변환 실패: %v", err)
	}

	return value, nil
}

// GetVMMonitoringInfoWithIP는 VM의 IP 주소를 사용하여 모니터링 정보를 조회합니다
func (p *PrometheusService) GetVMMonitoringInfoWithIP(vmIP string) (*VMMonitoringInfo, error) {
	// 먼저 사용 가능한 모든 인스턴스를 확인하고 매칭되는 것을 찾기
	availableInstance := p.findMatchingInstance(vmIP)

	var instanceLabel string
	if availableInstance != "" {
		instanceLabel = availableInstance
	} else {
		// Node Exporter는 보통 :9100 포트를 사용
		instanceLabel = fmt.Sprintf("%s:9100", vmIP)
	}

	// CPU 사용률 조회 (IP 주소 기반)
	cpuUsage, err := p.GetVMCPUUsageByIP(instanceLabel)
	if err != nil {
		cpuUsage = 0
	}

	// 메모리 사용률 조회 (IP 주소 기반)
	memoryUsage, err := p.GetVMMemoryUsageByIP(instanceLabel)
	if err != nil {
		memoryUsage = 0
	}

	// 디스크 사용률 조회 (IP 주소 기반)
	diskUsage, err := p.GetVMDiskUsageByIP(instanceLabel)
	if err != nil {
		diskUsage = 0
	}

	// 네트워크 통계 조회 (IP 주소 기반)
	networkInBytes, networkOutBytes, err := p.GetVMNetworkStatsByIP(instanceLabel)
	if err != nil {
		networkInBytes = 0
		networkOutBytes = 0
	}

	return &VMMonitoringInfo{
		InstanceID:      vmIP, // IP 주소를 인스턴스 ID로 사용
		CPUUsage:        cpuUsage,
		MemoryUsage:     memoryUsage,
		DiskUsage:       diskUsage,
		NetworkInBytes:  networkInBytes,
		NetworkOutBytes: networkOutBytes,
		LastUpdated:     time.Now(),
	}, nil
}

// IP 주소 기반 쿼리 함수들
func (p *PrometheusService) GetVMCPUUsageByIP(instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance="%s"}[5m])) * 100)`, instanceLabel)
	result, err := p.executeQuery(query)
	if err != nil {
		return 0, fmt.Errorf("CPU 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

func (p *PrometheusService) GetVMMemoryUsageByIP(instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s"} / node_memory_MemTotal_bytes{instance="%s"})) * 100`, instanceLabel, instanceLabel)
	result, err := p.executeQuery(query)
	if err != nil {
		return 0, fmt.Errorf("메모리 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

func (p *PrometheusService) GetVMDiskUsageByIP(instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s",mountpoint="/"} / node_filesystem_size_bytes{instance="%s",mountpoint="/"})) * 100`, instanceLabel, instanceLabel)
	result, err := p.executeQuery(query)
	if err != nil {
		return 0, fmt.Errorf("디스크 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

func (p *PrometheusService) GetVMNetworkStatsByIP(instanceLabel string) (int64, int64, error) {
	// 네트워크 입력 바이트 쿼리
	inQuery := fmt.Sprintf(`rate(node_network_receive_bytes_total{instance="%s",device!~"lo|docker.*|veth.*"}[5m])`, instanceLabel)
	inResult, err := p.executeQuery(inQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("네트워크 입력 조회 실패: %v", err)
	}

	inBytes, err := p.parseFloatValue(inResult)
	if err != nil {
		inBytes = 0
	}

	// 네트워크 출력 바이트 쿼리
	outQuery := fmt.Sprintf(`rate(node_network_transmit_bytes_total{instance="%s",device!~"lo|docker.*|veth.*"}[5m])`, instanceLabel)
	outResult, err := p.executeQuery(outQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("네트워크 출력 조회 실패: %v", err)
	}

	outBytes, err := p.parseFloatValue(outResult)
	if err != nil {
		outBytes = 0
	}

	return int64(inBytes), int64(outBytes), nil
}

// findMatchingInstance는 VM IP와 매칭되는 Prometheus 인스턴스를 찾습니다
func (p *PrometheusService) findMatchingInstance(vmIP string) string {
	query := "up"
	result, err := p.executeQuery(query)
	if err != nil {
		return ""
	}

	if result.Status == "success" && len(result.Data.Result) > 0 {
		for _, item := range result.Data.Result {
			if instance, ok := item.Metric["instance"]; ok {
				// IP 주소가 포함된 인스턴스 찾기
				if strings.Contains(instance, vmIP) {
					return instance
				}

				// IPv4 주소를 포함한 인스턴스만 반환
				if strings.Contains(instance, ":9100") && !strings.Contains(instance, "[") {
					return instance
				}
			}
		}
	}

	return ""
}
