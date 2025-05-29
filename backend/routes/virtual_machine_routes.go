package routes

import (
	"github.com/Mungge/Fleecy-Cloud/handlers"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/gin-gonic/gin"
)

// SetupVirtualMachineRoutes는 VM 관련 라우트를 설정합니다
func SetupVirtualMachineRoutes(r *gin.Engine, vmRepo *repository.VirtualMachineRepository, participantRepo *repository.ParticipantRepository) {
	vmHandler := handlers.NewVirtualMachineHandler(vmRepo, participantRepo)

	// VM 관리 라우트 (인증 필요)
	vmRoutes := r.Group("/api/participants/:id/vms")
	vmRoutes.Use(middlewares.AuthMiddleware())
	{
		// 기본 CRUD 작업
		vmRoutes.POST("/", vmHandler.CreateVirtualMachine)
		vmRoutes.GET("/", vmHandler.GetVirtualMachines)
		vmRoutes.GET("/:vmId", vmHandler.GetVirtualMachine)
		vmRoutes.DELETE("/:vmId", vmHandler.DeleteVirtualMachine)

		// VM 모니터링
		vmRoutes.GET("/:vmId/monitor", vmHandler.MonitorVirtualMachine)
		vmRoutes.POST("/:vmId/health-check", vmHandler.HealthCheckVM)

		// VM 전원 제어
		vmRoutes.POST("/:vmId/power", vmHandler.VMPowerAction)

		// VM 작업 관리
		vmRoutes.POST("/:vmId/assign-task", vmHandler.AssignTaskToVM)
		vmRoutes.POST("/:vmId/complete-task", vmHandler.CompleteVMTask)

		// VM 통계 및 상태별 조회
		vmRoutes.GET("/stats", vmHandler.GetVMStats)
		vmRoutes.GET("/available", vmHandler.GetAvailableVMs)
		vmRoutes.GET("/busy", vmHandler.GetBusyVMs)
	}
}
