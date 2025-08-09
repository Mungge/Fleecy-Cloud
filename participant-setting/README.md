<p align="center">
	<img src="https://img.shields.io/badge/OpenStack-Node%20Exporter%20%7C%20Prometheus%20%7C%20Grafana-blue?style=for-the-badge&logo=OpenStack" alt="OpenStack Monitoring" />
</p>

# 🚀 OpenStack VM 모니터링: Node Exporter, Prometheus, Grafana 연동 가이드

<details>
<summary><strong>목표 (클릭하여 펼치기)</strong></summary>

✔️ OpenStack VM에 Node Exporter를 설치하여 시스템 메트릭을 수집<br>
✔️ DevStack VM에 설치된 Prometheus와 Grafana에서 OpenStack VM의 메트릭을 시각화 및 모니터링

</details>

---

```mermaid
flowchart LR
		A[OpenStack VM] -- Node Exporter --> B[Prometheus (DevStack VM)] -- 데이터 소스 --> C[Grafana (DevStack VM)]
```

| 구성 요소         | 역할 설명                                                 |
| ----------------- | --------------------------------------------------------- |
| **Node Exporter** | OpenStack VM의 시스템 메트릭(메모리, CPU, 디스크 등) 수집 |
| **Prometheus**    | Node Exporter에서 메트릭을 주기적으로 스크랩              |
| **Grafana**       | Prometheus 데이터를 시각화                                |

---

## 2️⃣ OpenStack 설치

### 2-1. devstack 설치

```bash
sudo apt-get update
sudo apt install git -y
git clone https://opendev.org/openstack/devstack
cd devstack
```

### 2-2. local.conf 생성

```bash
vim local.conf
```

> 💡 **TIP:** 현재 디렉토리의 local.conf 예시를 참고하세요.

### 2-3. 설치 실행

```bash
./stack.sh
```

---

## 3️⃣ 모니터링 설정

OpenStack을 설치한 VM에서 아래 스크립트를 실행하면 Prometheus 및 Grafana가 설치되고, VM 모니터링이 자동으로 설정됩니다.

### 3-1. VM 목록 수정

`install_monitoring.sh`의 `TARGETS=()`에 모니터링할 VM의 IP 주소를 추가하세요.

```bash
./install_monitoring.sh
```

### 3-2. grafana.ini 파일 수정

웹 임베딩 허용을 위해 `grafana.ini` 파일을 수정합니다.

```bash
vim ~/etc/grafana/grafana.ini
```

다음 설정을 파일에 추가하거나 수정하세요:

```ini
# 익명 접근 허용 설정
[auth.anonymous]
enabled = true
org_name = Main Org.
org_role = Viewer

# 보안 설정 - 임베딩 허용
[security]
allow_embedding = true
```

설정 변경 후 Grafana 서비스를 재시작

```bash
sudo systemctl restart grafana-server
```

---

## 📚 참고 자료

- [Prometheus Node Exporter 공식 문서](https://prometheus.io/docs/guides/node-exporter/)
- [Grafana Dashboards - Node Exporter Full](https://grafana.com/grafana/dashboards/1860-node-exporter-full/)
- [Prometheus 공식 문서](https://prometheus.io/docs/introduction/overview/)
- [Grafana 공식 문서](https://grafana.com/docs/)
