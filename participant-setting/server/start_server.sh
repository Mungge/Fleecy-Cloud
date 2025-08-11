#!/bin/bash

# Fleecy Cloud Participant Server 시작 스크립트

echo "Fleecy Cloud Participant Server를 시작합니다..."

# 가상환경 확인 및 생성
if [ ! -d "venv" ]; then
    echo "Python 가상환경을 생성합니다..."
    python3 -m venv venv
fi

# 가상환경 활성화
echo "가상환경을 활성화합니다..."
source venv/bin/activate

# 의존성 설치
echo "의존성을 설치합니다..."
pip install -r requirements.txt

# 환경 변수 파일 확인
if [ ! -f ".env" ]; then
    echo ".env 파일이 없습니다. .env.example을 복사합니다..."
    cp .env.example .env
    echo ".env 파일을 수정한 후 다시 실행해주세요."
    exit 1
fi

# 서버 시작
echo "서버를 시작합니다..."
python app.py
