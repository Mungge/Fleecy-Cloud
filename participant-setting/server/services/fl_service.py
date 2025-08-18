import logging
from datetime import datetime
from typing import Dict, Optional

from utils.openstack import get_openstack_vmList
from services.ssh_service import SSHService

logger = logging.getLogger(__name__)

class FederatedLearningService:
    def __init__(self):
        self.ssh_service = SSHService()

    def execute_fl_task(self, vm_id: str, fl_code: str, env_config: dict, 
                       entry_point: str = 'main.py', requirements: Optional[list] = None) -> Dict:
        """연합학습 작업 실행"""
        if requirements is None:
            requirements = []
            
        vm_list = get_openstack_vmList()
        
        # 해당 VM ID가 존재하는지 확인
        target_vm = self._find_vm_by_id(vm_list, vm_id)
        if not target_vm:
            return {
                'success': False,
                'error': f'VM with ID {vm_id} not found',
                'vm_id': vm_id
            }
        
        # Floating IP가 있는지 확인
        floating_ip = target_vm.get('floating_ip')
        if not floating_ip:
            return {
                'success': False,
                'error': f'VM {vm_id} has no floating IP assigned',
                'vm_id': vm_id,
                'vm_info': target_vm
            }
        
        # 작업 ID 생성
        task_id = f"fl-task-{datetime.now().strftime('%Y%m%d-%H%M%S')}"
        
        logger.info(f"Starting FL execution {task_id} on VM {vm_id} (IP: {floating_ip})")
        
        # SSH를 통해 연합학습 코드 전송 및 실행
        execution_result = self.ssh_service.deploy_and_execute_fl_code(
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
            'success': execution_result['success'],
            'message': execution_result.get('message', ''),
        }
        
        if execution_result['success']:
            response['ssh_output'] = execution_result.get('output', '')
            response['remote_path'] = execution_result.get('remote_path', '')
        else:
            response['error'] = execution_result.get('error', '')
        
        return response

    def get_task_logs(self, task_id: str, vm_id: str) -> Dict:
        """연합학습 작업 로그 조회"""
        # VM 정보 조회
        vm_list = get_openstack_vmList()
        target_vm = self._find_vm_by_id(vm_list, vm_id)
        
        if not target_vm:
            return {
                'success': False,
                'error': f'VM {vm_id} not found'
            }
        
        floating_ip = target_vm.get('floating_ip')
        if not floating_ip:
            return {
                'success': False,
                'error': f'VM {vm_id} has no floating IP'
            }
        
        # SSH로 로그 조회
        log_result = self.ssh_service.get_logs(floating_ip, task_id)
        
        if log_result['success']:
            return {
                'success': True,
                'task_id': task_id,
                'vm_id': vm_id,
                'log_content': log_result['log_content'],
                'process_running': log_result['process_running'],
                'process_info': log_result['process_info'],
                'error': log_result.get('error'),
                'timestamp': datetime.now().isoformat()
            }
        else:
            return {
                'success': False,
                'error': log_result['error']
            }

    def _find_vm_by_id(self, vm_list: list, vm_id: str) -> Optional[Dict]:
        """VM ID로 VM 정보 찾기"""
        for vm in vm_list:
            if vm.get('id') == vm_id:
                return vm
        return None
