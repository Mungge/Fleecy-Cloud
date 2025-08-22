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

		// 특정 연합학습 작업 업데이트
		federated.PUT("/:id", federatedLearningHandler.UpdateFederatedLearning)

		// 특정 연합학습 작업 삭제
		federated.DELETE("/:id", federatedLearningHandler.DeleteFederatedLearning)
	}
}
