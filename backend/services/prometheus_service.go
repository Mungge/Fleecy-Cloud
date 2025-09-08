package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// CreatePrometheusService: 새로운 PrometheusService 인스턴스를 생성합니다
func CreatePrometheusService(prometheusURL string) *PrometheusService {
	if prometheusURL == "" {
		prometheusURL = "http://127.0.0.1:9090"
	}
	
	return &PrometheusService{
		baseURL: strings.TrimSuffix(prometheusURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// executeQuery는 Prometheus 쿼리를 실행하고 결과를 반환합니다
func (p *PrometheusService) executeQuery(query string) (*PrometheusQueryResult, error) {
	return p.executeQueryWithContext(context.Background(), query)
}

// executeQueryWithContext는 컨텍스트와 함께 Prometheus 쿼리를 실행합니다
func (p *PrometheusService) executeQueryWithContext(ctx context.Context, query string) (*PrometheusQueryResult, error) {
	// URL 인코딩
	params := url.Values{}
	params.Add("query", query)

	queryURL := fmt.Sprintf("%s/api/v1/query?%s", p.baseURL, params.Encode())
	log.Printf("Prometheus 쿼리 실행: %s", queryURL)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %v", err)
	}

	resp, err := p.httpClient.Do(req)
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
	if result.Status != "success" {
		return 0, fmt.Errorf("쿼리 실행 실패: %s", result.Status)
	}

	if len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("결과 데이터가 없습니다")
	}

	if len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("잘못된 결과 형식")
	}

	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("값을 문자열로 변환할 수 없습니다")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("값을 float64로 변환 실패: %v", err)
	}

	return value, nil
}

// GetVMMonitoringInfoWithIP VM IP로 모니터링 정보를 조회합니다
func (p *PrometheusService) GetVMMonitoringInfoWithIP(vmIP string) (*VMMonitoringInfo, error) {
	// 타임아웃 설정
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 먼저 사용 가능한 모든 인스턴스를 확인하고 매칭되는 것을 찾기
	availableInstance := p.findMatchingInstance(vmIP)

	var instanceLabel string
	if availableInstance != "" {
		instanceLabel = availableInstance
	} else {
		instanceLabel = fmt.Sprintf("%s:9100", vmIP)
	}

	log.Printf("VM %s에 대해 인스턴스 라벨 사용: %s", vmIP, instanceLabel)

	// 각 메트릭 조회
	cpuUsage, err := p.GetVMCPUUsageByIPWithContext(ctx, instanceLabel)
	if err != nil {
		log.Printf("CPU 사용률 조회 실패: %v", err)
		cpuUsage = 0
	}

	memoryUsage, err := p.GetVMMemoryUsageByIPWithContext(ctx, instanceLabel)
	if err != nil {
		log.Printf("메모리 사용률 조회 실패: %v", err)
		memoryUsage = 0
	}

	diskUsage, err := p.GetVMDiskUsageByIPWithContext(ctx, instanceLabel)
	if err != nil {
		log.Printf("디스크 사용률 조회 실패: %v", err)
		diskUsage = 0
	}

	networkInBytes, networkOutBytes, err := p.GetVMNetworkStatsByIPWithContext(ctx, instanceLabel)
	if err != nil {
		log.Printf("네트워크 통계 조회 실패: %v", err)
		networkInBytes = 0
		networkOutBytes = 0
	}

	return &VMMonitoringInfo{
		InstanceID:      vmIP,
		CPUUsage:        cpuUsage,
		MemoryUsage:     memoryUsage,
		DiskUsage:       diskUsage,
		NetworkInBytes:  networkInBytes,
		NetworkOutBytes: networkOutBytes,
		LastUpdated:     time.Now(),
	}, nil
}

// GetVMCPUUsageByIP CPU 사용률 조회
func (p *PrometheusService) GetVMCPUUsageByIP(instanceLabel string) (float64, error) {
	return p.GetVMCPUUsageByIPWithContext(context.Background(), instanceLabel)
}

// GetVMCPUUsageByIPWithContext 컨텍스트와 함께 CPU 사용률 조회
func (p *PrometheusService) GetVMCPUUsageByIPWithContext(ctx context.Context, instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle",instance="%s"}[5m])) * 100)`, instanceLabel)
	result, err := p.executeQueryWithContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("CPU 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

// GetVMMemoryUsageByIP 메모리 사용률 조회
func (p *PrometheusService) GetVMMemoryUsageByIP(instanceLabel string) (float64, error) {
	return p.GetVMMemoryUsageByIPWithContext(context.Background(), instanceLabel)
}

// GetVMMemoryUsageByIPWithContext 컨텍스트와 함께 메모리 사용률 조회
func (p *PrometheusService) GetVMMemoryUsageByIPWithContext(ctx context.Context, instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s"} / node_memory_MemTotal_bytes{instance="%s"})) * 100`, instanceLabel, instanceLabel)
	result, err := p.executeQueryWithContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("메모리 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

// GetVMDiskUsageByIP 디스크 사용률 조회
func (p *PrometheusService) GetVMDiskUsageByIP(instanceLabel string) (float64, error) {
	return p.GetVMDiskUsageByIPWithContext(context.Background(), instanceLabel)
}

// GetVMDiskUsageByIPWithContext 컨텍스트와 함께 디스크 사용률 조회
func (p *PrometheusService) GetVMDiskUsageByIPWithContext(ctx context.Context, instanceLabel string) (float64, error) {
	query := fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s",mountpoint="/"} / node_filesystem_size_bytes{instance="%s",mountpoint="/"})) * 100`, instanceLabel, instanceLabel)
	result, err := p.executeQueryWithContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("디스크 사용률 조회 실패: %v", err)
	}
	return p.parseFloatValue(result)
}

// GetVMNetworkStatsByIP 네트워크 통계 조회
func (p *PrometheusService) GetVMNetworkStatsByIP(instanceLabel string) (int64, int64, error) {
	return p.GetVMNetworkStatsByIPWithContext(context.Background(), instanceLabel)
}

// GetVMNetworkStatsByIPWithContext 컨텍스트와 함께 네트워크 통계 조회
func (p *PrometheusService) GetVMNetworkStatsByIPWithContext(ctx context.Context, instanceLabel string) (int64, int64, error) {
	// 네트워크 입력 바이트 - 최근 1시간의 증가량
	inQuery := fmt.Sprintf(`sum(increase(node_network_receive_bytes_total{instance="%s",device!~"lo|docker.*|veth.*"}[1h]))`, instanceLabel)
	inResult, err := p.executeQueryWithContext(ctx, inQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("네트워크 입력 조회 실패: %v", err)
	}

	inBytes, err := p.parseFloatValue(inResult)
	if err != nil {
		inBytes = 0
	}

	// 네트워크 출력 바이트 - 최근 1시간의 증가량
	outQuery := fmt.Sprintf(`sum(increase(node_network_transmit_bytes_total{instance="%s",device!~"lo|docker.*|veth.*"}[1h]))`, instanceLabel)
	outResult, err := p.executeQueryWithContext(ctx, outQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("네트워크 출력 조회 실패: %v", err)
	}

	outBytes, err := p.parseFloatValue(outResult)
	if err != nil {
		outBytes = 0
	}

	return int64(inBytes), int64(outBytes), nil
}

func (p *PrometheusService) findMatchingInstance(vmIP string) string {
    query := "up{job=\"node-exporter\"}"
    result, err := p.executeQuery(query)
    if err != nil {
        return vmIP + ":9100"
    }

    // node-exporter job에서 첫 번째로 UP 상태인 인스턴스 사용
    if result.Status == "success" && len(result.Data.Result) > 0 {
        if instance, ok := result.Data.Result[0].Metric["instance"]; ok {
            log.Printf("첫 번째 node-exporter 인스턴스 사용: %s", instance)
            return instance
        }
    }

    return vmIP + ":9100"
}

// IsHealthy Prometheus 연결 상태 확인
func (p *PrometheusService) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "up"
	_, err := p.executeQueryWithContext(ctx, query)
	if err != nil {
		log.Printf("Prometheus 헬스체크 실패: %v", err)
		return false
	}
	return true
}

// IsInstanceUp 특정 인스턴스가 활성 상태인지 확인
func (p *PrometheusService) IsInstanceUp(instanceLabel string) bool {
	query := fmt.Sprintf(`up{instance="%s"}`, instanceLabel)
	result, err := p.executeQuery(query)
	if err != nil {
		return false
	}

	value, err := p.parseFloatValue(result)
	return err == nil && value == 1
}

// GetPrometheusURL Prometheus URL 반환
func (p *PrometheusService) GetPrometheusURL() string {
	return p.baseURL
}