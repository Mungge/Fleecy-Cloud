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
      
  # 연합학습 집계자 애플리케이션 메트릭스 (필요시 추가)
  - job_name: 'federated-aggregator'
    static_configs:
      - targets: ['host.docker.internal:8080']  # 애플리케이션 메트릭스 엔드포인트
    scrape_interval: 10s
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
cat <<'EOF' > /opt/monitoring/grafana/provisioning/dashboards/dashboards.yaml
apiVersion: 1

providers:
  - name: 'federated-learning'
    orgId: 1
    folder: 'Federated Learning'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# 9. 연합학습 집계자용 대시보드 생성
cat <<'EOF' > /opt/monitoring/grafana/provisioning/dashboards/federated-aggregator-dashboard.json
{
  "id": null,
  "title": "연합학습 집계자 모니터링",
  "tags": ["federated-learning", "aggregator"],
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "refresh": "30s",
  "panels": [
    {
      "id": 1,
      "title": "CPU 사용률",
      "type": "stat",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "targets": [
        {
          "expr": "100 - (avg(irate(node_cpu_seconds_total{mode=\"idle\"}[1m])) * 100)",
          "legendFormat": "CPU Usage %"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 60},
              {"color": "red", "value": 80}
            ]
          }
        }
      }
    },
    {
      "id": 2,
      "title": "메모리 사용률",
      "type": "stat",
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "targets": [
        {
          "expr": "((node_memory_MemTotal_bytes - node_memory_MemFree_bytes - node_memory_Buffers_bytes - node_memory_Cached_bytes) / node_memory_MemTotal_bytes) * 100",
          "legendFormat": "Memory Usage %"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 70},
              {"color": "red", "value": 85}
            ]
          }
        }
      }
    },
    {
      "id": 3,
      "title": "디스크 사용률",
      "type": "gauge",
      "gridPos": {"h": 8, "w": 8, "x": 0, "y": 8},
      "targets": [
        {
          "expr": "100 - ((node_filesystem_avail_bytes{mountpoint=\"/\"} * 100) / node_filesystem_size_bytes{mountpoint=\"/\"})",
          "legendFormat": "Root Disk Usage %"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "max": 100,
          "thresholds": {
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 70},
              {"color": "red", "value": 85}
            ]
          }
        }
      }
    },
    {
      "id": 4,
      "title": "네트워크 I/O",
      "type": "timeseries",
      "gridPos": {"h": 8, "w": 8, "x": 8, "y": 8},
      "targets": [
        {
          "expr": "rate(node_network_receive_bytes_total[1m])",
          "legendFormat": "Receive {{device}}"
        },
        {
          "expr": "rate(node_network_transmit_bytes_total[1m])",
          "legendFormat": "Transmit {{device}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "binBps"
        }
      }
    },
    {
      "id": 5,
      "title": "시스템 로드",
      "type": "timeseries",
      "gridPos": {"h": 8, "w": 8, "x": 16, "y": 8},
      "targets": [
        {
          "expr": "node_load1",
          "legendFormat": "1m load avg"
        },
        {
          "expr": "node_load5",
          "legendFormat": "5m load avg"
        },
        {
          "expr": "node_load15",
          "legendFormat": "15m load avg"
        }
      ]
    }
  ],
  "schemaVersion": 30,
  "version": 1
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

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
      - "--web.console.libraries=/etc/prometheus/console_libraries"
      - "--web.console.templates=/etc/prometheus/consoles"
      - "--storage.tsdb.retention.time=200h"
      - "--web.enable-lifecycle"
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
    container_name: grafana
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
      # 임베드 및 API 접근 허용
      - GF_SECURITY_ALLOW_EMBEDDING=true
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
      - GF_SECURITY_COOKIE_SAMESITE=none
      - GF_PANELS_DISABLE_SANITIZE_HTML=true
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

echo "========================================="
echo "연합학습 집계자 모니터링 시스템 설치 완료!"
echo "========================================="
echo "Grafana: http://localhost:3000 (admin/admin123)"
echo "Prometheus: http://localhost:9090"
echo "Node Exporter: http://localhost:9101"
echo "========================================="