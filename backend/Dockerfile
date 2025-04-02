FROM golang:1.23-alpine AS builder

WORKDIR /app

# 의존성 먼저 복사 및 다운로드 (캐싱 활용)
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -o backend .

# 최종 이미지 생성
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 빌드된 바이너리만 복사
COPY --from=builder /app/backend .

# 포트 노출
EXPOSE 8080

# 실행
CMD ["./backend"]