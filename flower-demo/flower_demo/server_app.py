# server_app.py
"""flower-demo: Minimal ServerApp with per-round checkpoint saving."""

from __future__ import annotations

import os
from pathlib import Path

from flwr.common import Context, ndarrays_to_parameters, parameters_to_ndarrays
from flwr.server import ServerApp, ServerAppComponents, ServerConfig
from flwr.server.strategy import FedAvg

from flower_demo.task import Net, get_weights, set_weights
import mlflow


# --- 집계 메트릭 함수(가중 평균) ---
def _metrics_agg_fit(metrics):
    # [(num_examples, {"train_loss": ...}), ...]
    tot = sum(n for n, _ in metrics) or 1
    avg = sum(n * m.get("train_loss", 0.0) for n, m in metrics) / tot
    return {"train_loss": float(avg)}

def _metrics_agg_eval(metrics):
    # metrics: List[Tuple[num_examples, {k: v, ...}]]
    tot = sum(n for n, _ in metrics) or 1
    keys = set().union(*(m.keys() for _, m in metrics))  # 모든 키 수집
    agg = {}
    for k in keys:
        # 없는 클라엔 0으로 취급(원하면 샘플 수 있는 클라만 평균하도록 바꿔도 됨)
        agg[k] = sum(n * float(m.get(k, 0.0)) for n, m in metrics) / tot
    return agg

# 라운드 종료 시마다 체크포인트 저장하는 커스텀 전략
class SaveFedAvg(FedAvg):
    def __init__(self, *, mlflow_conf: dict | None = None, **kwargs):
        super().__init__(**kwargs)
        self.mlflow_enabled = mlflow is not None
        self._mlflow_run = None

        if self.mlflow_enabled:
            conf = mlflow_conf or {}
            uri = conf.get("tracking_uri", os.environ.get("MLFLOW_TRACKING_URI", "file:./mlruns"))
            exp = conf.get("experiment_name", os.environ.get("MLFLOW_EXPERIMENT_NAME", "flower-demo"))
            run_name = conf.get("run_name", os.environ.get("MLFLOW_RUN_NAME", "server-run"))
            mlflow.set_tracking_uri(uri)
            mlflow.set_experiment(exp)
            self._mlflow_run = mlflow.start_run(run_name=run_name)

    def _ml_log(self, metrics: dict, step: int):
        if self.mlflow_enabled and self._mlflow_run and metrics:
            mlflow.log_metrics({k: float(v) for k, v in metrics.items()}, step=step)
    
    def aggregate_fit(self, server_round, results, failures):
        # 표준 FedAvg 집계
        aggregated_params, aggregated_metrics = super().aggregate_fit(server_round, results, failures)

        # 체크포인트 저장
        if aggregated_params is not None:
            ndarrays = parameters_to_ndarrays(aggregated_params)
            net = Net()
            set_weights(net, ndarrays)

            ckpt_dir = Path("./checkpoints")
            ckpt_dir.mkdir(parents=True, exist_ok=True)
            ckpt_path = ckpt_dir / f"round-{server_round:03d}.pt"

            try:
                import torch
                torch.save(net.state_dict(), ckpt_path.as_posix())
                print(f"[Server] Saved checkpoint: {ckpt_path}")
                if self.mlflow_enabled:
                    mlflow.log_artifact(ckpt_path.as_posix(), artifact_path="checkpoints")
            except Exception as e:
                print(f"[Server] Failed to save checkpoint for round {server_round}: {e}")

        # fit 메트릭 로깅 (train_loss 등)
        if aggregated_metrics:
            self._ml_log(aggregated_metrics, step=server_round)

        return aggregated_params, aggregated_metrics
    
    def aggregate_evaluate(self, server_round, results, failures):
        # 표준 FedAvg 평가 집계(평균 loss 반환)
        aggregated_loss, aggregated_metrics = super().aggregate_evaluate(server_round, results, failures)

        # eval 메트릭 로깅 (val_loss + accuracy)
        to_log = {}
        if aggregated_loss is not None:
            to_log["val_loss"] = float(aggregated_loss)
        if aggregated_metrics:
            to_log.update(aggregated_metrics)
        self._ml_log(to_log, step=server_round)

        return aggregated_loss, aggregated_metrics

    def __del__(self):
        if self.mlflow_enabled and self._mlflow_run:
            try:
                mlflow.end_run()
            except Exception:
                pass       
    


def server_fn(context: Context) -> ServerAppComponents:
    # 최소 설정만 run_config에서 사용
    num_rounds = int(context.run_config.get("num-server-rounds", 3))
    fraction_fit = float(context.run_config.get("fraction-fit", 1.0))

    # 초기 파라미터
    initial_parameters = ndarrays_to_parameters(get_weights(Net()))

    # 전략: 집계 함수/MLflow 포함
    strategy = SaveFedAvg(
        fraction_fit=fraction_fit,
        fraction_evaluate=1.0,                 # 필요 시 0.3~0.5로 낮추면 메모리 절약
        min_available_clients=2,
        initial_parameters=initial_parameters,
        fit_metrics_aggregation_fn=_metrics_agg_fit,
        evaluate_metrics_aggregation_fn=_metrics_agg_eval,
        mlflow_conf={
            "tracking_uri": os.environ.get("MLFLOW_TRACKING_URI", "http://localhost:5000"),
            "experiment_name": os.environ.get("MLFLOW_EXPERIMENT_NAME", "flower-demo"),
            "run_name": os.environ.get("MLFLOW_RUN_NAME", "server-run"),
        },
    )

    # 하이퍼파라미터 기록(한 번만)
    if mlflow is not None and strategy._mlflow_run:
        params = {"num_server_rounds": num_rounds, "fraction_fit": fraction_fit}
        le = context.run_config.get("local-epochs")
        if le is not None:
            params["local_epochs"] = le
        mlflow.log_params(params)

    config = ServerConfig(num_rounds=num_rounds)
    return ServerAppComponents(strategy=strategy, config=config)


# Create ServerApp
app = ServerApp(server_fn=server_fn)
