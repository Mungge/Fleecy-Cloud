package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/gin-gonic/gin"
)

func SetupParticipantRoutes(authorized *gin.RouterGroup, participantHandler *handlers.ParticipantHandler) {
	participants := authorized.Group("/participants")
	{
		// 기본 CRUD 라우트
		participants.GET("", participantHandler.GetParticipants)
		participants.GET("/available", participantHandler.GetAvailableParticipants)
		participants.POST("", participantHandler.CreateParticipant)
		participants.GET("/:id", participantHandler.GetParticipant)
		participants.PUT("/:id", participantHandler.UpdateParticipant)
		participants.DELETE("/:id", participantHandler.DeleteParticipant)
		
		// 헬스체크 라우트
		participants.POST("/:id/health-check", participantHandler.HealthCheckParticipant)
	}
}
