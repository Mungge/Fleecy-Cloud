# 로컬 값들 정의
locals {
  key_name = "${var.project_name}-${var.aggregator_id}-keypair"
}

# AWS 자격증명은 변수로 전달됨