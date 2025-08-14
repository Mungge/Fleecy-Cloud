from flask import Blueprint, jsonify, request
from datetime import datetime
import logging

from services.fl_service import FederatedLearningService

logger = logging.getLogger(__name__)

fl_bp = Blueprint('fl', __name__)
fl_service = FederatedLearningService()

@fl_bp.route('/api/fl/execute', methods=['POST'])
def execute_federated_learning():
    """연합학습 코드와 환경 변수를 받아서 지정된 VM에서 실행"""
    try:
        # 요청 데이터 검증
        if not request.is_json:
            return jsonify({'error': 'Content-Type must be application/json'}), 400
            
        data = request.get_json()
        
        # 필수 필드 검증
        required_fields = ['vm_id', 'fl_code', 'env_config']
        missing_fields = [field for field in required_fields if field not in data]
        if missing_fields:
            return jsonify({
                'error': f'Missing required fields: {missing_fields}',
                'required_fields': required_fields
            }), 400
        
        vm_id = data['vm_id']
        fl_code = data['fl_code']  # 연합학습 Python 코드 (문자열)
        env_config = data['env_config']  # 환경 변수 설정 (딕셔너리)
        
        # 선택적 필드
        entry_point = data.get('entry_point', 'main.py')  # 실행할 파일명
        requirements = data.get('requirements', [])  # 추가 패키지 리스트
        
        # 연합학습 실행
        result = fl_service.execute_fl_task(
            vm_id=vm_id,
            fl_code=fl_code,
            env_config=env_config,
            entry_point=entry_point,
            requirements=requirements
        )
        
        status_code = 201 if result['success'] else 500
        return jsonify(result), status_code
        
    except Exception as e:
        logger.error(f"Error executing federated learning: {str(e)}")
        return jsonify({'error': 'Failed to execute federated learning'}), 500

@fl_bp.route('/api/fl/logs/<string:task_id>', methods=['GET'])
def get_fl_logs(task_id):
    """연합학습 작업 로그 조회"""
    try:
        vm_id = request.args.get('vm_id')
        if not vm_id:
            return jsonify({'error': 'vm_id parameter is required'}), 400
        
        result = fl_service.get_task_logs(task_id, vm_id)
        
        status_code = 200 if result['success'] else 404
        return jsonify(result), status_code
        
    except Exception as e:
        logger.error(f"Error getting FL logs for {task_id}: {str(e)}")
        return jsonify({'error': 'Failed to retrieve logs'}), 500

@fl_bp.route('/api/fl/status/<string:vm_id>', methods=['GET'])
def get_task_status(vm_id):
    """현재 실행 중인 작업 상태"""
    # TODO: 실제 작업 상태 조회 로직 구현
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
