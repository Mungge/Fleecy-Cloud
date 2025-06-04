package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupCloudRoutes(authorized *gin.RouterGroup, cloudHandler *handlers.CloudHandler) {
	clouds := authorized.Group("/clouds")
	{
		clouds.GET("", cloudHandler.GetClouds)
		clouds.POST("", cloudHandler.AddCloud)
		clouds.DELETE("/:id", cloudHandler.DeleteCloud)
		clouds.POST("/upload", cloudHandler.UploadCloudCredential)
		clouds.GET("/:id/test", cloudHandler.TestCloudConnectionWithDetails)
	}
}
