# variables.tf
variable "project_id" {
  description = "GCP 프로젝트 ID"
  type        = string
}

variable "project_name" {
  description = "프로젝트 이름 (리소스 명명에 사용)"
  type        = string
}

variable "region" {
  description = "GCP 리전"
  type        = string
  default     = "us-central1"  # 개발용: 무료 리전
}

variable "zone" {
  description = "GCP 존"
  type        = string
  default     = "us-central1-a"  # 개발용: 무료 리전
}

variable "environment" {
  description = "환경 (dev, prod)"
  type        = string
  default     = "dev"
}

variable "instance_type" {
  description = "VM 인스턴스 타입"
  type        = string
  default     = "f1-micro"  # 개발용: AWS t2.micro과 유사한 무료 티어
}

variable "allowed_ips" {
  description = "접근 허용할 IP CIDR 블록들"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "custom_ports" {
  description = "프로덕션 환경에서 허용할 사용자 지정 포트들"
  type        = list(number)
  default     = [8080, 9000]
}

variable "ssh_public_key_content" {
  description = "SSH 공개키 내용 (파일 경로가 아닌 실제 키 내용)"
  type        = string
  default= ""
}

variable "ssh_username" {
  description = "SSH 사용자명"
  type        = string
  default     = "ubuntu"
}

locals {
  # SSH 키 파일 경로들 확인
  ssh_key_files = [
    "${path.module}/../credentials/id_rsa.pub",
    "${path.module}/../credentials/ssh_key.pub", 
    "~/.ssh/id_rsa.pub"
  ]
  
  # 존재하는 SSH 키 파일 찾기
  existing_ssh_file = coalesce([
    for path in local.ssh_key_files : 
    fileexists(path) ? path : null
  ]...)
  
  # SSH 키 내용 결정 (변수 우선, 없으면 파일에서 읽기)
  ssh_public_key_content = var.ssh_public_key_content != "" ? var.ssh_public_key_content : (
    local.existing_ssh_file != null ? file(local.existing_ssh_file) : ""
  )
  
  # GCP용 SSH 키 메타데이터
  ssh_keys = "${var.ssh_username}:${local.ssh_public_key_content}"
}

variable "gcp_credentials_json" {
  description = "GCP 서비스 계정 키 JSON 내용 (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
  sensitive   = true
}

locals {
  # GCP 인증 방식 결정
  gcp_file_exists = fileexists("${path.module}/../credentials/service-account.json")
  
  # 프로덕션: 변수값 사용, 개발: 파일 사용
  gcp_credentials = var.gcp_credentials_json != "" ? var.gcp_credentials_json : (
    local.gcp_file_exists ? file("${path.module}/../credentials/service-account.json") : ""
  )
}