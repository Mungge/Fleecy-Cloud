#!/bin/bash

# region_latency_test.sh
# AWS 리전 간 latency 측정 스크립트

# 측정 결과를 저장할 파일명
RESULTS_FILE="aws_region_latency_$(hostname)_$(date +%Y%m%d_%H%M%S).csv"

# 측정할 대상 인스턴스 목록 - 여기에 각 리전 인스턴스의 IP 주소를 입력
# 형식: "IP주소,리전명" (쉼표로 구분)
TARGETS=(
  "54.180.100.196,ap-northeast-2" \
  "13.231.217.241,ap-northeast-1" \
  "52.77.229.124,ap-southeast-1" \
  "3.80.156.106,us-east-1"
)

# CSV 헤더 생성
echo "Source,SourceRegion,Target,TargetRegion,Timestamp,MinLatency(ms),AvgLatency(ms),MaxLatency(ms),StdDev(ms),PacketLoss(%)" > $RESULTS_FILE

# 현재 인스턴스의 리전 정보 (수동 설정)
SOURCE_REGION="ap-northeast-2"  # 현재 사용 중인 리전으로 변경
SOURCE_HOSTNAME=$(hostname)

# 각 대상에 대해 ping 테스트 실행
for target_info in "${TARGETS[@]}"; do
	# 대상 IP와 리전 추출
  TARGET_IP=$(echo $target_info | cut -d',' -f1)
  TARGET_REGION=$(echo $target_info | cut -d',' -f2)

  echo "리전 $SOURCE_REGION에서 리전 $TARGET_REGION($TARGET_IP)으로 latency 측정 중..."
  
  # 현재 시간 기록
  TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

	# ping 테스트 실행 (5개 패킷), 타임아웃 2초
  echo "Ping 명령어 시작: ping -c 30 $TARGET_IP"
  PING_RESULT=$(ping -c 5 -W 2 $TARGET_IP)
  echo "Ping 명령어 완료"

  # ping 성공 여부 확인
  if [ $? -eq 0 ]; then
    PACKET_LOSS=$(echo "$PING_RESULT" | grep -oP '\d+(?=% packet loss)')
    
    # RTT 통계 추출
    RTT_STATS=$(echo "$PING_RESULT" | grep 'rtt')
    MIN_RTT=$(echo "$RTT_STATS" | awk -F/ '{print $4}')
    AVG_RTT=$(echo "$RTT_STATS" | awk -F/ '{print $5}')
    MAX_RTT=$(echo "$RTT_STATS" | awk -F/ '{print $6}')
    STDDEV=$(echo "$RTT_STATS" | awk -F/ '{print $7}' | awk '{print $1}')

		# 결과를 CSV 파일에 추가
    echo "$SOURCE_HOSTNAME,$SOURCE_REGION,$TARGET_IP,$TARGET_REGION,$TIMESTAMP,$MIN_RTT,$AVG_RTT,$MAX_RTT,$STDDEV,$PACKET_LOSS" >> $RESULTS_FILE
    
    # 화면에 결과 출력
    echo "결과: 최소=${MIN_RTT}ms, 평균=${AVG_RTT}ms, 최대=${MAX_RTT}ms, 패킷손실=${PACKET_LOSS}%"
  else
    # ping 실패 시 오류 기록
    echo "$SOURCE_HOSTNAME,$SOURCE_REGION,$TARGET_IP,$TARGET_REGION,$TIMESTAMP,Error,Error,Error,Error,100" >> $RESULTS_FILE
    echo "오류: ping 실패"
  fi

  echo "----------------------------------------"
done

echo "모든 측정 완료. 결과는 $RESULTS_FILE 파일에 저장되었습니다."

# 결과 요약 보여주기
echo ""
echo "== 리전 간 latency 측정 결과 요약 =="
echo "출발지 리전: $SOURCE_REGION"
echo ""
echo "대상 리전          | 평균 Latency (ms) | 패킷 손실 (%)"
echo "--------------------|-------------------|---------------"

for target_info in "${TARGETS[@]}"; do
  TARGET_IP=$(echo $target_info | cut -d',' -f1)
  TARGET_REGION=$(echo $target_info | cut -d',' -f2)
  
  # CSV 파일에서 해당 대상에 대한 결과 가져오기
  RESULT_LINE=$(grep "$TARGET_IP" $RESULTS_FILE | tail -1)

  if [[ $RESULT_LINE == *Error* ]]; then
    echo "$(printf '%-20s' $TARGET_REGION) | 연결 실패          | 100%"
  else
    AVG_RTT=$(echo "$RESULT_LINE" | cut -d',' -f7)
    PACKET_LOSS=$(echo "$RESULT_LINE" | cut -d',' -f10)
    echo "$(printf '%-20s' $TARGET_REGION) | $(printf '%-19s' $AVG_RTT) | $PACKET_LOSS%"
  fi
done