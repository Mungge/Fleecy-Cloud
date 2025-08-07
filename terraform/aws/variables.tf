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

# CSV 파일에서 AWS 키 읽기 (개발용)
locals {
  # 프로덕션에서는 변수값 사용, 개발에서는 CSV 파일 사용
  use_csv_file = var.aws_access_key == "" || var.aws_secret_key == ""
  
  # CSV 파일 읽고 파싱 (파일이 있을 때만)
  csv_file_exists = fileexists("${path.module}/../credentials/keys.csv")
  csv_content     = local.csv_file_exists ? file("${path.module}/../credentials/keys.csv") : ""
  csv_lines       = local.csv_content != "" ? split("\n", local.csv_content) : []
  credential_data = length(local.csv_lines) > 1 ? split(",", local.csv_lines[1]) : ["", ""]
  
  # 최종 키 값 결정 (변수 우선, 없으면 CSV에서 읽기)
  aws_access_key = var.aws_access_key != "" ? var.aws_access_key : (
    length(local.credential_data) > 0 ? trimspace(local.credential_data[0]) : ""
  )
  aws_secret_key = var.aws_secret_key != "" ? var.aws_secret_key : (
    length(local.credential_data) > 1 ? trimspace(local.credential_data[1]) : ""
  )
}