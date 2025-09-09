package aggregator

// Request/Response structures

// UpdateStatusRequest 상태 업데이트 요청
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateMetricsRequest 메트릭 업데이트 요청
type UpdateMetricsRequest struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	NetworkUsage float64 `json:"network_usage"`
}

// CreateAggregatorRequest Aggregator 생성 요청
type CreateAggregatorRequest struct {
	Name          string  `json:"name" binding:"required"`
	Algorithm     string  `json:"algorithm" binding:"required"`
	Region        string  `json:"region" binding:"required"`
	Storage       string  `json:"storage" binding:"required"`
	InstanceType  string  `json:"instanceType" binding:"required"`
	CloudProvider string  `json:"cloudProvider" binding:"required,oneof=aws gcp"`
	EstimatedCost string  `json:"estimatedCost" binding:"required"`
}

// CreateAggregatorResponse Aggregator 생성 응답
type CreateAggregatorResponse struct {
	AggregatorID    string `json:"aggregatorId"`
	Status          string `json:"status"`
	TerraformStatus string `json:"terraformStatus,omitempty"`
}