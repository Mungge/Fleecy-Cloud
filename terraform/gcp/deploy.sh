#!/bin/bash

# GCP Terraform 배포 스크립트

set -e  # 에러 발생시 스크립트 중단

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 로그 함수
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 사용법 출력
usage() {
    echo "사용법: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "COMMANDS:"
    echo "  init     - Terraform 초기화"
    echo "  plan     - 실행 계획 확인"
    echo "  apply    - 인프라 배포"
    echo "  destroy  - 인프라 삭제"
    echo "  output   - 출력 값 확인"
    echo ""
    echo "OPTIONS:"
    echo "  -e, --env ENV     환경 설정 (dev, prod) [기본값: dev]"
    echo "  -p, --project ID  GCP 프로젝트 ID"
    echo "  -h, --help        도움말 출력"
    echo ""
    echo "예시:"
    echo "  $0 init --project my-gcp-project"
    echo "  $0 apply --env dev"
    echo "  $0 destroy --env prod"
}

# 기본 값
ENVIRONMENT="dev"
PROJECT_ID=""
COMMAND=""

# 파라미터 파싱
while [[ $# -gt 0 ]]; do
    case $1 in
        init|plan|apply|destroy|output)
            COMMAND="$1"
            shift
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -p|--project)
            PROJECT_ID="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "알 수 없는 옵션: $1"
            usage
            exit 1
            ;;
    esac
done

# 필수 파라미터 확인
if [[ -z "$COMMAND" ]]; then
    log_error "명령어를 지정해주세요."
    usage
    exit 1
fi

# GCP 인증 확인
log_info "GCP 인증 상태 확인..."
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
    log_error "GCP에 인증되지 않았습니다. 'gcloud auth login'을 실행하세요."
    exit 1
fi

# 현재 프로젝트 확인
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null || echo "")
if [[ -n "$PROJECT_ID" ]]; then
    if [[ "$CURRENT_PROJECT" != "$PROJECT_ID" ]]; then
        log_info "GCP 프로젝트를 $PROJECT_ID로 설정합니다..."
        gcloud config set project "$PROJECT_ID"
    fi
elif [[ -z "$CURRENT_PROJECT" ]]; then
    log_error "GCP 프로젝트가 설정되지 않았습니다. --project 옵션을 사용하거나 'gcloud config set project PROJECT_ID'를 실행하세요."
    exit 1
fi

# 환경별 변수 파일 확인
VAR_FILE="terraform.${ENVIRONMENT}.tfvars"
if [[ ! -f "$VAR_FILE" ]]; then
    log_warn "환경별 변수 파일 '$VAR_FILE'이 없습니다. terraform.tfvars를 사용합니다."
    VAR_FILE="terraform.tfvars"
fi

if [[ ! -f "$VAR_FILE" ]]; then
    log_error "변수 파일 '$VAR_FILE'이 없습니다. terraform.tfvars.example을 참고하여 작성하세요."
    exit 1
fi

# Terraform 명령 실행
case $COMMAND in
    init)
        log_info "Terraform을 초기화합니다..."
        terraform init
        ;;
    plan)
        log_info "Terraform 실행 계획을 확인합니다... (환경: $ENVIRONMENT)"
        terraform plan -var-file="$VAR_FILE" -var="environment=$ENVIRONMENT"
        ;;
    apply)
        log_info "인프라를 배포합니다... (환경: $ENVIRONMENT)"
        terraform apply -var-file="$VAR_FILE" -var="environment=$ENVIRONMENT"
        
        if [[ $? -eq 0 ]]; then
            log_info "배포가 완료되었습니다!"
            echo ""
            log_info "연결 정보:"
            terraform output ssh_connection
        fi
        ;;
    destroy)
        log_warn "인프라를 삭제합니다... (환경: $ENVIRONMENT)"
        echo "정말로 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        read -p "계속하려면 'yes'를 입력하세요: " confirm
        
        if [[ "$confirm" == "yes" ]]; then
            terraform destroy -var-file="$VAR_FILE" -var="environment=$ENVIRONMENT"
        else
            log_info "삭제가 취소되었습니다."
        fi
        ;;
    output)
        log_info "출력 값을 확인합니다..."
        terraform output
        ;;
esac