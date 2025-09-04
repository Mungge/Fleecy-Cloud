# main.tf

# Ubuntu 이미지 조회 (AWS AMI와 동일한 역할)
data "google_compute_image" "ubuntu" {
  family  = "ubuntu-2204-lts"
  project = "ubuntu-os-cloud"
}

# VPC 네트워크 생성
resource "google_compute_network" "main" {
  name                    = "${var.project_name}-vpc"
  auto_create_subnetworks = false  # 수동으로 서브넷 생성
  mtu                     = 1460
}

# 서브넷 생성
resource "google_compute_subnetwork" "public" {
  name          = "${var.project_name}-public-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  network       = google_compute_network.main.id
}

# 방화벽 규칙 - SSH (항상 필요)
resource "google_compute_firewall" "ssh" {
  name    = "${var.project_name}-allow-ssh"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = var.allowed_ips
  target_tags   = ["${var.project_name}-server"]

  description = "Allow SSH access"
}

# 방화벽 규칙 - 개발 환경 (모든 TCP 포트)
resource "google_compute_firewall" "dev_all_tcp" {
  count = var.environment == "dev" ? 1 : 0
  
  name    = "${var.project_name}-allow-dev-all-tcp"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["1-65535"]
  }

  source_ranges = var.allowed_ips
  target_tags   = ["${var.project_name}-server"]

  description = "Allow all TCP ports for development"
}

# 방화벽 규칙 - HTTP/HTTPS (프로덕션)
resource "google_compute_firewall" "web" {
  count = var.environment != "dev" ? 1 : 0
  
  name    = "${var.project_name}-allow-web"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = var.allowed_ips
  target_tags   = ["${var.project_name}-server"]

  description = "Allow HTTP and HTTPS"
}

# 방화벽 규칙 - 사용자 지정 포트들 (프로덕션)
resource "google_compute_firewall" "custom_ports" {
  count = var.environment != "dev" && length(var.custom_ports) > 0 ? 1 : 0
  
  name    = "${var.project_name}-allow-custom-ports"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = [for port in var.custom_ports : tostring(port)]
  }

  source_ranges = var.allowed_ips
  target_tags   = ["${var.project_name}-server"]

  description = "Allow custom ports"
}

# 방화벽 규칙 - 모니터링 포트들 (프로덕션)
resource "google_compute_firewall" "monitoring" {
  count = var.environment != "dev" ? 1 : 0
  
  name    = "${var.project_name}-allow-monitoring"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["9090", "3000", "16686"]  # Prometheus, Grafana, Jaeger
  }

  source_ranges = var.allowed_ips
  target_tags   = ["${var.project_name}-server"]

  description = "Allow monitoring ports"
}

# 컴퓨트 인스턴스 (AWS EC2와 동일)
resource "google_compute_instance" "main" {
  name         = "${var.project_name}-server"
  machine_type = var.instance_type
  zone         = var.zone

  # 부팅 디스크 설정
  boot_disk {
    initialize_params {
      image = data.google_compute_image.ubuntu.self_link
      size  = 20  # GB
      type  = "pd-standard"  # 표준 영구 디스크
    }
  }

  # 네트워크 인터페이스
  network_interface {
    network    = google_compute_network.main.id
    subnetwork = google_compute_subnetwork.public.id
    
    # 외부 IP 할당 (AWS의 map_public_ip_on_launch와 동일)
    access_config {
      // 빈 블록은 임시 외부 IP를 자동 할당
    }
  }

  # 메타데이터 (SSH 키는 클라우드 API를 통해 이미 설정됨)
  metadata = {
    startup-script = var.startup_script
  }

  # 네트워크 태그 (방화벽 규칙 적용을 위해)
  tags = ["${var.project_name}-server"]

  # 서비스 계정 (최소 권한)
  service_account {
    email  = google_service_account.vm_service_account.email
    scopes = ["cloud-platform"]
  }

  labels = {
    environment = var.environment
    project     = var.project_name
  }
}

# VM용 서비스 계정 생성
resource "google_service_account" "vm_service_account" {
  account_id   = "${var.project_name}-vm-sa"
  display_name = "${var.project_name} VM Service Account"
  description  = "Service account for ${var.project_name} VM instances"
}