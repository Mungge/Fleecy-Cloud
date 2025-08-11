# 로컬 값들 정의
locals {
  key_name = "${var.project_name}-keypair-${var.environment}"
  dev_private_key_path = pathexpand("~/.ssh/${local.key_name}.pem")
  
  # 개발환경: 기존 키 존재 여부 확인 (개인키만)
  dev_key_exists = var.environment == "dev" && fileexists(local.dev_private_key_path)
}

# 배포환경: 시스템 서버 DB에서 키 가져오기
data "http" "get_keypair" {
  count = var.environment != "dev" ? 1 : 0
  
  url = "${var.api_endpoint}/keypairs/${var.project_name}/${var.environment}"
  
  request_headers = {
    Accept = "application/json"
    Authorization = "Bearer ${var.api_token}"
  }
  
  # 키가 없어도 에러 발생하지 않도록
  lifecycle {
    postcondition {
      condition = self.status_code == 200 || self.status_code == 404
      error_message = "API call failed. Expected 200 (key exists) or 404 (key not found), got ${self.status_code}"
    }
  }
}

# 배포환경: DB에서 가져온 키 파싱
locals {
  # DB에서 키 정보 파싱 (404면 키가 없음)
  db_keypair_exists = var.environment != "dev" ? (
    length(data.http.get_keypair) > 0 && data.http.get_keypair[0].status_code == 200
  ) : false
  
  db_public_key = var.environment != "dev" && local.db_keypair_exists ? (
    jsondecode(data.http.get_keypair[0].response_body).public_key
  ) : ""
}

# csv를 통한 자격 증명
locals {
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