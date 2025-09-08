# GCP 프로젝트 설정
project_id   = "fleecy-cloud"
project_name = "fleecy-cloud"

# 지역 설정 (gcp의 무료 요금제는 특정 리전에서만 가능)
region = "us-central1"    # 무료 요금제 지원 리전
zone   = "us-central1-a"  # 무료 요금제 지원 존

# 환경 설정
environment = "dev"  # dev 또는 prod

# 인스턴스 설정
instance_type = "f1-micro"  # aws t2.micro와 유사

# 보안 설정
allowed_ips = ["0.0.0.0/0"]  # 모든 IP에서 접근 허용 (개발용)

# SSH 설정
ssh_public_key_path = "~/.ssh/id_rsa.pub"
ssh_username        = "ubuntu"

# 프로덕션 환경용 포트 설정
custom_ports = [8080, 9000, 5000, 9092]  # 9092: Kafka