import flwr as fl
from typing import Dict, List, Tuple, Optional, Union
from flwr.common import (
    FitRes, 
    Parameters, 
    Scalar,
    parameters_to_ndarrays,
    ndarrays_to_parameters,
)
import numpy as np

class FedYogi(fl.server.strategy.FedAvg):
    """FedYogi 전략 구현 (서버 측 Yogi 최적화 사용)"""
    
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
        eta: float = 0.1,  # 서버 학습률
        eta_l: float = 0.1,  # 클라이언트 학습률
        beta_1: float = 0.9,  # 1차 모멘트 감쇠율
        beta_2: float = 0.99,  # 2차 모멘트 감쇠율
        tau: float = 1e-3,  # Yogi 안정화 항
    ):
        self.eta = eta
        self.eta_l = eta_l
        self.beta_1 = beta_1
        self.beta_2 = beta_2
        self.tau = tau
        self.m_t = None  # 1차 모멘트 (속도)
        self.v_t = None  # 2차 모멘트 (가속도)
        self.round = 0   # 현재 라운드
        
        # FedYogi는 fit_config에 eta_l 값을 포함하도록 on_fit_config_fn을 재정의
        if on_fit_config_fn is None:
            on_fit_config_fn = self.default_on_fit_config
        else:
            original_on_fit_config_fn = on_fit_config_fn
            on_fit_config_fn = lambda round_num: {
                **original_on_fit_config_fn(round_num),
                "eta_l": self.eta_l,
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
            "eta_l": self.eta_l,
        }
    
    def aggregate_fit(
        self,
        server_round: int,
        results: List[Tuple[fl.server.client_proxy.ClientProxy, FitRes]],
        failures: List[Union[Tuple[fl.server.client_proxy.ClientProxy, FitRes], BaseException]],
    ) -> Tuple[Optional[Parameters], Dict[str, Scalar]]:
        """FedYogi 방식으로 클라이언트 모델 파라미터 집계"""
        if not results:
            return None, {}
        
        # 일반적인 FedAvg 방식 집계
        aggregated_parameters, metrics = super().aggregate_fit(server_round, results, failures)
        
        if aggregated_parameters is None:
            return None, {}
        
        # 집계된 파라미터 변환
        aggregated_ndarrays = parameters_to_ndarrays(aggregated_parameters)
        
        # 초기 파라미터가 있으면 변화량 계산
        if hasattr(self, "global_model_parameters") and self.global_model_parameters is not None:
            previous_ndarrays = parameters_to_ndarrays(self.global_model_parameters)
            
            # FedYogi 업데이트
            delta_t = [
                new - old for new, old in zip(aggregated_ndarrays, previous_ndarrays)
            ]
            
            # 처음 m_t, v_t 초기화
            if self.m_t is None or self.v_t is None:
                self.m_t = [np.zeros_like(x) for x in delta_t]
                self.v_t = [np.zeros_like(x) for x in delta_t]
            
            # 라운드 증가
            self.round += 1
            
            # m_t 업데이트 (1차 모멘트 - Adam과 동일)
            self.m_t = [
                self.beta_1 * m + (1 - self.beta_1) * delta
                for m, delta in zip(self.m_t, delta_t)
            ]
            
            # v_t 업데이트 (2차 모멘트 - Yogi 방식)
            self.v_t = [
                v - (1 - self.beta_2) * (delta ** 2) * np.sign(v - delta ** 2)
                for v, delta in zip(self.v_t, delta_t)
            ]
            
            # 바이어스 보정
            m_hat = [m / (1 - self.beta_1 ** self.round) for m in self.m_t]
            v_hat = [v / (1 - self.beta_2 ** self.round) for v in self.v_t]
            
            # 적응적 학습률 적용
            adapted_parameters = [
                old + self.eta * m / (np.sqrt(v) + self.tau)
                for old, m, v in zip(previous_ndarrays, m_hat, v_hat)
            ]
            
            # 파라미터 갱신
            aggregated_parameters = ndarrays_to_parameters(adapted_parameters)
        
        # 현재 파라미터 저장
        self.global_model_parameters = aggregated_parameters
        
        return aggregated_parameters, metrics
    
    def __repr__(self) -> str:
        return f"FedYogi(eta={self.eta}, eta_l={self.eta_l}, beta_1={self.beta_1}, beta_2={self.beta_2}, tau={self.tau})"