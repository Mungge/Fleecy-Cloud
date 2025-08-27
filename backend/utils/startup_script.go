package utils

const startup_script = `#!/bin/bash
# 모든 출력을 로그 파일로 리디렉션하여 디버깅을 쉽게 만듭니다.
exec > /tmp/startup_script.log 2>&1

set -e

PROM_VERSION="2.52.0"
TARGETS_JSON="targets.json"

TARGETS=(
  "localhost:9100"
)

echo "[1/6] Prometheus 다운로드 및 설치"
wget -q https://github.com/prometheus/prometheus/releases/download/v$${PROM_VERSION}/prometheus-$${PROM_VERSION}.linux-amd64.tar.gz
tar -xzf prometheus-$${PROM_VERSION}.linux-amd64.tar.gz
rm prometheus-$${PROM_VERSION}.linux-amd64.tar.gz
mv prometheus-$${PROM_VERSION}.linux-amd64 prometheus

echo "[2/6] Prometheus 설정 생성"
cat > prometheus/prometheus.yml <<'EOF'
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'openstack-vms'
    file_sd_configs:
      - files:
          - 'targets.json'
        refresh_interval: 30s
EOF

echo "[3/6] targets.json 생성"
cat > prometheus/$${TARGETS_JSON} <<'EOF'
[
  {
    "labels": {
      "job": "openstack-vm"
    },
    "targets": [
EOF

for i in "$${!TARGETS[@]}"; do
  SEP=","
  if [ "$i" -eq $(($${#TARGETS[@]} - 1)) ]; then SEP=""; fi
  echo "      \"$${TARGETS[$i]}\"$${SEP}" >> prometheus/$${TARGETS_JSON}
done

cat >> prometheus/$${TARGETS_JSON} <<'EOF'
    ]
  }
]
EOF

echo "[4/6] Prometheus 실행"
nohup ./prometheus/prometheus --config.file=prometheus/prometheus.yml > prometheus.log 2>&1 &

echo "[5/6] Grafana 설치"
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://packages.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://packages.grafana.com/oss/deb stable main" | sudo tee /etc/apt/sources.list.d/grafana.list > /dev/null
sudo apt-get update

# apt-get이 질문하지 않도록 noninteractive 옵션을 추가합니다.
echo "Grafana 설치 시작..."
export DEBIAN_FRONTEND=noninteractive
sudo apt-get install -y grafana

sudo systemctl enable grafana-server
sudo systemctl start grafana-server

echo
echo "설치 완료"
echo "Prometheus: http://localhost:9090"
echo "Grafana   : http://localhost:3000 (admin/admin)"
echo "prometheus/targets.json에서 VM 추가 가능"
`