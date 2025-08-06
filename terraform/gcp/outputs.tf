# outputs.tf

output "instance_id" {
  description = "생성된 인스턴스의 ID"
  value       = google_compute_instance.main.id
}

output "instance_name" {
  description = "생성된 인스턴스의 이름"
  value       = google_compute_instance.main.name
}

output "external_ip" {
  description = "인스턴스의 외부 IP 주소"
  value       = google_compute_instance.main.network_interface[0].access_config[0].nat_ip
}

output "internal_ip" {
  description = "인스턴스의 내부 IP 주소"
  value       = google_compute_instance.main.network_interface[0].network_ip
}

output "vpc_network_name" {
  description = "VPC 네트워크 이름"
  value       = google_compute_network.main.name
}

output "subnet_name" {
  description = "서브넷 이름"
  value       = google_compute_subnetwork.public.name
}

output "ssh_connection" {
  description = "SSH 연결 명령어"
  value       = "ssh ${var.ssh_username}@${google_compute_instance.main.network_interface[0].access_config[0].nat_ip}"
}

output "zone" {
  description = "인스턴스가 생성된 존"
  value       = google_compute_instance.main.zone
}