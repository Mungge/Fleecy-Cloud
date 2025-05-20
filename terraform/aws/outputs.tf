output "vpc_id" {
  value = aws_vpc.main.id
}

output "instance_id" {
  value = aws_instance.web.id
}

output "instance_public_ip" {
  value = aws_instance.web.public_ip
}