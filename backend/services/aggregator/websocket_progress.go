package aggregator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ProgressMessage 진행 상황 메시지
type ProgressMessage struct {
	AggregatorID string `json:"aggregator_id"`
	Stage        int    `json:"stage"`        // 1-5
	TotalStages  int    `json:"total_stages"` // 5
	Message      string `json:"message"`
	Status       string `json:"status"` // "progress", "success", "error"
	Error        string `json:"error,omitempty"`
	Timestamp    int64  `json:"timestamp"`
}

// SSEProgressTracker Server-Sent Events를 통한 진행 상황 추적
type SSEProgressTracker struct {
	connections map[string]http.ResponseWriter
	mutex       sync.RWMutex
}

// NewWebSocketProgressTracker 새 SSE 진행 상황 추적기 생성
func NewWebSocketProgressTracker() *SSEProgressTracker {
	return &SSEProgressTracker{
		connections: make(map[string]http.ResponseWriter),
	}
}

// HandleWebSocket SSE 연결 처리
func (tracker *SSEProgressTracker) HandleWebSocket(w http.ResponseWriter, r *http.Request, aggregatorID string) {
	// SSE 헤더 설정
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	tracker.mutex.Lock()
	tracker.connections[aggregatorID] = w
	tracker.mutex.Unlock()

	log.Printf("SSE connected for aggregator: %s", aggregatorID)

	// 연결 유지 (컨텍스트가 취소될 때까지)
	<-r.Context().Done()
	log.Printf("SSE connection closed for aggregator %s", aggregatorID)

	// 연결 정리
	tracker.mutex.Lock()
	delete(tracker.connections, aggregatorID)
	tracker.mutex.Unlock()
}

// SendProgress 진행 상황 전송
func (tracker *SSEProgressTracker) SendProgress(aggregatorID string, stage int, message string) {
	tracker.sendMessage(ProgressMessage{
		AggregatorID: aggregatorID,
		Stage:        stage,
		TotalStages:  5,
		Message:      message,
		Status:       "progress",
		Timestamp:    getCurrentTimestamp(),
	})
}

// SendSuccess 성공 메시지 전송
func (tracker *SSEProgressTracker) SendSuccess(aggregatorID string, message string) {
	tracker.sendMessage(ProgressMessage{
		AggregatorID: aggregatorID,
		Stage:        5,
		TotalStages:  5,
		Message:      message,
		Status:       "success",
		Timestamp:    getCurrentTimestamp(),
	})
}

// SendError 에러 메시지 전송
func (tracker *SSEProgressTracker) SendError(aggregatorID string, stage int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	
	tracker.sendMessage(ProgressMessage{
		AggregatorID: aggregatorID,
		Stage:        stage,
		TotalStages:  5,
		Message:      message,
		Status:       "error",
		Error:        errorMsg,
		Timestamp:    getCurrentTimestamp(),
	})
}

// sendMessage 메시지 전송 (내부 메서드)
func (tracker *SSEProgressTracker) sendMessage(msg ProgressMessage) {
	tracker.mutex.RLock()
	w, exists := tracker.connections[msg.AggregatorID]
	tracker.mutex.RUnlock()

	if !exists {
		return // 연결이 없으면 무시
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal progress message: %v", err)
		return
	}

	// SSE 형식으로 전송
	sseData := fmt.Sprintf("data: %s\n\n", string(data))
	if _, err := fmt.Fprint(w, sseData); err != nil {
		log.Printf("Failed to send SSE message: %v", err)
		// 연결이 끊어진 경우 정리
		tracker.mutex.Lock()
		delete(tracker.connections, msg.AggregatorID)
		tracker.mutex.Unlock()
		return
	}

	// 응답 플러시
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix() * 1000 // 밀리초 단위
}
