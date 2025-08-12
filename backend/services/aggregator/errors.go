package aggregator

import "errors"

// Aggregator 서비스 관련 에러들
var (
	ErrAggregatorNotFound     = errors.New("aggregator not found or access denied")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrInvalidAggregatorID    = errors.New("invalid aggregator ID")
	ErrAggregatorCreateFailed = errors.New("failed to create aggregator")
	ErrAggregatorUpdateFailed = errors.New("failed to update aggregator")
	ErrAggregatorDeleteFailed = errors.New("failed to delete aggregator")
	ErrTerraformDeployFailed  = errors.New("terraform deployment failed")
	ErrInvalidMetricsData     = errors.New("invalid metrics data")
	ErrInvalidStatus          = errors.New("invalid status")
)