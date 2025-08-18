package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
)

func SetupVirtualMachineRoutes(r *gin.Engine, vmRepo *repository.VirtualMachineRepository, participantRepo *repository.ParticipantRepository) {
	vmHandler := handlers.NewVirtualMachineHandler(vmRepo, participantRepo)

	// VM 관리 라우트
	vmRoutes := r.Group("/api/participants/:id/vms")
	vmRoutes.Use(middlewares.AuthMiddleware())
	{
		// 기본 CRUD 작업
		vmRoutes.GET("/", vmHandler.GetVirtualMachines)
		vmRoutes.GET("/:vmId", vmHandler.GetVirtualMachine)

		// VM 모니터링
		vmRoutes.GET("/:vmId/monitor", vmHandler.MonitorVirtualMachine)

		// VM 작업 관리
		vmRoutes.POST("/:vmId/assign-task", vmHandler.AssignTaskToVM)

		// VM 통계 및 상태별 조회
		vmRoutes.GET("/stats", vmHandler.GetVMStats)
		vmRoutes.GET("/available", vmHandler.GetAvailableVMs)
		vmRoutes.GET("/busy", vmHandler.GetBusyVMs)

		// Participant 소유의 모든 VM 조회
		vmRoutes.GET("/all", vmHandler.GetVMRequests)

		// VM 선택 (연합학습용)
		vmRoutes.POST("/select", vmHandler.SelectOptimalVM)

		// VM 사용률 조회 (모니터링용)
		vmRoutes.GET("/utilizations", vmHandler.GetVMUtilizations)

		// 라운드로빈 초기화
		vmRoutes.POST("/reset-round-robin", vmHandler.ResetVMSelectionRoundRobin)
	}
}
