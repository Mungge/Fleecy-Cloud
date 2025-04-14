
# Aggregator 모듈 초기화
from .fedavg import FedAvg
from .fedprox import FedProx
from .fedadagrad import FedAdagrad
from .fedadam import FedAdam
from .fedyogi import FedYogi

__all__ = [
    'FedAvg',
    'FedProx',
    'FedAdagrad',
    'FedAdam',
    'FedYogi',
]