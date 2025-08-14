#!/usr/bin/env bash
set -euo pipefail

# 위치
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." &>/dev/null && pwd)"

VENV_DIR="${PROJECT_ROOT}/.venv"
PY_SCRIPT="${SCRIPT_DIR}/aggregator_optimization.py"

TEMP_DIR="${PROJECT_ROOT}/temp"
INPUT_DIR="${TEMP_DIR}/input"
OUTPUT_DIR="${TEMP_DIR}/output"

# requirements.txt 우선순위: scripts/ → 프로젝트 루트
REQ_FILE="${SCRIPT_DIR}/requirements.txt"
[[ -f "${REQ_FILE}" ]] || REQ_FILE="${PROJECT_ROOT}/requirements.txt"

# 사용법: run_optimizer.sh <input_json> <output_json>
if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <input_json> <output_json>" >&2
  exit 2
fi
INPUT_JSON="$1"
OUTPUT_JSON="$2"


# Python 3.12 강제 선택
choose_py312() {
  if command -v python3.12 >/dev/null 2>&1; then
    echo "python3.12"
    return
  fi
  if command -v python3 >/dev/null 2>&1; then
    local ver
    ver="$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')"
    if [[ "${ver}" == "3.12" ]]; then
      echo "python3"
      return
    fi
  fi
  echo "ERROR: Python 3.12 is required but not found." >&2
  echo "Install Python 3.12 (e.g., apt/brew/pyenv) and retry." >&2
  exit 3
}
PYTHON_CMD="$(choose_py312)"


# venv 생성/활성화
if [[ ! -d "${VENV_DIR}" ]]; then
  "${PYTHON_CMD}" -m venv "${VENV_DIR}"
fi
# shellcheck disable=SC1091
source "${VENV_DIR}/bin/activate"

# 의존성 설치
python -m pip install --upgrade pip
[[ -f "${REQ_FILE}" ]] && pip install -r "${REQ_FILE}"

# 파이썬 실행
python "${PY_SCRIPT}" "${INPUT_JSON}" "${OUTPUT_JSON}"

# 출력 확인
if [[ ! -f "${OUTPUT_JSON}" ]]; then
  echo "ERROR: output json not created: ${OUTPUT_JSON}" >&2
  exit 4
fi
