package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

// HTTP 미들웨어 - 메트릭 및 트레이싱
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		ctx, span := otel.Tracer("http").Start(r.Context(), r.URL.Path)
		defer span.End()
		
		// HTTP 요청에 트레이싱 컨텍스트 추가
		r = r.WithContext(ctx)
		
		// 응답 래핑하여 상태 코드 캡처
		wrapped := newResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		
		// 메트릭 기록
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(wrapped.statusCode)).Inc()
		
		// 트레이싱에 정보 추가
		span.SetAttributes(
			semconv.HTTPMethodKey.String(r.Method),
			semconv.HTTPURLKey.String(r.URL.String()),
			semconv.HTTPStatusCodeKey.Int(wrapped.statusCode),
		)
	})
}

// HTTP 응답 래퍼
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// 연합학습 라운드 시뮬레이션 함수 (예시)
func simulateFederatedLearningRound(ctx context.Context) {
	ctx, span := otel.Tracer("federated-learning").Start(ctx, "learning-round")
	defer span.End()

	start := time.Now()

	// 참여자 수 업데이트 (예시)
	participants := 5
	federatedLearningParticipants.Set(float64(participants))
	span.SetAttributes(attribute.Int("fl.participants", participants))

	// 학습 라운드 실행
	time.Sleep(2 * time.Second)

	// 라운드 완료
	federatedLearningRounds.Inc()
	duration := time.Since(start).Seconds()
	federatedLearningRoundDuration.Observe(duration)

	span.SetAttributes(
		attribute.String("fl.round_status", "completed"),
		attribute.Float64("fl.round_duration_seconds", duration),
	)
}

func main() {
	// Jaeger 트레이서 초기화
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	
	// 라우터 설정
	r := mux.NewRouter()
	
	// 메트릭 엔드포인트
	r.Handle("/metrics", promhttp.Handler())
	
	// API 엔드포인트
	api := r.PathPrefix("/api").Subrouter()
	api.Use(metricsMiddleware)
	
	// 실제 API 엔드포인트 정의
	api.HandleFunc("/start-learning", func(w http.ResponseWriter, r *http.Request) {
		simulateFederatedLearningRound(r.Context())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Learning round started"))
	}).Methods("POST")
	
	// 서버 시작
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}