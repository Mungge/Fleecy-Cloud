#!/bin/bash

set -e

# 🔧 Prometheus 버전
PROM_VERSION="2.52.0"
TARGETS_JSON="targets.json"

# TARGETS 수정하여 Monitoring할 VM 설정
TARGETS=(
  "172.24.4.101:9100"
  "172.24.4.102:9100"
)

echo "[1/6] Prometheus 다운로드 및 설치"
wget -q https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/prometheus-${PROM_VERSION}.linux-amd64.tar.gz
tar -xzf prometheus-${PROM_VERSION}.linux-amd64.tar.gz
rm prometheus-${PROM_VERSION}.linux-amd64.tar.gz
mv prometheus-${PROM_VERSION}.linux-amd64 prometheus

echo "[2/6] Prometheus 설정 생성"
cat > prometheus/prometheus.yml <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'openstack-vms'
    file_sd_configs:
      - files:
          - '${TARGETS_JSON}'
        refresh_interval: 30s
EOF

echo "[3/6] targets.json 생성"
cat > prometheus/${TARGETS_JSON} <<EOF
[
  {
    "labels": {
      "job": "openstack-vm"
    },
    "targets": [
EOF

for i in "${!TARGETS[@]}"; do
  SEP=","
  if [ "$i" -eq $((${#TARGETS[@]} - 1)) ]; then SEP=""; fi
  echo "      \"${TARGETS[$i]}\"${SEP}" >> prometheus/${TARGETS_JSON}
done

cat >> prometheus/${TARGETS_JSON} <<EOF
    ]
  }
]
EOF

echo "[4/6] Prometheus 실행"
nohup ./prometheus/prometheus --config.file=prometheus/prometheus.yml > prometheus.log 2>&1 &

echo "[5/6] Grafana GPG 키 및 저장소 설정"
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://packages.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null

echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://packages.grafana.com/oss/deb stable main" | sudo tee /etc/apt/sources.list.d/grafana.list > /dev/null

echo "[6/6] Grafana 설치 및 실행"
sudo apt-get update
sudo apt-get install -y grafana
sudo systemctl enable grafana-server
sudo systemctl start grafana-server

echo
echo "🎉 설치 완료!"
echo "────────────────────────────"
echo "🔗 Prometheus: http://localhost:9090"
echo "🔗 Grafana   : http://localhost:3000 (admin/admin)"
echo "📁 prometheus/targets.json에서 VM 추가 가능"
