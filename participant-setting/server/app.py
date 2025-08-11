from flask import Flask, jsonify, request
from flask_cors import CORS
import logging
import os
from datetime import datetime

# Flask 앱 초기화
app = Flask(__name__)
CORS(app)  # CORS 설정으로 크로스 오리진 요청 허용

# 로깅 설정
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 기본 라우트
@app.route('/')
def home():
    """메인 페이지"""
    return jsonify({
        'message': 'Fleecy Cloud Participant Server',
        'status': 'running',
        'timestamp': datetime.now().isoformat()
    })

# 헬스 체크 엔드포인트
@app.route('/health')
def health_check():
    """서버 상태 확인"""
    return jsonify({
        'status': 'healthy',
        'timestamp': datetime.now().isoformat()
    })

# 참가자 정보 엔드포인트
@app.route('/api/participant/info', methods=['GET'])
def get_participant_info():
    """참가자 정보 반환"""
    return jsonify({
        'participant_id': 'participant-001',
        'status': 'active',
        'capabilities': {
            'cpu_cores': 4,
            'memory_gb': 8,
            'storage_gb': 100
        },
        'timestamp': datetime.now().isoformat()
    })

# 모니터링 데이터 엔드포인트
@app.route('/api/monitoring/metrics', methods=['GET'])
def get_monitoring_metrics():
    """모니터링 메트릭 반환"""
    # 실제 환경에서는 시스템 메트릭을 수집해야 합니다
    return jsonify({
        'cpu_usage': 25.5,
        'memory_usage': 60.2,
        'disk_usage': 45.8,
        'network_io': {
            'bytes_sent': 1024000,
            'bytes_received': 2048000
        },
        'timestamp': datetime.now().isoformat()
    })

# 작업 상태 엔드포인트
@app.route('/api/tasks/status', methods=['GET'])
def get_task_status():
    """현재 실행 중인 작업 상태"""
    return jsonify({
        'active_tasks': [
            {
                'task_id': 'fl-training-001',
                'type': 'federated_learning',
                'status': 'running',
                'progress': 75.5,
                'started_at': '2025-08-11T10:30:00Z'
            }
        ],
        'completed_tasks': 5,
        'failed_tasks': 0,
        'timestamp': datetime.now().isoformat()
    })

# 작업 제출 엔드포인트
@app.route('/api/tasks/submit', methods=['POST'])
def submit_task():
    """새로운 작업 제출"""
    try:
        task_data = request.get_json()
        
        if not task_data:
            return jsonify({'error': 'No task data provided'}), 400
        
        # 실제 환경에서는 작업을 큐에 추가하거나 처리 로직을 구현해야 합니다
        response = {
            'task_id': f"task-{datetime.now().strftime('%Y%m%d-%H%M%S')}",
            'status': 'accepted',
            'message': 'Task submitted successfully',
            'submitted_at': datetime.now().isoformat(),
            'task_data': task_data
        }
        
        logger.info(f"Task submitted: {response['task_id']}")
        return jsonify(response), 201
        
    except Exception as e:
        logger.error(f"Error submitting task: {str(e)}")
        return jsonify({'error': 'Failed to submit task'}), 500

# 에러 핸들러
@app.errorhandler(404)
def not_found(error):
    return jsonify({'error': 'Endpoint not found'}), 404

@app.errorhandler(500)
def internal_error(error):
    return jsonify({'error': 'Internal server error'}), 500

if __name__ == '__main__':
    # 환경 변수에서 포트 설정 (기본값: 5000)
    port = int(os.environ.get('PORT', 5000))
    host = os.environ.get('HOST', '0.0.0.0')
    debug = os.environ.get('DEBUG', 'False').lower() == 'true'
    
    logger.info(f"Starting Fleecy Cloud Participant Server on {host}:{port}")
    app.run(host=host, port=port, debug=debug)
