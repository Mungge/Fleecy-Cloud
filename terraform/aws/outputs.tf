# 백엔드가 기대하는 출력 형식
output "instance_id" {
  value       = aws_instance.main.id
  description = "EC2 인스턴스 ID"
}

output "public_ip" {
  value       = aws_instance.main.public_ip
  description = "EC2 인스턴스 공인 IP"
}

output "private_ip" {
  value       = aws_instance.main.private_ip
  description = "EC2 인스턴스 사설 IP"
}

# 기존 outputs (호환성 유지)
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
  }
  description = "SSH 키 관리 상태 정보"
}

# SSH 연결 정보
output "ssh_connection_info" {
  value = {
    ssh_command = "ssh -i ~/.ssh/${local.key_name}.pem ${var.ssh_username}@${aws_instance.main.public_ip}"
    key_name = local.key_name
  }
  description = "SSH 연결 정보"
}

# 키페어 생성 정보
output "new_keypair_created" {
  value = {
    message = "키페어가 생성되어 DB에 저장되었습니다"
    key_name = local.key_name
  }
  description = "생성된 키페어 정보"
}