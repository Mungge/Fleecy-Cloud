package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Mungge/Fleecy-Cloud/config"
	"github.com/Mungge/Fleecy-Cloud/handlers"
	authHandlers "github.com/Mungge/Fleecy-Cloud/handlers/auth"
	"github.com/Mungge/Fleecy-Cloud/middlewares"
	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/Mungge/Fleecy-Cloud/repository"
	"github.com/Mungge/Fleecy-Cloud/routes"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Prometheus 메트릭 정의
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 연합학습 관련 메트릭
	federatedLearningRounds = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "federated_learning_rounds_total",
			Help: "Total number of federated learning rounds",
		},
	)

	federatedLearningParticipants = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "federated_learning_participants",
			Help: "Number of current active federated learning participants",
		},
	)

	federatedLearningRoundDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "federated_learning_round_duration_seconds",
			Help:    "Duration of federated learning rounds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
	)
)

func init() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	// Prometheus 메트릭 등록
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(federatedLearningRounds)
	prometheus.MustRegister(federatedLearningParticipants)
	prometheus.MustRegister(federatedLearningRoundDuration)
}

// Jaeger 설정
func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// OTLP exporter 생성 (Jaeger가 OTLP HTTP 포트로 수신)
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("jaeger:4318"), // docker-compose 내부 통신
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("federated-learning-backend"),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// Gin 메트릭 미들웨어
func ginMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		ctx, span := otel.Tracer("http").Start(c.Request.Context(), c.Request.URL.Path)
		defer span.End()
		
		// 요청에 트레이싱 컨텍스트 추가
		c.Request = c.Request.WithContext(ctx)
		
		// 다음 핸들러 실행
		c.Next()
		
		// 메트릭 기록
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, http.StatusText(c.Writer.Status())).Inc()
		
		// 트레이싱에 정보 추가
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
		)
	}
}

func main() {
	// Jaeger 트레이서 초기화
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Jaeger 트레이서 초기화 실패: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("트레이서 종료 실패: %v", err)
		}
	}()

	// 데이터베이스 연결
	if err := config.ConnectDatabase(); err != nil {
		log.Fatalf("데이터베이스 연결 실패: %v", err)
	}
	db := config.GetDB()

	err = db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.CloudConnection{},
		&models.FederatedLearning{},
		&models.Participant{},
		&models.VirtualMachine{},
		&models.ParticipantFederatedLearning{},
		&models.Aggregator{},
		&models.TrainingRound{},
	)
	if err != nil {
		log.Fatalf("데이터베이스 마이그레이션 실패: %v", err)
	}

	// 리포지토리 초기화
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)  // 새로 추가
	cloudRepo := repository.NewCloudRepository(db)
	flRepo := repository.NewFederatedLearningRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	aggregatorRepo := repository.NewAggregatorRepository(db)
	vmRepo := repository.NewVirtualMachineRepository(db)

	// 핸들러 초기화
	authHandler := authHandlers.NewAuthHandler(
		userRepo,
		refreshTokenRepo,  // 새로 추가
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
	)
	cloudHandler := handlers.NewCloudHandler(cloudRepo)
	flHandler := handlers.NewFederatedLearningHandler(flRepo)
	participantHandler := handlers.NewParticipantHandler(participantRepo)
	aggregatorHandler := handlers.NewAggregatorHandler(aggregatorRepo, flRepo)

	// Gin 라우터 설정
	r := gin.Default()

	// CORS 설정
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 메트릭 미들웨어 적용
	r.Use(ginMetricsMiddleware())

	// 라우트 설정 - 각 도메인별로 개별 설정
	// 인증 라우트 (인증 미들웨어 없음)
	routes.SetupAuthRoutes(r, authHandler)

	// 인증이 필요한 라우트 그룹
	authorized := r.Group("/api")
	authorized.Use(middlewares.AuthMiddleware())

	// 각 도메인별 라우트 설정
	routes.SetupCloudRoutes(authorized, cloudHandler)
	routes.SetupParticipantRoutes(authorized, participantHandler)
	routes.SetupFederatedLearningRoutes(authorized, flHandler)
	routes.SetupAggregatorRoutes(authorized, aggregatorHandler)
	
	// VM 라우트 설정 (전체 엔진에 설정, 인증은 내부에서 처리)
	routes.SetupVirtualMachineRoutes(r, vmRepo, participantRepo)

	// 서버 시작
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}