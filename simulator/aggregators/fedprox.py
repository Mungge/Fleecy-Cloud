import flwr as fl
from typing import Dict, List, Tuple, Optional, Union
from flwr.common import (
    Scalar,
)
import numpy as np

class FedProx(fl.server.strategy.FedAvg):
    """FedProx 전략 구현 (FedAvg 기반에 Proximal term 추가)"""
    
    def __init__(
        self,
        *,
        fraction_fit: float = 1.0,
        fraction_evaluate: float = 1.0,
        min_fit_clients: int = 2,
        min_evaluate_clients: int = 2,
        min_available_clients: int = 2,
        evaluate_fn=None,
        on_fit_config_fn=None,
        on_evaluate_config_fn=None,
        accept_failures: bool = True,
        initial_parameters=None,
        fit_metrics_aggregation_fn=None,
        evaluate_metrics_aggregation_fn=None,
        proximal_mu: float = 0.01,  # Proximal term 계수
    ):
        self.proximal_mu = proximal_mu
        
        # FedProx는 fit_config에 mu 값을 포함하도록 on_fit_config_fn을 재정의
        if on_fit_config_fn is None:
            on_fit_config_fn = self.default_on_fit_config
        else:
            original_on_fit_config_fn = on_fit_config_fn
            on_fit_config_fn = lambda round_num: {
                **original_on_fit_config_fn(round_num),
                "proximal_mu": self.proximal_mu,
            }
        
        super().__init__(
            fraction_fit=fraction_fit,
            fraction_evaluate=fraction_evaluate,
            min_fit_clients=min_fit_clients,
            min_evaluate_clients=min_evaluate_clients,
            min_available_clients=min_available_clients,
            evaluate_fn=evaluate_fn,
            on_fit_config_fn=on_fit_config_fn,
            on_evaluate_config_fn=on_evaluate_config_fn,
            accept_failures=accept_failures,
            initial_parameters=initial_parameters,
            fit_metrics_aggregation_fn=fit_metrics_aggregation_fn,
            evaluate_metrics_aggregation_fn=evaluate_metrics_aggregation_fn,
        )
    
    def default_on_fit_config(self, round_num: int) -> Dict[str, Scalar]:
        """클라이언트에 전달할 기본 학습 설정"""
        return {
            "proximal_mu": self.proximal_mu,
        }
    
    def __repr__(self) -> str:
        return f"FedProx(mu={self.proximal_mu})"