#!/bin/bash
set -e

# 1. OS 업데이트 및 필수 패키지 설치
apt-get update -y
apt-get install -y apt-transport-https ca-certificates curl software-properties-common wget

# 2. Docker 설치
if ! command -v docker > /dev/null 2>&1; then
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io
fi

# 3. Docker Compose 설치 (최신 버전)
if ! command -v docker-compose > /dev/null 2>&1; then
  DOCKER_COMPOSE_VERSION="2.21.0"
  curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  chmod +x /usr/local/bin/docker-compose
fi

# 4. Docker 서비스 활성화 및 시작
systemctl enable docker
systemctl start docker

# 5. 작업 디렉토리 생성
mkdir -p /opt/monitoring/prometheus
mkdir -p /opt/monitoring/grafana/provisioning/datasources
mkdir -p /opt/monitoring/grafana/provisioning/dashboards
mkdir -p /opt/monitoring/grafana/provisioning/dashboards
mkdir -p /opt/monitoring/data/prometheus
mkdir -p /opt/monitoring/data/grafana

# 6. prometheus.yml 생성
cat <<'EOF' > /opt/monitoring/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']

  - job_name: 'jaeger'
    metrics_path: /metrics
    static_configs:
      - targets: ['jaeger:14269']
EOF

# 7. Grafana provisioning - datasources.yaml
cat <<'EOF' > /opt/monitoring/grafana/provisioning/datasources/datasources.yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    version: 1
    editable: false
EOF

# 8. Grafana provisioning - dashboards.yaml
cat <<'EOF' > /opt/monitoring/grafana/provisioning/dashboards.yaml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# 9. Grafana sample dashboard (JSON)
cat <<'EOF' > /opt/monitoring/grafana/provisioning/dashboards/sample-dashboard.json
{
  "id": null,
  "title": "Sample Monitoring Dashboard",
  "panels": [
    {
      "type": "graph",
      "title": "Node Exporter CPU Usage",
      "targets": [
        {
          "expr": "rate(node_cpu_seconds_total[1m])",
          "legendFormat": "{{cpu}}"
        }
      ],
      "datasource": "Prometheus"
    }
  ],
  "schemaVersion": 16,
  "version": 0
}
EOF

# 10. docker-compose.yml 생성
cat <<'EOF' > /opt/monitoring/docker-compose.yml
version: "3.8"

services:
  node-exporter:
    image: quay.io/prometheus/node-exporter:latest
    container_name: node-exporter
    restart: unless-stopped
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - "--path.procfs=/host/proc"
      - "--path.rootfs=/rootfs"
      - "--path.sysfs=/host/sys"
      - "--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)"
    ports:
      - "9101:9100"
    networks:
      - monitoring

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "4318:4318"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - COLLECTOR_OTLP_HTTP_PORT=4318
    networks:
      - monitoring
    restart: always
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:16686"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
    ports:
      - "9090:9090"
    networks:
      - monitoring
    restart: always
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:9090"]
      interval: 10s
      timeout: 5s
      retries: 5

  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    ports:
      - "3000:3000"
    networks:
      - monitoring
    depends_on:
      prometheus:
        condition: service_healthy
    restart: always

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus_data:
  grafana_data:
EOF

# 11. 모니터링 경로 이동 및 실행
cd /opt/monitoring
docker-compose up -d
