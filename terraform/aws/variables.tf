variable "aws_region" {
  description = "AWS 리전"
  type        = string
  default     = "ap-northeast-2"
}

variable "availability_zone" {
  description = "가용 영역"
  type        = string
  default     = "ap-northeast-2a"
}

variable "project_name" {
  description = "프로젝트 이름"
  type        = string
  default     = "fleecy-cloud"
}

variable "instance_type" {
  description = "EC2 인스턴스 스펙"
  type        = string
  default     = "t2.micro"
}

variable "ssh_public_key_content" {
  description = "SSH 공개키 내용 (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
}

variable "ssh_username" {
  description = "SSH 사용자명"
  type        = string
  default     = "ubuntu"
}

variable "environment" {
  description = "환경 (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "allowed_ips" {
  description = "허용할 IP 주소 목록"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "custom_ports" {
  description = "추가로 열고 싶은 포트 목록 (개발환경이 아닐 때)"
  type        = list(number)
  default     = [8080, 9000, 5000]
}

variable "aws_access_key" {
  description = "AWS Access Key ID (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS Secret Access Key (프로덕션용, DB에서 전달)"
  type        = string
  default     = ""
  sensitive   = true
}

# 배포 환경용: API에서 키 정보 가져오기
variable "api_endpoint" {
  description = "키 관리 API 엔드포인트"
  type        = string
  default     = ""
}

variable "api_token" {
  description = "API 인증 토큰"
  type        = string
  default     = ""
  sensitive   = true
}

# 배포 환경에서 API를 통해 공개키 가져오기
data "http" "public_key" {
  count = var.environment != "dev" && var.api_endpoint != "" ? 1 : 0
  
  url = "${var.api_endpoint}/v1/keys/${var.project_name}/${var.environment}/public"
  
  request_headers = {
    Accept = "application/json"
    Authorization = "Bearer ${var.api_token}"
  }
}

# SSH 키 동적 처리 (수정됨)
locals {
  # SSH 키 파일 경로들 확인 (개발용)
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
  
  # API에서 받은 공개키 파싱
  api_public_key = length(data.http.public_key) > 0 ? jsondecode(data.http.public_key[0].response_body).public_key : ""
  
  # SSH 키 내용 결정 (우선순위: 1.변수값, 2.API, 3.파일)
  ssh_public_key_content = (
    # 1. 변수로 직접 전달된 경우 (프로덕션)
    var.ssh_public_key_content != "" ? var.ssh_public_key_content :
    # 2. API에서 가져온 경우 (배포환경)
    local.api_public_key != "" ? local.api_public_key :
    # 3. 로컬 파일에서 읽는 경우 (개발환경)
    local.existing_ssh_file != null ? file(local.existing_ssh_file) : ""
  )
  
  # AWS 자격증명 처리
  csv_file_exists = fileexists("${path.module}/../credentials/keys.csv")
  csv_content     = local.csv_file_exists ? file("${path.module}/../credentials/keys.csv") : ""
  csv_lines       = local.csv_content != "" ? split("\n", local.csv_content) : []
  credential_data = length(local.csv_lines) > 1 ? split(",", local.csv_lines[1]) : ["", ""]
  
  # 최종 키 값 결정 (변수 우선 = 배포용, 없으면 CSV에서 읽기 = 개발용)
  aws_access_key = var.aws_access_key != "" ? var.aws_access_key : (
    length(local.credential_data) > 0 ? trimspace(local.credential_data[0]) : ""
  )
  aws_secret_key = var.aws_secret_key != "" ? var.aws_secret_key : (
    length(local.credential_data) > 1 ? trimspace(local.credential_data[1]) : ""
  )
}