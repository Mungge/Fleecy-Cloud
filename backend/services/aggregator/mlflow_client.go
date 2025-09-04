// services/aggregator/mlflow_client.go
package aggregator

import (
   "bytes"
   "encoding/json"
   "fmt"
   "net/http"
   "time"
)

// MLflowClient MLflow API 클라이언트
type MLflowClient struct {
   BaseURL    string
   HTTPClient *http.Client
}

// NewMLflowClient 새 MLflowClient 생성
func NewMLflowClient(baseURL string) *MLflowClient {
   return &MLflowClient{
   	BaseURL: baseURL,
   	HTTPClient: &http.Client{
   		Timeout: 30 * time.Second,
   	},
   }
}

// ExperimentCreateRequest 실험 생성 요청
type ExperimentCreateRequest struct {
   Name string `json:"name"`
   ArtifactLocation string `json:"artifact_location,omitempty"`
   Tags []ExperimentTag `json:"tags,omitempty"`
}

// ExperimentCreateResponse 실험 생성 응답
type ExperimentCreateResponse struct {
   ExperimentID string `json:"experiment_id"`
}

// ExperimentTag 실험 태그
type ExperimentTag struct {
   Key   string `json:"key"`
   Value string `json:"value"`
}

// RunCreateRequest 실행 생성 요청
type RunCreateRequest struct {
   ExperimentID string    `json:"experiment_id"`
   UserID       string    `json:"user_id,omitempty"`
   StartTime    int64     `json:"start_time,omitempty"`
   Tags         []RunTag  `json:"tags,omitempty"`
}

// RunCreateResponse 실행 생성 응답
type RunCreateResponse struct {
   Run RunInfo `json:"run"`
}

// RunInfo 실행 정보
type RunInfo struct {
   Info RunData `json:"info"`
}

// RunData 실행 데이터
type RunData struct {
   RunID        string `json:"run_id"`
   RunUUID      string `json:"run_uuid"`
   ExperimentID string `json:"experiment_id"`
   UserID       string `json:"user_id"`
   Status       string `json:"status"`
   StartTime    int64  `json:"start_time"`
   EndTime      int64  `json:"end_time"`
   ArtifactURI  string `json:"artifact_uri"`
   LifecycleStage string `json:"lifecycle_stage"`
}

// RunTag 실행 태그
type RunTag struct {
   Key   string `json:"key"`
   Value string `json:"value"`
}

// MetricRequest 메트릭 로그 요청
type MetricRequest struct {
   RunID     string `json:"run_id"`
   Key       string `json:"key"`
   Value     float64 `json:"value"`
   Timestamp int64  `json:"timestamp"`
   Step      int64  `json:"step,omitempty"`
}

// ParamRequest 파라미터 로그 요청
type ParamRequest struct {
   RunID string `json:"run_id"`
   Key   string `json:"key"`
   Value string `json:"value"`
}

// CreateExperiment MLflow 실험을 생성합니다
func (c *MLflowClient) CreateExperiment(name string, artifactLocation string, tags []ExperimentTag) (*ExperimentCreateResponse, error) {
   request := ExperimentCreateRequest{
   	Name:             name,
   	ArtifactLocation: artifactLocation,
   	Tags:             tags,
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return nil, fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/experiments/create", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return nil, fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return nil, fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   var response ExperimentCreateResponse
   if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
   	return nil, fmt.Errorf("failed to decode response: %w", err)
   }

   return &response, nil
}

// GetExperiment 실험 정보를 가져옵니다
func (c *MLflowClient) GetExperiment(experimentID string) (*ExperimentCreateResponse, error) {
   url := fmt.Sprintf("%s/api/2.0/mlflow/experiments/get?experiment_id=%s", c.BaseURL, experimentID)
   
   resp, err := c.HTTPClient.Get(url)
   if err != nil {
   	return nil, fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return nil, fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   var response struct {
   	Experiment struct {
   		ExperimentID string `json:"experiment_id"`
   		Name         string `json:"name"`
   	} `json:"experiment"`
   }

   if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
   	return nil, fmt.Errorf("failed to decode response: %w", err)
   }

   return &ExperimentCreateResponse{
   	ExperimentID: response.Experiment.ExperimentID,
   }, nil
}

// CreateRun 새 실행을 생성합니다
func (c *MLflowClient) CreateRun(experimentID, userID string, tags []RunTag) (*RunCreateResponse, error) {
   request := RunCreateRequest{
   	ExperimentID: experimentID,
   	UserID:       userID,
   	StartTime:    time.Now().UnixMilli(),
   	Tags:         tags,
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return nil, fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/runs/create", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return nil, fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return nil, fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   var response RunCreateResponse
   if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
   	return nil, fmt.Errorf("failed to decode response: %w", err)
   }

   return &response, nil
}

// LogMetric 메트릭을 로그합니다
func (c *MLflowClient) LogMetric(runID, key string, value float64, timestamp int64, step int64) error {
   request := MetricRequest{
   	RunID:     runID,
   	Key:       key,
   	Value:     value,
   	Timestamp: timestamp,
   	Step:      step,
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/runs/log-metric", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   return nil
}

// LogParam 파라미터를 로그합니다
func (c *MLflowClient) LogParam(runID, key, value string) error {
   request := ParamRequest{
   	RunID: runID,
   	Key:   key,
   	Value: value,
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/runs/log-parameter", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   return nil
}

// UpdateRun 실행 상태를 업데이트합니다
func (c *MLflowClient) UpdateRun(runID, status string) error {
   request := map[string]interface{}{
   	"run_id": runID,
   	"status": status,
   	"end_time": time.Now().UnixMilli(),
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/runs/update", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   return nil
}

// LogBatch 여러 메트릭/파라미터를 한번에 로그합니다
func (c *MLflowClient) LogBatch(runID string, metrics []MetricRequest, params []ParamRequest) error {
   request := map[string]interface{}{
   	"run_id": runID,
   	"metrics": metrics,
   	"params": params,
   }

   payload, err := json.Marshal(request)
   if err != nil {
   	return fmt.Errorf("failed to marshal request: %w", err)
   }

   url := fmt.Sprintf("%s/api/2.0/mlflow/runs/log-batch", c.BaseURL)
   resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(payload))
   if err != nil {
   	return fmt.Errorf("failed to make request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return fmt.Errorf("MLflow API returned status %d", resp.StatusCode)
   }

   return nil
}

// HealthCheck MLflow 서버 상태를 확인합니다
func (c *MLflowClient) HealthCheck() error {
   url := fmt.Sprintf("%s/health", c.BaseURL)
   
   resp, err := c.HTTPClient.Get(url)
   if err != nil {
   	return fmt.Errorf("failed to connect to MLflow server: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
   	return fmt.Errorf("MLflow server health check failed with status %d", resp.StatusCode)
   }

   return nil
}