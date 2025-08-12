package initialization

import (
	"context"
	"log"
	"os"

	"github.com/Mungge/Fleecy-Cloud/config"
	aggregatorhandler "github.com/Mungge/Fleecy-Cloud/handlers/aggregator"
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/services"
	aggregatorservice "github.com/Mungge/Fleecy-Cloud/services/aggregator"
	"github.com/joho/godotenv"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Repositories는 애플리케이션에서 사용되는 모든 리포지토리를 포함합니다
type Repositories struct {
	UserRepo         *repository.UserRepository
	RefreshTokenRepo *repository.RefreshTokenRepository
	CloudRepo        *repository.CloudRepository
	FLRepo           *repository.FederatedLearningRepository
	ParticipantRepo  *repository.ParticipantRepository
	AggregatorRepo   *repository.AggregatorRepository
	VMRepo           *repository.VirtualMachineRepository
}

// Dependencies는 애플리케이션의 모든 의존성을 관리합니다
type Dependencies struct {
	// Repositories
	Repositories *Repositories
	
	// Aggregator Services
	AggregatorService   *aggregatorservice.AggregatorService
	MetricsService      *aggregatorservice.AggregatorMetricsService
	TrainingService     *aggregatorservice.AggregatorTrainingService
	OptimizationService aggregatorservice.OptimizationService 
	
	// Aggregator Handler
	AggregatorHandler *aggregatorhandler.AggregatorHandler
}

// InitializeApplication은 애플리케이션의 모든 초기화 작업을 수행합니다
func InitializeApplication() error {
	log.Println("애플리케이션 초기화 시작...")

	// 0. 환경 변수 로드(.env)
	loadDotEnv()

	// 1. 데이터베이스 연결
	if err := config.ConnectDatabase(); err != nil {
		return err
	}

	// 2. 데이터베이스 마이그레이션
	if err := RunDatabaseMigration(); err != nil {
		return err
	}

	// 3. 초기 데이터 로드
	if err := LoadInitialData(); err != nil {
		log.Printf("초기 데이터 로드 실패: %v", err)
	}

	log.Println("애플리케이션 초기화 완료!")
	return nil
}

// RunDatabaseMigration은 데이터베이스 마이그레이션을 실행합니다
func RunDatabaseMigration() error {
	log.Println("데이터베이스 마이그레이션 시작...")
	
	db := config.GetDB()
	err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.CloudConnection{},
		&models.CloudPrice{},
		&models.CloudLatency{},
		&models.FederatedLearning{},
		&models.Participant{},
		&models.VirtualMachine{},
		&models.ParticipantFederatedLearning{},
		&models.Aggregator{},
		&models.TrainingRound{},
	)
	if err != nil {
		return err
	}
	
	log.Println("데이터베이스 마이그레이션 완료")
	return nil
}

// LoadInitialData는 Asset 파일로부터 초기 데이터를 로드합니다
func LoadInitialData() error {
	log.Println("초기 데이터 로드 시작...")
	
	db := config.GetDB()
	if err := services.InitializeDataFromAssets(db); err != nil {
		return err
	}
	
	log.Println("초기 데이터 로드 완료")
	return nil
}

// loadDotEnv는 .env 파일을 찾아 로드합니다. (백엔드 디렉토리/루트 등 공통 위치 검색)
func loadDotEnv() {
	candidates := []string{
		".env",              // 현재 작업 디렉토리
		"../.env",           // 백엔드 하위에서 실행 시 루트의 .env
		"../../.env",        // 더 상위
		"backend/.env",      // 루트에서 백엔드 하위
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("환경 변수 로드 완료: %s", p)
				return
			}
		}
	}
	log.Printf(".env 파일을 찾지 못했습니다. 시스템 환경 변수를 사용합니다.")
}

// InitializeRepositories는 모든 리포지토리를 초기화합니다
func InitializeRepositories() *Repositories {
	log.Println("리포지토리 초기화 시작...")
	
	db := config.GetDB()
	
	repos := &Repositories{
		UserRepo:         repository.NewUserRepository(db),
		RefreshTokenRepo: repository.NewRefreshTokenRepository(db),
		CloudRepo:        repository.NewCloudRepository(db),
		FLRepo:           repository.NewFederatedLearningRepository(db),
		ParticipantRepo:  repository.NewParticipantRepository(db),
		AggregatorRepo:   repository.NewAggregatorRepository(db),
		VMRepo:           repository.NewVirtualMachineRepository(db),
	}
	
	log.Println("리포지토리 초기화 완료")
	return repos
}

// NewDependencies는 모든 의존성을 초기화하고 주입합니다 
func NewDependencies() *Dependencies {
	log.Println("Aggregator 의존성 초기화 시작...")
	
	// Repository 초기화
	repos := InitializeRepositories()
	
	// Aggregator Service 초기화 (새로운 구조)
	aggregatorService := aggregatorservice.NewAggregatorService(repos.AggregatorRepo, repos.FLRepo)
	metricsService := aggregatorservice.NewAggregatorMetricsService(repos.AggregatorRepo)
	trainingService := aggregatorservice.NewAggregatorTrainingService(repos.AggregatorRepo)
	
	// OptimizationService 어댑터 사용
	originalOptimizationService := services.NewOptimizationService()
	optimizationService := aggregatorservice.NewOptimizationServiceAdapter(originalOptimizationService)
	
	// Aggregator Handler 초기화 (새로운 구조)
	aggregatorHandler := aggregatorhandler.NewAggregatorHandler(
		aggregatorService,
		metricsService,
		trainingService,
		optimizationService,
	)
	
	log.Println("Aggregator 의존성 초기화 완료")
	
	return &Dependencies{
		Repositories:        repos,
		AggregatorService:   aggregatorService,
		MetricsService:      metricsService,
		TrainingService:     trainingService,
		OptimizationService: optimizationService,
		AggregatorHandler:   aggregatorHandler,
	}
}

// ShutdownTracer는 트레이서를 안전하게 종료합니다
func ShutdownTracer(tp *sdktrace.TracerProvider) {
	if err := tp.Shutdown(context.Background()); err != nil {
		log.Printf("트레이서 종료 실패: %v", err)
	}
}