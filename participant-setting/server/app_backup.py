from flask import Flask, jsonify, request
from flask_cors import CORS
import logging
import os
from datetime import datetime
import subprocess
import ipaddress
import ast
from typing import List, Dict, Optional
import paramiko
import json
from dotenv import load_dotenv

# 환경변수 로드 (python-dotenv가 설치된 경우)
try:
    load_dotenv()
except ImportError:
    pass

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

# VM 목록 조회 API 엔드포인트
@app.route('/api/vms', methods=['GET'])
def get_vm_list():
    """VM ID와 Floating IP 목록 조회 API"""
    try:
        vm_list = get_openstack_vmList()
        if vm_list is None:
            vm_list = []
        
        return jsonify({
            'status': 'success',
            'count': len(vm_list),
            'vms': vm_list,
            'timestamp': datetime.now().isoformat()
        })
    except Exception as e:
        logger.error(f"Error in get_vm_list endpoint: {str(e)}")
        return jsonify({'error': 'Failed to retrieve VM list'}), 500

# 연합학습 실행 엔드포인트
@app.route('/api/fl/execute', methods=['POST'])
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
        
        # OpenStack에서 VM 목록 조회
        vm_list = get_openstack_vmList()
        
        # 해당 VM ID가 존재하는지 확인
        target_vm = None
        for vm in vm_list:
            if vm.get('id') == vm_id:
                target_vm = vm
                break
        
        if not target_vm:
            return jsonify({
                'error': f'VM with ID {vm_id} not found',
                'vm_id': vm_id
            }), 404
        
        # Floating IP가 있는지 확인
        floating_ip = target_vm.get('floating_ip')
        if not floating_ip:
            return jsonify({
                'error': f'VM {vm_id} has no floating IP assigned',
                'vm_id': vm_id,
                'vm_info': target_vm
            }), 400
        
        # 작업 ID 생성
        task_id = f"fl-task-{datetime.now().strftime('%Y%m%d-%H%M%S')}"
        
        logger.info(f"Starting FL execution {task_id} on VM {vm_id} (IP: {floating_ip})")
        
        # SSH를 통해 연합학습 코드 전송 및 실행
        execution_result = deploy_and_execute_fl_code(
            floating_ip=floating_ip,
            task_id=task_id,
            fl_code=fl_code,
            env_config=env_config,
            entry_point=entry_point,
            requirements=requirements
        )
        
        # 응답 생성
        response = {
            'task_id': task_id,
            'vm_id': vm_id,
            'target_ip': floating_ip,
            'submitted_at': datetime.now().isoformat(),
            'entry_point': entry_point,
            'status': 'started' if execution_result['success'] else 'failed',
            'message': execution_result.get('message', ''),
        }
        
        if execution_result['success']:
            response['ssh_output'] = execution_result.get('output', '')
            response['remote_path'] = execution_result.get('remote_path', '')
        else:
            response['error'] = execution_result.get('error', '')
        
        status_code = 201 if execution_result['success'] else 500
        return jsonify(response), status_code
        
    except Exception as e:
        logger.error(f"Error executing federated learning: {str(e)}")
        return jsonify({'error': 'Failed to execute federated learning'}), 500

# 연합학습 작업 로그 조회 엔드포인트
@app.route('/api/fl/logs/<string:task_id>', methods=['GET'])
def get_fl_logs(task_id):
    """연합학습 작업 로그 조회"""
    try:
        vm_id = request.args.get('vm_id')
        if not vm_id:
            return jsonify({'error': 'vm_id parameter is required'}), 400
        
        # VM 정보 조회
        vm_list = get_openstack_vmList()
        target_vm = next((vm for vm in vm_list if vm.get('id') == vm_id), None)
        
        if not target_vm:
            return jsonify({'error': f'VM {vm_id} not found'}), 404
        
        floating_ip = target_vm.get('floating_ip')
        if not floating_ip:
            return jsonify({'error': f'VM {vm_id} has no floating IP'}), 400
        
        # SSH로 로그 파일 조회
        ssh_user = os.environ.get('SSH_USER', 'ubuntu')
        ssh_key_path = os.environ.get('SSH_KEY_PATH', '~/.ssh/id_rsa')
        ssh_port = int(os.environ.get('SSH_PORT', '22'))
        
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        client.connect(
            hostname=floating_ip,
            port=ssh_port,
            username=ssh_user,
            key_filename=os.path.expanduser(ssh_key_path),
            timeout=10
        )
        
        # 로그 파일 읽기
        log_path = f'/tmp/fl-workspace/{task_id}/{task_id}.log'
        stdin, stdout, stderr = client.exec_command(f'cat {log_path}')
        log_content = stdout.read().decode('utf-8')
        error = stderr.read().decode('utf-8')
        
        # 프로세스 상태 확인
        stdin, stdout, stderr = client.exec_command(f'ps aux | grep {task_id} | grep -v grep')
        process_status = stdout.read().decode('utf-8')
        
        client.close()
        
        return jsonify({
            'task_id': task_id,
            'vm_id': vm_id,
            'log_content': log_content,
            'process_running': bool(process_status.strip()),
            'process_info': process_status.strip(),
            'error': error if error else None,
            'timestamp': datetime.now().isoformat()
        })
        
    except Exception as e:
        logger.error(f"Error getting FL logs for {task_id}: {str(e)}")
        return jsonify({'error': 'Failed to retrieve logs'}), 500

# 작업 상태 엔드포인트
@app.route('/api/fl/status/<string:vm_id>', methods=['GET'])
def get_task_status(vm_id):
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

def deploy_and_execute_fl_code(floating_ip: str, task_id: str, fl_code: str, 
                              env_config: dict, entry_point: str, requirements: list) -> dict:
    """SSH를 통해 연합학습 코드를 VM에 배포하고 실행"""
    try:
        # SSH 연결 설정
        ssh_user = os.environ.get('SSH_USER', 'ubuntu')
        ssh_key_path = os.environ.get('SSH_KEY_PATH', '~/.ssh/id_rsa')
        ssh_port = int(os.environ.get('SSH_PORT', '22'))
        
        # SSH 클라이언트 생성
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        
        # SSH 연결
        key_path = os.path.expanduser(ssh_key_path)
        client.connect(
            hostname=floating_ip,
            port=ssh_port,
            username=ssh_user,
            key_filename=key_path,
            timeout=10
        )
        
        # SFTP 클라이언트 생성
        sftp = client.open_sftp()
        
        # 원격 작업 디렉토리 생성
        remote_work_dir = f'/tmp/fl-workspace/{task_id}'
        client.exec_command(f'mkdir -p {remote_work_dir}')
        
        # 1. 연합학습 코드 파일 업로드
        remote_code_path = f'{remote_work_dir}/{entry_point}'
        with sftp.open(remote_code_path, 'w') as f:
            f.write(fl_code)
        
        # 2. 환경 변수 설정 파일 생성 (.env 파일)
        env_content = []
        for key, value in env_config.items():
            env_content.append(f'{key}={value}')
        env_file_content = '\n'.join(env_content)
        
        remote_env_path = f'{remote_work_dir}/.env'
        with sftp.open(remote_env_path, 'w') as f:
            f.write(env_file_content)
        
        # 3. requirements.txt 파일 생성 (필요한 경우)
        if requirements:
            requirements_content = '\n'.join(requirements)
            remote_req_path = f'{remote_work_dir}/requirements.txt'
            with sftp.open(remote_req_path, 'w') as f:
                f.write(requirements_content)
            
            # 패키지 설치
            install_cmd = f'cd {remote_work_dir} && pip install -r requirements.txt'
            stdin, stdout, stderr = client.exec_command(install_cmd)
            install_output = stdout.read().decode('utf-8')
            install_error = stderr.read().decode('utf-8')
            
            if install_error:
                logger.warning(f"Package installation warnings: {install_error}")
        
        sftp.close()
        
        # 4. 연합학습 코드 실행
        # 환경 변수를 source하고 Python 스크립트 실행
        execute_cmd = (
            f'cd {remote_work_dir} && '
            f'export $(cat .env | xargs) && '
            f'nohup python3 {entry_point} > {task_id}.log 2>&1 &'
        )
        
        stdin, stdout, stderr = client.exec_command(execute_cmd)
        output = stdout.read().decode('utf-8')
        error = stderr.read().decode('utf-8')
        
        # 실행 확인 (프로세스가 시작되었는지 체크)
        check_cmd = f'ps aux | grep {entry_point} | grep -v grep'
        stdin, stdout, stderr = client.exec_command(check_cmd)
        process_check = stdout.read().decode('utf-8')
        
        client.close()
        
        if error and 'nohup' not in error:  # nohup 메시지는 정상
            logger.error(f"Error executing FL code on {floating_ip}: {error}")
            return {
                'success': False,
                'error': error,
                'message': 'Failed to execute federated learning code'
            }
        
        logger.info(f"Successfully deployed and started FL task {task_id} on {floating_ip}")
        return {
            'success': True,
            'output': output,
            'remote_path': remote_work_dir,
            'message': f'Federated learning code deployed and started in {remote_work_dir}',
            'process_check': process_check.strip() if process_check else 'Process check unavailable'
        }
        
    except Exception as e:
        logger.error(f"Failed to deploy FL code to {floating_ip}: {str(e)}")
        return {
            'success': False,
            'error': str(e),
            'message': 'Failed to deploy and execute federated learning code'
        }

def get_openstack_vmList() -> List[Dict[str, Optional[str]]]:
    """Openstack에서 VM ID와 Floating IP 정보를 가져오는 함수"""
    cmd = (
        "cd ~/devstack && "
        "source openrc admin demo && "
        "openstack server list --format value --column ID --column Networks"
    )
    result = subprocess.run(
        cmd,
        shell=True,
        executable='/bin/bash',
        capture_output=True,
        text=True,
        timeout=15
    )

    if result.returncode != 0:
        return []

    lines = [line.strip() for line in result.stdout.strip().splitlines() if line.strip()]
    vm_list: List[Dict[str, Optional[str]]] = []

    for line in lines:
        parts = line.split(None, 1)
        if len(parts) < 2:
            continue
        vm_id, networks_str = parts[0].strip(), parts[1].strip()

        all_ips: List[str] = []
        parsed = None
        try:
            parsed = ast.literal_eval(networks_str)
        except Exception:
            parsed = None

        if isinstance(parsed, dict):
            for v in parsed.values():
                if isinstance(v, list):
                    for item in v:
                        if isinstance(item, str):
                            all_ips.append(item.strip())
        else:
            tokens = [t.strip() for t in networks_str.split(',') if t.strip()]
            for t in tokens:
                if '=' in t:
                    all_ips.append(t.split('=')[-1].strip())
                else:
                    all_ips.append(t)

        floating_ip = all_ips[-1] if all_ips else None

        vm_info = {
            'id': vm_id,
            'floating_ip': floating_ip
        }
        vm_list.append(vm_info)

    return vm_list

# 에러 핸들러
@app.errorhandler(404)
def not_found(error):
    return jsonify({'error': 'Endpoint not found'}), 404

@app.errorhandler(500)
def internal_error(error):
    return jsonify({'error': 'Internal server error'}), 500

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 5000))
    host = os.environ.get('HOST', '0.0.0.0')
    debug = os.environ.get('DEBUG', 'False').lower() == 'true'
    
    logger.info(f"Starting Fleecy Cloud Participant Server on {host}:{port}")
    app.run(host=host, port=port, debug=debug)

