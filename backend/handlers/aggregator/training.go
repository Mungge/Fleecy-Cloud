package aggregator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Mungge/Fleecy-Cloud/repository"
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
	mlflowClient   *MLflowClient
	aggregatorRepo *repository.AggregatorRepository
}

func NewMLflowHandler(mlflowURL string, aggregatorRepo *repository.AggregatorRepository) *MLflowHandler {
	return &MLflowHandler{
		mlflowClient:   NewMLflowClient(mlflowURL),
		aggregatorRepo: aggregatorRepo,
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
	aggregator, err := h.aggregatorRepo.GetAggregatorByID(aggregatorID)
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

	// MLflow 실험명 결정
	var experimentName string
	if aggregator.MLflowExperimentName != nil && *aggregator.MLflowExperimentName != "" {
		experimentName = *aggregator.MLflowExperimentName
	} else {
		// 기본 실험명 사용 (레거시 지원)
		experimentName = "flower-demo"
	}

	// MLflow에서 실험 조회
	experiments, err := h.mlflowClient.GetExperiments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "MLflow 실험 조회 실패",
			"details": err.Error(),
		})
		return
	}

	// 해당 실험 찾기
	var experimentID string
	for _, exp := range experiments {
		if exp.Name == experimentName {
			experimentID = exp.ExperimentID
			break
		}
	}

	if experimentID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("실험 '%s'을 찾을 수 없습니다", experimentName),
		})
		return
	}

	// runs 조회
	runs, err := h.mlflowClient.GetRuns(experimentID)
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
		history, err := h.mlflowClient.GetMetricHistory(latestRun.Info.RunID, key)
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
	trainingHistory := h.buildTrainingHistory(stepData, latestRun)

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
	aggregator, err := h.aggregatorRepo.GetAggregatorByID(aggregatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregator 조회 실패"})
		return
	}

	if aggregator == nil || aggregator.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregator를 찾을 수 없습니다"})
		return
	}

	// MLflow 실험명 결정
	var experimentName string
	if aggregator.MLflowExperimentName != nil && *aggregator.MLflowExperimentName != "" {
		experimentName = *aggregator.MLflowExperimentName
	} else {
		experimentName = "flower-demo" // 기본값
	}

	// 실험 및 runs 조회
	experiments, err := h.mlflowClient.GetExperiments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MLflow 연결 실패"})
		return
	}

	var experimentID string
	for _, exp := range experiments {
		if exp.Name == experimentName {
			experimentID = exp.ExperimentID
			break
		}
	}

	if experimentID == "" {
		// 실험이 없으면 기본값 반환
		c.JSON(http.StatusOK, gin.H{
			"accuracy":     0,
			"loss":         0,
			"currentRound": 0,
			"status":       "pending",
			"f1_score":     0,
			"precision":    0,
			"recall":       0,
		})
		return
	}

	runs, err := h.mlflowClient.GetRuns(experimentID)
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
	})
}

// Helper 메서드들

// buildTrainingHistory는 MLflow run 데이터로부터 학습 히스토리를 생성합니다
func (h *MLflowHandler) buildTrainingHistory(stepData map[int]map[string]interface{}, runInfo *MLflowRun) []TrainingRound {
	var trainingHistory []TrainingRound

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

		round := TrainingRound{
			Round:             step,
			Accuracy:          getFloat("accuracy"),
			Loss:              getFloat("val_loss"),
			F1Score:           getFloat("f1_macro"),
			Precision:         getFloat("precision_macro"),
			Recall:            getFloat("recall_macro"),
			Duration:          120, // 기본값
			ParticipantsCount: 3,   // 기본값
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
