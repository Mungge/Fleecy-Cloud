import subprocess
import ast
from typing import List, Dict, Optional
import logging

from config.settings import Config

logger = logging.getLogger(__name__)

def get_openstack_vmList() -> List[Dict[str, Optional[str]]]:
    """Openstack에서 VM ID와 Floating IP 정보를 가져오는 함수"""
    cmd = (
        f"cd {Config.DEVSTACK_PATH} && "
        "source openrc admin demo && "
        "openstack server list --format value --column ID --column Networks"
    )
    
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            executable='/bin/bash',
            capture_output=True,
            text=True,
            timeout=Config.OPENSTACK_TIMEOUT
        )

        if result.returncode != 0:
            logger.error(f"OpenStack command failed: {result.stderr}")
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
        
    except subprocess.TimeoutExpired:
        logger.error("OpenStack command timed out")
        return []
    except Exception as e:
        logger.error(f"Error getting VM list: {str(e)}")
        return []
