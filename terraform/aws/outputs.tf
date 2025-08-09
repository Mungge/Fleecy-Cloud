output "instance_public_ip" {
  description = "EC2 Instance Public IP"
  value       = aws_instance.main.public_ip
}

output "ssh_command" {
  description = "SSH Connection Command"
  value       = "ssh -i ~/.ssh/${local.key_name}.pem ubuntu@${aws_instance.main.public_ip}"
}

output "key_management_info" {
  description = "키 관리 방식 정보"
  value = {
    environment = var.environment
    key_source = (
      var.ssh_public_key_content != "" ? "직접 변수" :
      local.api_public_key != "" ? "API" :
      "로컬 파일"
    )
    key_name = local.key_name
  }
}

output "private_key_path" {
  description = "개발 환경용 개인키 파일 경로"
  value       = var.environment == "dev" ? pathexpand("~/.ssh/${local.key_name}.pem") : "DB에서 다운로드 필요"
}