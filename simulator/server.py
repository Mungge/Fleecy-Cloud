import os
import time
import json
from pathlib import Path
import flwr as fl

from aggregators.fedavg import FedAvg
from aggregators.fedprox import FedProx
from aggregators.fedadagrad import FedAdagrad
from aggregators.fedadam import FedAdam
from aggregators.fedyogi import FedYogi
from monitoring.monitor import CPUMonitor

# 환경 변수 가져오기
AGGREGATOR_TYPE = os.environ.get("AGGREGATOR_TYPE").lower()
NUM_ROUNDS = int(os.environ.get("NUM_ROUNDS", "10"))
MIN_CLIENTS = int(os.environ.get("MIN_CLIENTS", "3"))
RESULTS_DIR = os.environ.get("RESULTS_DIR", "./results")

# 결과 디렉토리 생성
Path(RESULTS_DIR).mkdir(parents=True, exist_ok=True)

# CPU 모니터링 인스턴스 초기화
cpu_monitor = CPUMonitor(
    container_name="fl-server",
    output_file=f"{RESULTS_DIR}/server_cpu_{AGGREGATOR_TYPE}.csv"
)

def get_aggregator_strategy():
    """선택된 Aggregation 전략 반환"""
    if AGGREGATOR_TYPE == "fedavg":
        return FedAvg(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
        )
    elif AGGREGATOR_TYPE == "fedprox":
        return FedProx(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
            proximal_mu=0.01,  # Proximal term coefficient
        )
    elif AGGREGATOR_TYPE == "fedadagrad":
        return FedAdagrad(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
            eta=0.1,  # Server learning rate
            eta_l=0.1,  # Client learning rate
            tau=1e-9,  # Controls stability
        )
    elif AGGREGATOR_TYPE == "fedadam":
        return FedAdam(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
            eta=0.1,  # Server learning rate
            eta_l=0.1,  # Client learning rate
            beta_1=0.9,
            beta_2=0.99,
            tau=1e-3,
        )
    elif AGGREGATOR_TYPE == "fedyogi":
        return FedYogi(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
            eta=0.1,  # Server learning rate
            eta_l=0.1,  # Client learning rate
            beta_1=0.9,
            beta_2=0.99,
            tau=1e-3,
        )
    else:
        print(f"Unknown aggregator: {AGGREGATOR_TYPE}, using FedAvg")
        return FedAvg(
            min_fit_clients=MIN_CLIENTS,
            min_evaluate_clients=MIN_CLIENTS,
            min_available_clients=MIN_CLIENTS,
        )

def server_evaluation_callback(server_round: int, results, failures):
    """서버 평가 콜백 - CPU 사용량 기록 및 라운드별 정확도 기록"""
    if results:
        accuracy_aggregated = sum([r[1].metrics["accuracy"] * r[1].num_examples for r in results]) / sum([r[1].num_examples for r in results])
        print(f"Round {server_round} accuracy: {accuracy_aggregated:.4f}")
        
        # 결과 저장
        with open(f"{RESULTS_DIR}/{AGGREGATOR_TYPE}_results.json", "a") as f:
            json.dump({
                "round": server_round,
                "accuracy": float(accuracy_aggregated),
                "timestamp": time.time(),
            }, f)
            f.write("\n")
    
    return None

def main():
    """Flower 서버 시작"""
    print(f"Starting Flower server with {AGGREGATOR_TYPE} aggregator")
    print(f"Number of rounds: {NUM_ROUNDS}")
    print(f"Minimum clients: {MIN_CLIENTS}")
    
    # CPU 모니터링 시작
    cpu_monitor.start()
    
    # 서버 전략 설정
    strategy = get_aggregator_strategy()
    
    # Flower 서버 설정 및 시작
    fl.server.start_server(
        server_address="0.0.0.0:9090",
        config=fl.server.ServerConfig(num_rounds=NUM_ROUNDS),
        strategy=strategy,
        grpc_max_message_length=1024*1024*1024,  # 1GB
    )
    
    # CPU 모니터링 중지
    cpu_monitor.stop()

if __name__ == "__main__":
    main()