import paramiko
import os
import logging
from typing import Dict

from config.settings import Config

logger = logging.getLogger(__name__)

class SSHService:
    def __init__(self):
        self.ssh_user = Config.SSH_USER
        self.ssh_key_path = Config.SSH_KEY_PATH
        self.ssh_port = Config.SSH_PORT

    def deploy_and_execute_fl_code(self, floating_ip: str, task_id: str, fl_code: str, 
                                  env_config: dict, entry_point: str, requirements: list) -> dict:
        """SSH를 통해 연합학습 코드를 VM에 배포하고 실행"""
        try:
            # SSH 클라이언트 생성
            client = paramiko.SSHClient()
            client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            
            # SSH 연결
            key_path = os.path.expanduser(self.ssh_key_path)
            client.connect(
                hostname=floating_ip,
                port=self.ssh_port,
                username=self.ssh_user,
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

    def get_logs(self, floating_ip: str, task_id: str) -> Dict:
        """SSH를 통해 원격 로그 파일 조회"""
        try:
            client = paramiko.SSHClient()
            client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            client.connect(
                hostname=floating_ip,
                port=self.ssh_port,
                username=self.ssh_user,
                key_filename=os.path.expanduser(self.ssh_key_path),
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
            
            return {
                'success': True,
                'log_content': log_content,
                'process_running': bool(process_status.strip()),
                'process_info': process_status.strip(),
                'error': error if error else None
            }
            
        except Exception as e:
            logger.error(f"Error getting logs from {floating_ip}: {str(e)}")
            return {
                'success': False,
                'error': str(e)
            }
