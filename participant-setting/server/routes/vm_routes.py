from flask import Blueprint, jsonify
from datetime import datetime
import logging

from utils.openstack import get_openstack_vmList

logger = logging.getLogger(__name__)

vm_bp = Blueprint('vm', __name__)

@vm_bp.route('/api/vms', methods=['GET'])
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
