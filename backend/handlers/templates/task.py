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
def _collate_to_dict(batch):
    """
    torchvision ImageFolder는 (image, label) 튜플을 반환하므로
    Flower 예제와 맞추기 위해 {"img": x, "label": y} 형태로 변환합니다.
    """
    imgs, labels = zip(*batch)
    return {"img": torch.stack(imgs, dim=0), "label": torch.tensor(labels, dtype=torch.long)}


def load_data(partition_id: int = 0, num_partitions: int = 1) -> Tuple[DataLoader, DataLoader]:
    """
    Remote Federation 전용: 각 VM이 ~/Covid19-dataset에 있는 고유한 로컬 데이터셋 사용
    
    데이터 구조:
      ~/Covid19-dataset/
        train/
          Covid/
          Normal/
          Viral Pneumonia/
    
    환경 변수:
      - BATCH_SIZE (기본값: 32)
      - NUM_WORKERS (기본값: 2)  
      - VAL_SPLIT (기본값: 0.2)
    """
    data_path = os.path.expanduser("~/Covid19-dataset/train")
    batch_size = int(os.environ.get("BATCH_SIZE", "32"))
    num_workers = int(os.environ.get("NUM_WORKERS", "2"))
    val_split = float(os.environ.get("VAL_SPLIT", "0.2"))
    
    if not os.path.exists(data_path):
        raise FileNotFoundError(f"Dataset not found: {data_path}")
    
    print(f"[Remote Federation] Using local dataset: {data_path}")
    
    # Transform 정의
    train_transform = T.Compose([
        T.Resize((224, 224)),
        T.RandomHorizontalFlip(p=0.5),
        T.RandomRotation(20),
        T.ToTensor(),
        T.Normalize(mean=(0.5, 0.5, 0.5), std=(0.5, 0.5, 0.5)),
    ])
    
    # 각 클라이언트의 전체 데이터셋 로드
    full_dataset = ImageFolder(data_path, transform=train_transform)
    
    # 클라이언트별 데이터를 train/val로 분할 (파티셔닝 없음)
    n_total = len(full_dataset)
    n_val = max(1, int(n_total * val_split))
    n_train = n_total - n_val
    train_dataset, val_dataset = random_split(full_dataset, [n_train, n_val], 
                                            generator=torch.Generator().manual_seed(42))

    # DataLoader 생성
    trainloader = DataLoader(
        train_dataset,
        batch_size=batch_size,
        shuffle=True,
        num_workers=num_workers,
        pin_memory=True,
        collate_fn=_collate_to_dict,
    )
    valloader = DataLoader(
        val_dataset,
        batch_size=batch_size,
        shuffle=False,
        num_workers=num_workers,
        pin_memory=True,
        collate_fn=_collate_to_dict,
    )
    
    print(f"Loaded {len(train_dataset)} train samples, {len(val_dataset)} val samples")
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
