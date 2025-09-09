package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupFederatedLearningRoutes(authorized *gin.RouterGroup, federatedLearningHandler *handlers.FederatedLearningHandler) {
	federated := authorized.Group("/federated-learning")
	{
		// 연합학습 생성 (특정 Aggregator에)
		federated.POST("", federatedLearningHandler.CreateFederatedLearning)

		// 연합학습 목록 조회
		federated.GET("", federatedLearningHandler.GetFederatedLearnings)

		// 특정 연합학습 작업 조회
		federated.GET("/:id", federatedLearningHandler.GetFederatedLearning)

		// 특정 연합학습 작업의 로그 조회
		federated.GET("/:id/logs", federatedLearningHandler.GetFederatedLearningLogs)

		// 특정 연합학습 작업의 로그 스트리밍
		federated.GET("/:id/logs/stream", federatedLearningHandler.StreamFederatedLearningLogs)

		// 특정 연합학습 작업의 MLflow 대시보드 URL 조회
		federated.GET("/:id/mlflow", federatedLearningHandler.GetMLflowDashboardURL)

		// 특정 연합학습 작업의 MLflow 메트릭 조회
		federated.GET("/:id/metrics", federatedLearningHandler.GetMLflowMetrics)

		// 특정 연합학습 작업의 최신 메트릭 조회 (폴링용)
		federated.GET("/:id/metrics/latest", federatedLearningHandler.GetLatestMetrics)

		// 특정 연합학습 작업의 MLflow 메트릭을 DB에 동기화
		federated.POST("/:id/metrics/sync", federatedLearningHandler.SyncMLflowMetricsToDatabase)

		// 특정 연합학습 작업의 저장된 학습 히스토리 조회 (DB에서)
		federated.GET("/:id/training-history", federatedLearningHandler.GetStoredTrainingHistory)

		// 특정 연합학습 작업 업데이트
		federated.PUT("/:id", federatedLearningHandler.UpdateFederatedLearning)

		// 특정 연합학습 작업 삭제
		federated.DELETE("/:id", federatedLearningHandler.DeleteFederatedLearning)

		// 최적의 글로벌 모델 다운로드
		federated.GET("/:id/models/round/:round/download/:filename", federatedLearningHandler.DownloadGlobalModel)
	}
}
