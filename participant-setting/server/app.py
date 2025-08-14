from flask import Flask
from flask_cors import CORS
import logging

from config.settings import Config
from routes.main_routes import main_bp
from routes.vm_routes import vm_bp
from routes.fl_routes import fl_bp

def create_app():
    """Flask 애플리케이션 팩토리"""
    app = Flask(__name__)
    
    # CORS 설정
    CORS(app)
    
    # 로깅 설정
    logging.basicConfig(level=logging.INFO)
    
    # Blueprint 등록
    app.register_blueprint(main_bp)
    app.register_blueprint(vm_bp)
    app.register_blueprint(fl_bp)
    
    return app

if __name__ == '__main__':
    app = create_app()
    logger = logging.getLogger(__name__)
    
    logger.info(f"Starting Fleecy Cloud Participant Server on {Config.HOST}:{Config.PORT}")
    app.run(host=Config.HOST, port=Config.PORT, debug=Config.DEBUG)
