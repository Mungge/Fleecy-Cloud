#!/bin/bash

set -e

NODE_EXPORTER_VERSION="1.8.1"

echo "[1/4] node_exporter 다운로드"
wget -q https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
tar -xzf node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
rm node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64.tar.gz
sudo mv node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64/node_exporter /usr/local/bin/
rm -rf node_exporter-${NODE_EXPORTER_VERSION}.linux-amd64

echo "[2/4] systemd 서비스 유닛 파일 생성"
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

echo "[3/4] systemd 서비스 등록 및 실행"
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable node_exporter
sudo systemctl start node_exporter

echo "[4/4] 상태 확인"
sudo systemctl status node_exporter --no-pager

echo
echo "🎉 node_exporter 실행 완료: http://<이VM의IP>:9100/metrics"
