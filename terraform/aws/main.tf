# Ubuntu AMI 조회
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# 키 페어 이름 설정
locals {
  key_name = "${var.project_name}-keypair-${var.environment}"
}

# 1. 개발 환경용: Terraform에서 키 페어 생성
resource "tls_private_key" "dev" {
  count = var.environment == "dev" ? 1 : 0
  
  algorithm = "RSA"
  rsa_bits  = 4096
}

# 2. 개발 환경용: 개인키를 로컬에 저장
resource "local_file" "private_key_dev" {
  count = var.environment == "dev" ? 1 : 0
  
  content         = tls_private_key.dev[0].private_key_pem
  filename        = pathexpand("~/.ssh/${local.key_name}.pem")
  file_permission = "0600"
}

# 3. 개발 환경용 키 페어 (공개키 소스 변경)
resource "aws_key_pair" "dev" {
  count = var.environment == "dev" ? 1 : 0
  
  key_name   = local.key_name
  public_key = tls_private_key.dev[0].public_key_openssh  # ← 변경됨
}

# 배포 환경용 키 페어 (키 변경 무시)
resource "aws_key_pair" "prod" {
  count = var.environment != "dev" ? 1 : 0
  
  key_name   = local.key_name
  public_key = local.ssh_public_key_content

  tags = {
    Name = local.key_name
    Environment = var.environment
  }
  
  lifecycle {
    ignore_changes = [public_key]
  }
}

# 환경에 따라 사용할 키 페어 결정
locals {
  key_pair_name = var.environment == "dev" ? aws_key_pair.dev[0].key_name : aws_key_pair.prod[0].key_name
}

# VPC 생성
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# 인터넷 게이트웨이
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# 퍼블릭 서브넷
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-public-subnet"
  }
}

# 라우팅 테이블
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-route-table"
  }
}

# 라우팅 테이블 연결
resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# 보안그룹 (환경별 동적 설정)
resource "aws_security_group" "main" {
  name        = "${var.project_name}-sg"
  description = "Security group for federated learning"
  vpc_id      = aws_vpc.main.id

  # SSH 접근 (항상 필요)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.allowed_ips
    description = "SSH"
  }

  # 개발 환경: 모든 TCP 포트 허용
  dynamic "ingress" {
    for_each = var.environment == "dev" ? [1] : []
    content {
      from_port   = 1
      to_port     = 65535
      protocol    = "tcp"
      cidr_blocks = var.allowed_ips
      description = "All TCP ports for development"
    }
  }

  # 프로덕션 환경: 기본 포트들만
  dynamic "ingress" {
    for_each = var.environment != "dev" ? [1] : []
    content {
      from_port   = 80
      to_port     = 80
      protocol    = "tcp"
      cidr_blocks = var.allowed_ips
      description = "HTTP"
    }
  }

  dynamic "ingress" {
    for_each = var.environment != "dev" ? [1] : []
    content {
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      cidr_blocks = var.allowed_ips
      description = "HTTPS"
    }
  }

  # 프로덕션 환경: 사용자 지정 포트들
  dynamic "ingress" {
    for_each = var.environment != "dev" ? var.custom_ports : []
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = var.allowed_ips
      description = "Custom port ${ingress.value}"
    }
  }

  # 모니터링 관련 보안그룹
  dynamic "ingress" {
    for_each = var.environment != "dev" ? [9090, 3000, 16686] : []
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = var.allowed_ips
      description = "Monitoring port ${ingress.value}"
    }
  }

  # 모든 아웃바운드 허용
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-sg"
  }
}

# EC2 인스턴스
resource "aws_instance" "main" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  key_name               = local.key_name
  vpc_security_group_ids = [aws_security_group.main.id]

  user_data = file("${path.module}/../common/scripts/setup-monitoring.sh")

  tags = {
    Name = "${var.project_name}-server"
  }
}