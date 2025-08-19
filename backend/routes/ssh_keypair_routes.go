package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupSSHKeypairRoutes(router *gin.RouterGroup, handler *handlers.SSHKeypairHandler) {
	keypairRoutes := router.Group("/keypairs")
	{
		// 집계자별 키페어 관리
		keypairRoutes.GET("/aggregator/:aggregatorId", handler.GetKeypairByAggregatorID)
		keypairRoutes.GET("/aggregator/:aggregatorId/private-key", handler.DownloadPrivateKey)
		keypairRoutes.DELETE("/aggregator/:aggregatorId", handler.DeleteKeypairByAggregatorID)

		// 사용자별 키페어 목록
		keypairRoutes.GET("/user/:userId", handler.ListKeypairsByUser)
	}
}
