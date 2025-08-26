package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
)

func SetupVirtualMachineRoutes(r *gin.Engine, participantRepo *repository.ParticipantRepository) {
	vmHandler := handlers.NewVirtualMachineHandler(participantRepo)

	// VM 관리 라우트
	vmRoutes := r.Group("/api/participants/:id/vms")
	vmRoutes.Use(middlewares.AuthMiddleware())
	{
		// OpenStack VM 직접 조회
		vmRoutes.GET("/all", vmHandler.GetVMRequests)

		// VM 통계 (실시간)
		vmRoutes.GET("/stats", vmHandler.GetVMStats)

		// VM 선택 (연합학습용)
		vmRoutes.POST("/select", vmHandler.SelectOptimalVM)

		// VM 사용률 조회 (모니터링용)
		vmRoutes.GET("/utilizations", vmHandler.GetVMUtilizations)

		// 라운드로빈 초기화
		vmRoutes.POST("/reset-round-robin", vmHandler.ResetVMSelectionRoundRobin)
	}
}
