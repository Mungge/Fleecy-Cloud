package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	// 환경 변수에서 데이터베이스 연결 정보 가져오기
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "kimminkyoung")
	dbPass := getEnv("DB_PASSWORD", "161106")
	dbName := getEnv("DB_NAME", "fleecy_cloud")

	// 데이터베이스 연결 문자열 생성
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// 데이터베이스 연결
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %v", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 테스트 실패: %v", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
} 