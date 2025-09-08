package aggregator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/Mungge/Fleecy-Cloud/utils"
	"github.com/gin-gonic/gin"
)

// MLflow API 응답 구조체
type MLflowExperiment struct {
	ExperimentID   string `json:"experiment_id"`
	Name           string `json:"name"`
	LifecycleStage string `json:"lifecycle_stage"`
}

type MLflowRun struct {
	Info struct {
		RunID     string `json:"run_id"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
		Status    string `json:"status"`
	} `json:"info"`
	Data struct {
		Metrics []struct {
			Key       string  `json:"key"`
			Value     float64 `json:"value"`
			Timestamp int64   `json:"timestamp"`
			Step      int     `json:"step"`
		} `json:"metrics"`
		Params []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"data"`
}

type TrainingRound struct {
	Round             int     `json:"round"`
	Accuracy          float64 `json:"accuracy"`
	Loss              float64 `json:"loss"`
	F1Score           float64 `json:"f1_score"`
	Precision         float64 `json:"precision"`
	Recall            float64 `json:"recall"`
	Duration          int     `json:"duration"`
	ParticipantsCount int     `json:"participantsCount"`
	Timestamp         string  `json:"timestamp"`
}

type MLflowClient struct {
	BaseURL string
	Client  *http.Client
}

func NewMLflowClient(baseURL string) *MLflowClient {
	return &MLflowClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 핸들러 구조체
type MLflowHandler struct {
	aggregatorRepo    *repository.AggregatorRepository
	prometheusService *services.PrometheusService
}

func NewMLflowHandler(mlflowURL string, aggregatorRepo *repository.AggregatorRepository, prometheusService *services.PrometheusService) *MLflowHandler {
	return &MLflowHandler{
		aggregatorRepo:    aggregatorRepo,
		prometheusService: prometheusService,
	}
}

// 실험 목록 조회
func (m *MLflowClient) GetExperiments() ([]MLflowExperiment, error) {
	resp, err := m.Client.Get(fmt.Sprintf("%s/api/2.0/mlflow/experiments/search?max_results=10", m.BaseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Experiments []MLflowExperiment `json:"experiments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Experiments, nil
}

// 특정 실험의 runs 조회
func (m *MLflowClient) GetRuns(experimentID string) ([]MLflowRun, error) {
	payload := map[string]interface{}{
		"experiment_ids": []string{experimentID},
		"max_results":    1000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := m.Client.Post(
		fmt.Sprintf("%s/api/2.0/mlflow/runs/search", m.BaseURL),
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Runs []MLflowRun `json:"runs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Runs, nil
}

// 특정 run의 메트릭 히스토리 조회
func (m *MLflowClient) GetMetricHistory(runID, metricKey string) ([]struct {
	Step      int     `json:"step"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}, error) {
	url := fmt.Sprintf("%s/ajax-api/2.0/mlflow/metrics/get-history-bulk-interval?run_ids=%s&metric_key=%s&max_results=50",
		m.BaseURL, runID, metricKey)

	resp, err := m.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Metrics []struct {
			Step      int     `json:"step"`
			Value     float64 `json:"value"`
			Timestamp int64   `json:"timestamp"`
		} `json:"metrics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Metrics, nil
}

// MLflow 실험 생성
func (m *MLflowClient) CreateExperiment(experimentName string) (string, error) {
	payload := map[string]interface{}{
		"name": experimentName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := m.Client.Post(
		fmt.Sprintf("%s/api/2.0/mlflow/experiments/create", m.BaseURL),
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ExperimentID string `json:"experiment_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ExperimentID, nil
}

// GetTrainingHistory godoc
// @Summary 학습 히스토리 조회
// @Description MLflow에서 연합학습 메트릭을 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {array} TrainingRound
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/training-history [get]
func (h *MLflowHandler) GetTrainingHistory(c *gin.Context) {
	// 사용자 인증 확인
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")

	// DB에서 aggregator 정보 조회 (권한 확인 포함)
	aggregator, err := h.aggregatorRepo.GetAggregatorByIDWithFederatedLearning(aggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Aggregator 조회 실패",
			"details": err.Error(),
		})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// MLflow 클라이언트를 aggregator IP로 동적 생성
	var mlflowURL string
	if aggregator.PublicIP != "" {
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
	} else {
		// TODO: 실제 aggregator IP로 변경 필요
		// fallback으로 환경변수나 고정 IP 사용
		aggregatorIP := os.Getenv("AGGREGATOR_IP")
		if aggregatorIP == "" {
			aggregatorIP = "YOUR_ACTUAL_AGGREGATOR_IP" // 실제 IP로 변경하세요
		}
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregatorIP)
		log.Printf("GetTrainingHistory - PublicIP가 없어서 환경변수/기본값 사용: %s", aggregatorIP)
	}
	
	mlflowClient := NewMLflowClient(mlflowURL)
	log.Printf("GetTrainingHistory - Using MLflow URL: %s for aggregator: %s (PublicIP: %s)", mlflowURL, aggregatorID, aggregator.PublicIP)

	// MLflow에서 실험 조회
	experiments, err := mlflowClient.GetExperiments()
	if err != nil {
		log.Printf("MLflow 연결 실패: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "MLflow 실험 조회 실패",
			"details": err.Error(),
		})
		return
	}

	log.Printf("MLflow에서 찾은 실험들:")
	for _, exp := range experiments {
		log.Printf("- Experiment ID: %s, Name: %s", exp.ExperimentID, exp.Name)
	}

	// aggregator의 federated learning ID를 통해 실험 찾기
	var experimentID string
	var expectedExperimentName string
	
	// aggregator에 연결된 federated learning 조회
	if aggregator.FederatedLearning != nil {
		expectedExperimentName = fmt.Sprintf("federated-learning-%s", aggregator.FederatedLearning.ID)
		log.Printf("연결된 FederatedLearning ID: %s", aggregator.FederatedLearning.ID)
	} else {
		// FederatedLearning 관계가 로드되지 않은 경우, aggregator ID로 시도
		expectedExperimentName = fmt.Sprintf("federated-learning-%s", aggregatorID)
		log.Printf("FederatedLearning 관계 없음, aggregator ID 사용: %s", aggregatorID)
	}
	
	log.Printf("찾고 있는 실험명: %s", expectedExperimentName)
	
	for _, exp := range experiments {
		if exp.Name == expectedExperimentName {
			experimentID = exp.ExperimentID
			log.Printf("GetTrainingHistory - 매칭되는 연합학습 실험 '%s' 사용: %s", exp.Name, experimentID)
			break
		}
	}

	if experimentID == "" {
		log.Printf("GetTrainingHistory - 실험을 찾을 수 없음. 빈 배열 반환")
		// 실험이 없으면 빈 배열 반환
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	log.Printf("GetTrainingHistory - 사용할 실험 ID: %s", experimentID)

	// runs 조회
	runs, err := mlflowClient.GetRuns(experimentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "MLflow runs 조회 실패",
			"details": err.Error(),
		})
		return
	}

	if len(runs) == 0 {
		// 빈 배열 반환
		c.JSON(http.StatusOK, []TrainingRound{})
		return
	}

	// 2. 가장 최신 run을 찾습니다. (시작 시간 기준)
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Info.StartTime > runs[j].Info.StartTime
	})
	latestRun := &runs[0]

	// 3. (수정) 요약 정보 대신, 조회할 고정된 메트릭 키 목록을 정의합니다.
	metricKeys := []string{"accuracy", "val_loss", "train_loss", "f1_macro", "precision_macro", "recall_macro"}

	// 4. 라운드(step)별로 모든 메트릭 데이터를 취합할 맵을 준비합니다.
	stepData := make(map[int]map[string]interface{})

	// 5. 정의된 모든 메트릭 '키'에 대해 GetMetricHistory를 '각각' 호출합니다.
	for _, key := range metricKeys {
		history, err := mlflowClient.GetMetricHistory(latestRun.Info.RunID, key)
		if err != nil {
			// 특정 메트릭 조회가 실패해도 로그만 남기고 계속 진행 (예: train_loss가 없는 경우)
			fmt.Printf("정보: 메트릭 '%s'의 히스토리 조회 실패 (값이 없을 수 있음): %v\n", key, err)
			continue
		}

		// 6. 조회된 히스토리를 stepData 맵에 채워넣습니다.
		for _, point := range history {
			if _, ok := stepData[point.Step]; !ok {
				stepData[point.Step] = make(map[string]interface{})
			}
			stepData[point.Step][key] = point.Value
			stepData[point.Step]["timestamp"] = point.Timestamp
		}
	}

	// 7. buildTrainingHistory 헬퍼 함수를 호출하여 최종 결과를 생성합니다.
	trainingHistory := h.buildTrainingHistory(stepData, latestRun, aggregator)

	// 만약 trainingHistory가 nil이면 빈 슬라이스로 바꿔서 JSON `[]`을 반환하도록 보장
	if trainingHistory == nil {
		trainingHistory = make([]TrainingRound, 0)
	}

	c.JSON(http.StatusOK, trainingHistory)
}

// 실시간 메트릭 조회
func (h *MLflowHandler) GetRealTimeMetrics(c *gin.Context) {
	// 사용자 인증 확인
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")

	// DB에서 aggregator 정보 조회
	aggregator, err := h.aggregatorRepo.GetAggregatorByIDWithFederatedLearning(aggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// MLflow 클라이언트를 aggregator IP로 동적 생성
	var mlflowURL string
	if aggregator.PublicIP != "" {
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
	} else {
		// TODO: 실제 aggregator IP로 변경 필요
		// fallback으로 환경변수나 고정 IP 사용
		aggregatorIP := os.Getenv("AGGREGATOR_IP")
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregatorIP)
	}
	
	mlflowClient := NewMLflowClient(mlflowURL)
	log.Printf("GetRealTimeMetrics - Using MLflow URL: %s for aggregator: %s (PublicIP: %s)", mlflowURL, aggregatorID, aggregator.PublicIP)

	// 실험 및 runs 조회
	experiments, err := mlflowClient.GetExperiments()
	if err != nil {
		log.Printf("GetRealTimeMetrics - MLflow 연결 실패: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MLflow 연결 실패"})
		return
	}

	log.Printf("GetRealTimeMetrics - MLflow에서 찾은 실험들:")
	for _, exp := range experiments {
		log.Printf("- Experiment ID: %s, Name: %s", exp.ExperimentID, exp.Name)
	}

	// aggregator의 federated learning ID를 통해 실험 찾기
	var experimentID string
	var expectedExperimentName string
	
	// aggregator에 연결된 federated learning 조회
	if aggregator.FederatedLearning != nil {
		expectedExperimentName = fmt.Sprintf("federated-learning-%s", aggregator.FederatedLearning.ID)
	} else {
		// FederatedLearning 관계가 로드되지 않은 경우, aggregator ID로 시도
		expectedExperimentName = fmt.Sprintf("federated-learning-%s", aggregatorID)
	}
	
	for _, exp := range experiments {
		if exp.Name == expectedExperimentName {
			experimentID = exp.ExperimentID
			break
		}
	}
	
	if experimentID == "" {
		log.Printf("GetRealTimeMetrics - 실험을 찾을 수 없음. 기본값 반환")
		// 실험이 없으면 기본값 반환
		c.JSON(http.StatusOK, gin.H{
			"accuracy":     0,
			"loss":         0,
			"currentRound": 0,
			"status":       "pending",
			"f1_score":     0,
			"precision":    0,
			"recall":       0,
			"run_id":       "", // 추가
		})
		return
	}

	log.Printf("GetRealTimeMetrics - 사용할 실험 ID: %s", experimentID)

	runs, err := mlflowClient.GetRuns(experimentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "runs 조회 실패"})
		return
	}

	if len(runs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"accuracy":     0,
			"loss":         0,
			"currentRound": 0,
			"status":       "pending",
			"f1_score":     0,
			"precision":    0,
			"recall":       0,
			"run_id":       "", // 추가
		})
		return
	}

	latestRun := runs[0]

	// 최신 메트릭 추출
	latestMetrics, maxStep := h.extractLatestMetrics(latestRun.Data.Metrics)

	status := h.mapRunStatus(latestRun.Info.Status)

	c.JSON(http.StatusOK, gin.H{
		"accuracy":     getMetricValue(latestMetrics, "accuracy", 0.0) * 100, // 퍼센트로 변환
		"loss":         getMetricValue(latestMetrics, "val_loss", 0.0),
		"currentRound": maxStep,
		"status":       status,
		"f1_score":     getMetricValue(latestMetrics, "f1_macro", 0.0),
		"precision":    getMetricValue(latestMetrics, "precision_macro", 0.0),
		"recall":       getMetricValue(latestMetrics, "recall_macro", 0.0),
		"run_id":       latestRun.Info.RunID, // 추가된 부분
	})
}

// GetSystemMetrics godoc
// @Summary 집계자 시스템 메트릭 조회
// @Description Prometheus를 통해 집계자 VM의 실시간 시스템 메트릭을 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/system-metrics [get]
func (h *MLflowHandler) GetSystemMetrics(c *gin.Context) {
	// 사용자 인증 확인
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")

	// DB에서 aggregator 정보 조회
	aggregator, err := h.aggregatorRepo.GetAggregatorByID(aggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// Aggregator의 IP 주소 확인
	if aggregator.PublicIP == "" {
		log.Printf("GetSystemMetrics - PublicIP가 없음: %s", aggregatorID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "시스템 메트릭을 조회할 수 없습니다",
			"details": "Aggregator의 Public IP가 설정되지 않았습니다",
		})
		return
	}

	// Prometheus를 통해 실제 시스템 메트릭 조회
	if h.prometheusService != nil {
		vmInfo, err := h.prometheusService.GetVMMonitoringInfoWithIP(aggregator.PublicIP)
		if err != nil {
			log.Printf("GetSystemMetrics - Prometheus 조회 실패: %v", err)
			// Prometheus 조회 실패 시 오류 반환
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "시스템 메트릭을 조회할 수 없습니다",
				"details": "Prometheus 서버 연결 실패",
			})
			return
		}

		// 실제 Prometheus 데이터 반환
		c.JSON(http.StatusOK, gin.H{
			"cpu_usage":     vmInfo.CPUUsage,
			"memory_usage":  vmInfo.MemoryUsage,
			"disk_usage":    vmInfo.DiskUsage,
			"network_in":    vmInfo.NetworkInBytes,
			"network_out":   vmInfo.NetworkOutBytes,
			"last_updated":  vmInfo.LastUpdated.Format(time.RFC3339),
		})
	} else {
		// Prometheus 서비스가 없으면 서비스 불가 오류 반환
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "시스템 메트릭 서비스가 사용할 수 없습니다",
			"details": "Prometheus 서비스가 구성되지 않았습니다",
		})
	}
}

// GetMLflowInfo godoc
// @Summary MLflow 정보 조회
// @Description 집계자의 MLflow 실험 정보와 URL을 조회합니다.
// @Tags aggregators
// @Accept json
// @Produce json
// @Param id path string true "Aggregator ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/aggregators/{id}/mlflow-info [get]
func (h *MLflowHandler) GetMLflowInfo(c *gin.Context) {
	// 사용자 인증 확인
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증이 필요합니다"})
		return
	}

	aggregatorID := c.Param("id")

	// DB에서 aggregator 정보 조회
	aggregator, err := h.aggregatorRepo.GetAggregatorByIDWithFederatedLearning(aggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// MLflow URL 구성
	var mlflowURL string
	if aggregator.PublicIP != "" {
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregator.PublicIP)
	} else {
		// fallback으로 환경변수나 고정 IP 사용
		aggregatorIP := os.Getenv("AGGREGATOR_IP")
		if aggregatorIP == "" {
			aggregatorIP = "localhost" // 로컬 개발용
		}
		mlflowURL = fmt.Sprintf("http://%s:5000", aggregatorIP)
	}

	// 실험 이름 결정
	var experimentName string
	if aggregator.FederatedLearning != nil {
		experimentName = fmt.Sprintf("federated-learning-%s", aggregator.FederatedLearning.ID)
	} else {
		experimentName = fmt.Sprintf("federated-learning-%s", aggregatorID)
	}

	// MLflow UI URL 구성
	experimentURL := fmt.Sprintf("%s/#/experiments", mlflowURL)
	if aggregator.MLflowExperimentID != nil && *aggregator.MLflowExperimentID != "" {
		experimentURL = fmt.Sprintf("%s/#/experiments/%s", mlflowURL, *aggregator.MLflowExperimentID)
	}

	c.JSON(http.StatusOK, gin.H{
		"mlflow_url":        mlflowURL,
		"experiment_url":    experimentURL,
		"experiment_name":   experimentName,
		"experiment_id":     aggregator.MLflowExperimentID,
		"has_experiment":    aggregator.MLflowExperimentID != nil && *aggregator.MLflowExperimentID != "",
		"aggregator_ip":     aggregator.PublicIP,
		"mlflow_accessible": aggregator.PublicIP != "",
	})
}

// Helper 메서드들

// buildTrainingHistory는 MLflow run 데이터로부터 학습 히스토리를 생성합니다
func (h *MLflowHandler) buildTrainingHistory(stepData map[int]map[string]interface{}, runInfo *MLflowRun, aggregator *models.Aggregator) []TrainingRound {
	var trainingHistory []TrainingRound

	// 실제 참가자 수를 가져옵니다
	participantCount := h.getActualParticipantCount(aggregator)
	log.Printf("buildTrainingHistory - 실제 참가자 수: %d", participantCount)

	for step, metrics := range stepData {
		// 맵에서 안전하게 float 값을 가져오는 헬퍼 함수
		getFloat := func(key string) float64 {
			if val, ok := metrics[key].(float64); ok {
				return val
			}
			return 0.0
		}
		// 안전하게 타임스탬프 가져오기
		var roundTimestamp int64
		if ts, ok := metrics["timestamp"].(int64); ok {
			roundTimestamp = ts
		} else {
			roundTimestamp = runInfo.Info.StartTime // 실패 시 run의 시작 시간으로 대체
		}

		// 라운드 지속 시간을 계산합니다 (실제로는 각 라운드별 시간을 계산해야 하지만, 여기서는 추정값 사용)
		duration := h.calculateRoundDuration(step, stepData)

		round := TrainingRound{
			Round:             step,
			Accuracy:          getFloat("accuracy"),
			Loss:              getFloat("val_loss"),
			F1Score:           getFloat("f1_macro"),
			Precision:         getFloat("precision_macro"),
			Recall:            getFloat("recall_macro"),
			Duration:          duration,       // 계산된 지속 시간
			ParticipantsCount: participantCount, // 실제 참가자 수
			Timestamp:         time.Unix(roundTimestamp/1000, 0).Format(time.RFC3339),
		}
		if round.Loss == 0.0 {
			round.Loss = getFloat("train_loss")
		}
		trainingHistory = append(trainingHistory, round)
	}

	// 최종 결과를 라운드 번호로 정렬
	sort.Slice(trainingHistory, func(i, j int) bool {
		return trainingHistory[i].Round < trainingHistory[j].Round
	})

	return trainingHistory
}

// extractLatestMetrics는 최신 step의 메트릭을 추출합니다
func (h *MLflowHandler) extractLatestMetrics(metrics []struct {
	Key       string  `json:"key"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	Step      int     `json:"step"`
}) (map[string]float64, int) {
	latestMetrics := make(map[string]float64)
	maxStep := 0

	for _, metric := range metrics {
		if metric.Step > maxStep {
			maxStep = metric.Step
		}
	}

	for _, metric := range metrics {
		if metric.Step == maxStep {
			latestMetrics[metric.Key] = metric.Value
		}
	}

	return latestMetrics, maxStep
}

// mapRunStatus는 MLflow run 상태를 시스템 상태로 매핑합니다
func (h *MLflowHandler) mapRunStatus(status string) string {
	switch status {
	case "FINISHED":
		return "completed"
	case "FAILED":
		return "error"
	case "RUNNING":
		return "running"
	default:
		return "pending"
	}
}

// getMetricValue는 메트릭 값을 안전하게 가져옵니다
func getMetricValue(metrics map[string]float64, key string, defaultValue float64) float64 {
	if value, exists := metrics[key]; exists {
		return value
	}
	return defaultValue
}

// getActualParticipantCount는 실제 참가자 수를 조회합니다
func (h *MLflowHandler) getActualParticipantCount(aggregator *models.Aggregator) int {
	// 집계자에 연결된 연합학습이 있는 경우
	if aggregator.FederatedLearning != nil {
		// 데이터베이스에서 실제 참가자 수 조회
		// 여기서는 aggregator의 ParticipantCount 필드를 사용하거나
		// 별도 쿼리로 연합학습 참가자 수를 조회할 수 있습니다
		if aggregator.ParticipantCount > 0 {
			return aggregator.ParticipantCount
		}
		
		// FederatedLearning의 participant_count 사용
		if aggregator.FederatedLearning.ParticipantCount > 0 {
			return aggregator.FederatedLearning.ParticipantCount
		}
	}
	
	// aggregator 자체의 participant count 사용
	if aggregator.ParticipantCount > 0 {
		return aggregator.ParticipantCount
	}
	
	// 기본값 반환
	return 1
}

// calculateRoundDuration는 라운드 지속 시간을 계산합니다 (추정값)
func (h *MLflowHandler) calculateRoundDuration(step int, stepData map[int]map[string]interface{}) int {
	// 실제로는 각 라운드의 시작/종료 타임스탬프를 비교해야 하지만,
	// 여기서는 타임스탬프 기반 추정값을 계산합니다
	
	// 이전 스텝과 현재 스텝의 타임스탬프 차이를 계산
	if step > 1 {
		if prevStepData, exists := stepData[step-1]; exists {
			currentTs, currentOk := stepData[step]["timestamp"].(int64)
			prevTs, prevOk := prevStepData["timestamp"].(int64)
			
			if currentOk && prevOk && currentTs > prevTs {
				// 밀리초를 초로 변환
				durationSeconds := (currentTs - prevTs) / 1000
				if durationSeconds > 0 && durationSeconds < 3600 { // 1시간 이하인 경우만
					return int(durationSeconds)
				}
			}
		}
	}
	
	// 기본값 반환 (2분)
	return 120
}
