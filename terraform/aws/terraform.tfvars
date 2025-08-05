project_name   = "fleecy-cloud"
instance_type  = "t2.micro"
key_pair_name  = "fleecy-cloud-keypair"  # AWS 콘솔에서 미리 생성한 키페어 이름

# 보안 설정
environment    = "dev"                  # dev, staging, prod
allowed_ips    = ["0.0.0.0/0"]         # 개발용은 모든 IP, 프로덕션에서는 특정 IP만
custom_ports   = [8080, 9000, 5000]    # 프로덕션 환경에서 추가로 열 포트들