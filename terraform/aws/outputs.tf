# 기존 outputs 수정 (EC2 인스턴스 관련)
output "ssh_command" {
  value       = "ssh -i ~/.ssh/${local.key_name}.pem ${var.ssh_username}@${aws_instance.main.public_ip}"
  description = "SSH 연결 명령어"
}

output "instance_info" {
  value = {
    public_ip    = aws_instance.main.public_ip
    private_ip   = aws_instance.main.private_ip
    instance_id  = aws_instance.main.id
    key_name     = aws_instance.main.key_name
  }
  description = "EC2 인스턴스 정보"
}

# SSH 키 관리 상태 정보
output "key_management_info" {
  value = {
    environment = var.environment
    key_name = local.key_name
    
    # 개발환경 정보
    dev_existing_key = var.environment == "dev" ? local.dev_key_exists : null
    dev_private_key_path = var.environment == "dev" ? local.dev_private_key_path : null
    
    # 배포환경 정보
    prod_key_source = var.environment != "dev" ? (
      local.db_keypair_exists ? "existing_db" : "newly_created"
    ) : null
    
    db_key_exists = var.environment != "dev" ? local.db_keypair_exists : null
  }
  description = "SSH 키 관리 상태 정보"
}

# 개발환경 SSH 연결 정보
output "ssh_connection_info" {
  value = var.environment == "dev" ? {
    ssh_command = "ssh -i ${local.dev_private_key_path} ${var.ssh_username}@${aws_instance.main.public_ip}"
    private_key_path = local.dev_private_key_path
    key_exists = local.dev_key_exists
  } : null
  description = "개발환경 SSH 연결 정보"
}

# 새로 생성된 키 정보 (배포환경)
output "new_keypair_created" {
  value = var.environment != "dev" && !local.db_keypair_exists ? {
    message = "새 키페어가 생성되어 DB에 저장되었습니다"
    key_name = local.key_name
  } : null
  description = "새로 생성된 키페어 정보"
}