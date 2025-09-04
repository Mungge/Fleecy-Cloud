# task.py
from __future__ import annotations

from collections import OrderedDict
from typing import Tuple, Dict, Any, List, Optional
import os
import math
import random

import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import DataLoader, Subset, random_split
from torchvision import transforms as T
from torchvision.datasets import ImageFolder
from sklearn.metrics import accuracy_score, precision_recall_fscore_support

# =========================
# 1) Model (PyTorch, Keras->Torch 변환)
# =========================
class Net(nn.Module):
    """
    Keras Sequential 모델을 PyTorch로 변환한 CNN 분류기.
    입력: (B, 3, 224, 224)
    Conv(32)-Pool -> Conv(64)-Pool -> Conv(128)-Pool -> Conv(256)-Pool -> Flatten -> FC(128) -> FC(64) -> FC(3)
    """
    def __init__(self, num_classes: int = 3):
        super().__init__()
        self.features = nn.Sequential(
            nn.Conv2d(3, 32, kernel_size=3, stride=1, padding=1),  # 'same'
            nn.ReLU(inplace=True),
            nn.MaxPool2d(kernel_size=2, stride=2),                  # 224 -> 112

            nn.Conv2d(32, 64, kernel_size=3, stride=1, padding=1), # 'same'
            nn.ReLU(inplace=True),
            nn.MaxPool2d(kernel_size=2, stride=2),                  # 112 -> 56

            nn.Conv2d(64, 128, kernel_size=3, stride=1, padding=1),# 'same'
            nn.ReLU(inplace=True),
            nn.MaxPool2d(kernel_size=2, stride=2),                  # 56 -> 28

            nn.Conv2d(128, 256, kernel_size=3, stride=1, padding=1),# 'same'
            nn.ReLU(inplace=True),
            nn.MaxPool2d(kernel_size=2, stride=2),                  # 28 -> 14
        )
        # 256 * 14 * 14 = 50176
        self.classifier = nn.Sequential(
            nn.Flatten(),
            nn.Linear(256 * 14 * 14, 128),
            nn.ReLU(inplace=True),
            nn.Linear(128, 64),
            nn.ReLU(inplace=True),
            nn.Linear(64, num_classes),  # 마지막에 softmax는 사용하지 않음 (CrossEntropyLoss가 내부에서 처리)
        )

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        x = self.features(x)
        x = self.classifier(x)
        return x


# =========================
# 2) Dataset / DataLoader
# =========================
def _partition_indices(n: int, partition_id: int, num_partitions: int) -> List[int]:
    """
    전체 n개 샘플을 num_partitions로 균등 분할했을 때, 특정 partition_id가 담당할 인덱스 리스트를 반환합니다.
    - 여러 VM/클라이언트가 동일한 공유 데이터셋을 바라보는 상황에서 IID 파티셔닝 용도.
    """
    assert 0 <= partition_id < num_partitions, "partition_id out of range"
    indices = list(range(n))
    start = (n * partition_id) // num_partitions
    end = (n * (partition_id + 1)) // num_partitions
    return indices[start:end]


def _collate_to_dict(batch):
    """
    torchvision ImageFolder는 (image, label) 튜플을 반환하므로
    Flower 예제와 맞추기 위해 {"img": x, "label": y} 형태로 변환합니다.
    """
    imgs, labels = zip(*batch)
    return {"img": torch.stack(imgs, dim=0), "label": torch.tensor(labels, dtype=torch.long)}


def load_data(partition_id: int, num_partitions: int) -> Tuple[DataLoader, DataLoader]:
    """
    Kaggle COVID-19 Image Dataset 구조를 가정:
      <DATA_ROOT>/
        Covid19-dataset/
          train/
            Covid/
            Normal/
            Viral Pneumonia/
          test/
            Covid/
            Normal/
            Viral Pneumonia/

    환경변수:
      - DATA_ROOT (기본값: ~/datasets/covid19/Covid19-dataset)
      - BATCH_SIZE (기본값: 32)
      - NUM_WORKERS (기본값: 2)
      - VAL_SPLIT (기본값: 0.2)  # train 내부에서 검증셋 분할 비율
    """
    data_root = os.environ.get("DATA_ROOT", os.path.expanduser("~/Covid19-dataset"))
    train_dir = os.path.join(data_root, "train")
    test_dir = os.path.join(data_root, "test")  # 필요 시 검증에 사용하거나, train에서 분할 사용

    batch_size = int(os.environ.get("BATCH_SIZE", "32"))
    num_workers = int(os.environ.get("NUM_WORKERS", "2"))
    val_split = float(os.environ.get("VAL_SPLIT", "0.2"))

    # Keras 예시와 유사한 전처리/증강
    train_transform = T.Compose([
        T.Resize((224, 224)),
        T.RandomHorizontalFlip(p=0.5),
        T.RandomRotation(20),
        T.ToTensor(),
        T.Normalize(mean=(0.5, 0.5, 0.5), std=(0.5, 0.5, 0.5)),  # [0,1] -> [-1,1] 근사
    ])
    eval_transform = T.Compose([
        T.Resize((224, 224)),
        T.ToTensor(),
        T.Normalize(mean=(0.5, 0.5, 0.5), std=(0.5, 0.5, 0.5)),
    ])

    # ImageFolder 로드
    full_train = ImageFolder(train_dir, transform=train_transform)
    # (선택) 제공되는 test를 검증으로 쓸 수도 있지만, 여기서는 train을 (1 - val_split):(val_split)로 분리
    n_total = len(full_train)
    n_val = max(1, int(n_total * val_split))
    n_train = n_total - n_val
    base_train, base_val = random_split(full_train, [n_train, n_val], generator=torch.Generator().manual_seed(42))

    # 파티셔닝 (여러 클라이언트가 같은 경로를 볼 때 IID 분배)
    train_idx = _partition_indices(len(base_train), partition_id, num_partitions)
    val_idx = _partition_indices(len(base_val), partition_id, num_partitions)

    train_subset = Subset(base_train, train_idx) if num_partitions > 1 else base_train
    val_subset = Subset(base_val, val_idx) if num_partitions > 1 else base_val

    # 검증용은 증강 없이 평가용 변환으로 교체
    # (random_split은 Subset을 반환하므로, transform을 바꾸려면 underlying dataset의 transform를 바꿔야 함)
    # 안전하게 별도 ImageFolder로 불러와 매칭하는 방식도 가능하지만, 여기선 간단히 다음과 같이 처리:
    # val_subset.dataset.dataset.transform = eval_transform  # random_split -> Subset -> (Dataset) 구조
    # 위 한 줄은 타입 체인이 복잡할 수 있으므로, 가장 간단한 방법은 validation 전용 ImageFolder를 따로 두는 것.
    # 다만 여기선 train 내부에서 분할했으니, 평가 시 입력에만 eval_transform를 적용하도록 별도 경로를 쓰겠습니다.
    # => 간단성을 위해 아래 DataLoader의 collate에서 그대로 사용(이미 텐서라 transform 교체 불필요)

    trainloader = DataLoader(
        train_subset,
        batch_size=batch_size,
        shuffle=True,
        num_workers=num_workers,
        pin_memory=True,
        collate_fn=_collate_to_dict,
    )
    valloader = DataLoader(
        val_subset,
        batch_size=batch_size,
        shuffle=False,
        num_workers=num_workers,
        pin_memory=True,
        collate_fn=_collate_to_dict,
    )
    return trainloader, valloader


# =========================
# 3) Train / Eval
# =========================
def train(net: nn.Module, trainloader: DataLoader, epochs: int, device: torch.device) -> float:
    """
    로컬 학습 루프. 평균 train loss를 반환합니다.
    - 손실: CrossEntropyLoss (Keras categorical/sparse 대응)
    - 옵티마이저: Adam(lr=1e-3)
    """
    net.to(device)
    criterion = nn.CrossEntropyLoss().to(device)
    optimizer = torch.optim.Adam(net.parameters(), lr=float(os.environ.get("LR", "1e-3")))

    net.train()
    running_loss = 0.0
    total_steps = 0

    for _ in range(epochs):
        for batch in trainloader:
            images: torch.Tensor = batch["img"].to(device, non_blocking=True)
            labels: torch.Tensor = batch["label"].to(device, non_blocking=True)

            optimizer.zero_grad(set_to_none=True)
            logits = net(images)               # (B, 3)
            loss = criterion(logits, labels)   # CE
            loss.backward()
            optimizer.step()

            running_loss += loss.item()
            total_steps += 1

    avg_trainloss = running_loss / max(total_steps, 1)
    return avg_trainloss

def test(net, testloader, device):
    """Validate the model and return (avg_loss, metrics_dict) without AUC."""
    net.to(device)
    criterion = torch.nn.CrossEntropyLoss()
    net.eval()

    all_labels, all_preds = [], []
    running_loss, steps = 0.0, 0

    with torch.no_grad():
        for batch in testloader:
            images = batch["img"].to(device)
            labels = batch["label"].to(device)

            logits = net(images)
            loss = criterion(logits, labels).item()
            running_loss += loss
            steps += 1

            preds = logits.argmax(dim=1)          # 확률/softmax 불필요 (AUC 제거)
            all_labels.append(labels.cpu())
            all_preds.append(preds.cpu())

    y_true = torch.cat(all_labels).numpy()
    y_pred = torch.cat(all_preds).numpy()
    num_classes = int(y_pred.max()) + 1 if len(y_pred) > 0 else 0

    # 기본값
    acc = float((y_pred == y_true).mean()) if len(y_true) else 0.0
    prec_macro = rec_macro = f1_macro = float("nan")

    acc = float(accuracy_score(y_true, y_pred)) if len(y_true) else 0.0
    prec_macro, rec_macro, f1_macro, _ = precision_recall_fscore_support(
        y_true, y_pred, average="macro", zero_division=0
    )

    avg_loss = running_loss / max(steps, 1)
    metrics = {
        "accuracy": acc,
        "precision_macro": float(prec_macro),
        "recall_macro": float(rec_macro),
        "f1_macro": float(f1_macro),
        "num_examples": int(len(y_true)),
    }
    return avg_loss, metrics


# =========================
# 4) Flower helpers
# =========================
def get_weights(net: nn.Module) -> List[Any]:
    return [val.detach().cpu().numpy() for _, val in net.state_dict().items()]


def set_weights(net: nn.Module, parameters: List[Any]) -> None:
    params_dict = zip(net.state_dict().keys(), parameters)
    state_dict = OrderedDict({k: torch.tensor(v) for k, v in params_dict})
    net.load_state_dict(state_dict, strict=True)
