package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers/aggregator"
	"github.com/gin-gonic/gin"
)

func SetupAggregatorRoutes(authorized *gin.RouterGroup, aggregatorHandler *aggregator.AggregatorHandler, mlflowHandler *aggregator.MLflowHandler) {
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

		// MLflow 라우트들
        aggregators.GET("/:id/training-history", mlflowHandler.GetTrainingHistory)
        aggregators.GET("/:id/realtime-metrics", mlflowHandler.GetRealTimeMetrics)
        aggregators.GET("/:id/system-metrics", mlflowHandler.GetSystemMetrics)
        aggregators.GET("/:id/mlflow-info", mlflowHandler.GetMLflowInfo)

		// Aggregator 삭제
		aggregators.DELETE("/:id", aggregatorHandler.DeleteAggregator)
	}
}
