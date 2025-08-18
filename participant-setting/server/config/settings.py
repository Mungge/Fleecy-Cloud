import os
from dotenv import load_dotenv # type: ignore

# 환경변수 로드
try:
    load_dotenv()
except ImportError:
    pass

class Config:
    # Flask 설정
    HOST = os.environ.get('HOST', '0.0.0.0')
    PORT = int(os.environ.get('PORT', 5000))
    DEBUG = os.environ.get('DEBUG', 'False').lower() == 'true'
    
    # SSH 설정
    SSH_USER = os.environ.get('SSH_USER', 'ubuntu')
    SSH_KEY_PATH = os.environ.get('SSH_KEY_PATH', '~/.ssh/key.pem')
    SSH_PORT = int(os.environ.get('SSH_PORT', '22'))
    
    # OpenStack 설정
    OPENSTACK_TIMEOUT = 15
    DEVSTACK_PATH = "~/devstack"
