#!/usr/bin/env bash
set -euo pipefail

# 위치 계산
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." &>/dev/null && pwd)"

VENV_DIR="${PROJECT_ROOT}/.venv"
PY_SCRIPT="${SCRIPT_DIR}/aggregator_optimization.py"

# requirements.txt 우선순위: scripts/ → 프로젝트 루트
REQ_FILE="${SCRIPT_DIR}/requirements.txt"
if [[ ! -f "${REQ_FILE}" ]]; then
  REQ_FILE="${PROJECT_ROOT}/requirements.txt"
fi

# 사용법: run_optimizer.sh <input_json> <output_json> [--rm-input]
if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "Usage: $0 <input_json> <output_json> [--rm-input]" >&2
  exit 2
fi

INPUT_JSON="$1"
OUTPUT_JSON="$2"
RM_INPUT="false"
if [[ "${3:-}" == "--rm-input" ]]; then
  RM_INPUT="true"
fi

# cleanup: 언제 끝나든(정상/에러/시그널) 마지막에 실행
cleanup() {
  # 입력 파일 삭제 옵션
  if [[ "${RM_INPUT}" == "true" ]]; then
    rm -f -- "${INPUT_JSON}" || true
  fi
  # 가상환경 비활성화 (활성화 되어 있을 때만)
  if [[ -n "${VIRTUAL_ENV:-}" ]]; then
    # deactivate는 venv 활성화 후에만 존재하는 함수
    command -v deactivate >/dev/null 2>&1 && deactivate || true
  fi
}
trap cleanup EXIT INT TERM

# python3 확인
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found in PATH" >&2
  exit 3
fi

# 가상환경 생성(없으면) 및 활성화
if [[ ! -d "${VENV_DIR}" ]]; then
  python3 -m venv "${VENV_DIR}"
fi
# shellcheck disable=SC1091
source "${VENV_DIR}/bin/activate"

# 의존성 설치
python -m pip install --upgrade pip
if [[ -f "${REQ_FILE}" ]]; then
  pip install -r "${REQ_FILE}"
fi

# 파이썬 실행
python "${PY_SCRIPT}" "${INPUT_JSON}" "${OUTPUT_JSON}"

# 출력 파일 생성 확인 (옵션 A에서는 Go가 읽고 삭제)
if [[ ! -f "${OUTPUT_JSON}" ]]; then
  echo "ERROR: output json not created: ${OUTPUT_JSON}" >&2
  exit 4
fi
