package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers/aggregator"
	"github.com/gin-gonic/gin"
)

func SetupAggregatorRoutes(authorized *gin.RouterGroup, aggregatorHandler *aggregator.AggregatorHandler) {
	aggregators := authorized.Group("/aggregators")
	{
		// Aggregator 배치 최적화
		aggregators.POST("/optimization", aggregatorHandler.OptimizeAggregatorPlacement)

		// Aggregator 통계 조회
		aggregators.GET("/stats", aggregatorHandler.GetAggregatorStats)

		// Aggregator 목록 조회
		aggregators.GET("", aggregatorHandler.GetAggregators)

		// 새 Aggregator 생성
		aggregators.POST("", aggregatorHandler.CreateAggregator)

		// 특정 Aggregator 조회
		aggregators.GET("/:id", aggregatorHandler.GetAggregator)

		// Aggregator 상태 업데이트
		aggregators.PUT("/:id/status", aggregatorHandler.UpdateAggregatorStatus)

		// Aggregator 메트릭 업데이트
		aggregators.PUT("/:id/metrics", aggregatorHandler.UpdateAggregatorMetrics)

		// Aggregator 학습 히스토리 조회
		aggregators.GET("/:id/training-history", aggregatorHandler.GetTrainingHistory)

		// Aggregator 삭제
		aggregators.DELETE("/:id", aggregatorHandler.DeleteAggregator)
	}
}
