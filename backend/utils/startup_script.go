package utils

const startup_script = `
#!/bin/bash

set -e

# 아키텍처 자동 감지
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
    BINARY_ARCH="arm64"
    DEB_ARCH="arm64"
elif [ "$ARCH" = "x86_64" ]; then
    BINARY_ARCH="amd64"
    DEB_ARCH="amd64"
else
    echo "지원하지 않는 아키텍처: $ARCH"
    exit 1
fi

echo "========================================="
echo "연합학습 집계자 모니터링 시스템 설치 시작"
echo "시스템 아키텍처: $ARCH (바이너리: $BINARY_ARCH)"
echo "========================================="

echo "[1/8] 시스템 업데이트 및 필수 패키지 설치"
sudo apt-get update -y
sudo apt-get install -y wget curl tar adduser libfontconfig1 musl

echo "[2/8] 작업 디렉토리 생성"
sudo mkdir -p /opt/monitoring/{prometheus,grafana,node-exporter}
sudo mkdir -p /opt/monitoring/grafana/{data,provisioning/{datasources,dashboards}}
sudo mkdir -p /var/lib/prometheus
sudo mkdir -p /var/log/prometheus

echo "[3/8] Prometheus 다운로드 및 설치"
cd /tmp
PROM_URL="https://github.com/prometheus/prometheus/releases/download/v2.52.0/prometheus-2.52.0.linux-${BINARY_ARCH}.tar.gz"
echo "Prometheus 다운로드: $PROM_URL"
wget -q $PROM_URL
tar -xzf prometheus-2.52.0.linux-${BINARY_ARCH}.tar.gz
sudo cp prometheus-2.52.0.linux-${BINARY_ARCH}/prometheus /usr/local/bin/
sudo cp prometheus-2.52.0.linux-${BINARY_ARCH}/promtool /usr/local/bin/
sudo mkdir -p /etc/prometheus
sudo cp -r prometheus-2.52.0.linux-${BINARY_ARCH}/consoles /etc/prometheus/
sudo cp -r prometheus-2.52.0.linux-${BINARY_ARCH}/console_libraries /etc/prometheus/
rm -rf prometheus-2.52.0.linux-${BINARY_ARCH}*

echo "[4/8] Node Exporter 다운로드 및 설치"
NODE_URL="https://github.com/prometheus/node_exporter/releases/download/v1.8.1/node_exporter-1.8.1.linux-${BINARY_ARCH}.tar.gz"
echo "Node Exporter 다운로드: $NODE_URL"
wget -q $NODE_URL
tar -xzf node_exporter-1.8.1.linux-${BINARY_ARCH}.tar.gz
sudo cp node_exporter-1.8.1.linux-${BINARY_ARCH}/node_exporter /usr/local/bin/
rm -rf node_exporter-1.8.1.linux-${BINARY_ARCH}*

echo "[5/8] Grafana 다운로드 및 설치"
GRAFANA_URL="https://dl.grafana.com/enterprise/release/grafana-enterprise_10.4.0_${DEB_ARCH}.deb"
echo "Grafana 다운로드: $GRAFANA_URL"
wget -q $GRAFANA_URL
sudo dpkg -i grafana-enterprise_10.4.0_${DEB_ARCH}.deb || sudo apt-get install -f -y
rm grafana-enterprise_10.4.0_${DEB_ARCH}.deb

echo "[6/8] 설정 파일 생성"

# Prometheus 설정
sudo tee /etc/prometheus/prometheus.yml > /dev/null <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
      
  # 연합학습 집계자 애플리케이션 메트릭스 (필요시 추가)
  - job_name: 'federated-aggregator'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 10s
EOF

# Grafana datasources 설정
sudo mkdir -p /etc/grafana/provisioning/datasources
sudo tee /etc/grafana/provisioning/datasources/datasources.yaml > /dev/null <<EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://localhost:9090
    isDefault: true
    version: 1
    editable: false
EOF

# Grafana dashboards 설정
sudo mkdir -p /etc/grafana/provisioning/dashboards
sudo tee /etc/grafana/provisioning/dashboards/dashboards.yaml > /dev/null <<EOF
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

# 연합학습 집계자용 대시보드 생성
sudo tee /etc/grafana/provisioning/dashboards/federated-aggregator-dashboard.json > /dev/null <<'DASHBOARD_EOF'
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
DASHBOARD_EOF

echo "[7/8] systemd 서비스 파일 생성"

# Node Exporter 서비스
sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOF
[Unit]
Description=Node Exporter
After=network.target

[Service]
ExecStart=/usr/local/bin/node_exporter
User=nobody
Group=nogroup
Restart=always

[Install]
WantedBy=default.target
EOF

# Prometheus 서비스
sudo tee /etc/systemd/system/prometheus.service > /dev/null <<EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \\
    --config.file /etc/prometheus/prometheus.yml \\
    --storage.tsdb.path /var/lib/prometheus/ \\
    --web.console.templates=/etc/prometheus/consoles \\
    --web.console.libraries=/etc/prometheus/console_libraries \\
    --web.listen-address=0.0.0.0:9090 \\
    --web.enable-lifecycle \\
    --storage.tsdb.retention.time=200h

[Install]
WantedBy=multi-user.target
EOF

# Prometheus 사용자 생성
sudo useradd --no-create-home --shell /bin/false prometheus || true
sudo chown -R prometheus:prometheus /etc/prometheus
sudo chown -R prometheus:prometheus /var/lib/prometheus
sudo chown -R prometheus:prometheus /var/log/prometheus

# Grafana 설정 수정
sudo tee -a /etc/grafana/grafana.ini > /dev/null <<EOF

[security]
admin_user = admin
admin_password = admin123
allow_embedding = true
cookie_samesite = none

[auth.anonymous]
enabled = true
org_role = Viewer

[panels]
disable_sanitize_html = true

[users]
allow_sign_up = false
EOF

echo "[8/8] 서비스 시작 및 활성화"

# systemd 리로드
sudo systemctl daemon-reload

# Node Exporter 시작
sudo systemctl enable node_exporter
sudo systemctl start node_exporter

# Prometheus 시작
sudo systemctl enable prometheus
sudo systemctl start prometheus

# Grafana 시작
sudo systemctl enable grafana-server
sudo systemctl start grafana-server

echo
echo "========================================="
echo "연합학습 집계자 모니터링 시스템 설치 완료!"
echo "========================================="
echo "Grafana: http://localhost:3000 (admin/admin123)"
echo "Prometheus: http://localhost:9090"
echo "Node Exporter: http://localhost:9100"
echo "========================================="
echo
echo "서비스 상태 확인:"
sudo systemctl status node_exporter --no-pager -l
sudo systemctl status prometheus --no-pager -l  
sudo systemctl status grafana-server --no-pager -l`