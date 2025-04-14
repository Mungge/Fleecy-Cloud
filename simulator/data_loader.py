import torch
import torchvision
import torchvision.transforms as transforms
from torch.utils.data import DataLoader, Subset
import numpy as np
import os

def load_cifar10_partition(partition_id=0, num_partitions=5, batch_size=32):
    """CIFAR-10 데이터셋을 로드하고 파티션으로 분할"""
    transform = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))
    ])
    
    # 데이터 디렉토리 확인
    data_dir = './data'
    os.makedirs(data_dir, exist_ok=True)
    
    # 이미 다운로드된 데이터셋 사용 (download=False)
    print("이미 다운로드된 CIFAR-10 데이터셋 사용...")
    
    try:
        trainset = torchvision.datasets.CIFAR10(
            root=data_dir, train=True, download=False, transform=transform
        )
        
        testset = torchvision.datasets.CIFAR10(
            root=data_dir, train=False, download=False, transform=transform
        )
    except Exception as e:
        print(f"다운로드된 데이터셋 로드 실패: {e}")
        print("데이터셋이 없는 것 같습니다. 다운로드를 시도합니다.")
        
        # 데이터셋이 없는 경우 다운로드 시도
        try:
            trainset = torchvision.datasets.CIFAR10(
                root=data_dir, train=True, download=True, transform=transform
            )
            
            testset = torchvision.datasets.CIFAR10(
                root=data_dir, train=False, download=True, transform=transform
            )
        except Exception as e2:
            print(f"데이터셋 다운로드 및 로드 실패: {e2}")
            raise RuntimeError("CIFAR-10 데이터셋을 로드할 수 없습니다.")
    
    # 비IID 데이터 분할을 위한 설정
    num_classes = 10
    
    # 각 클래스별 인덱스 그룹화
    class_indices = [[] for _ in range(num_classes)]
    for idx, (_, label) in enumerate(trainset):
        class_indices[label].append(idx)
    
    # 파티션별 클래스 할당 (비IID 설정)
    classes_per_partition = 2  # 파티션별 클래스 수
    partition_classes = []
    
    for i in range(num_partitions):
        # 파티션마다 다른 클래스 조합 할당
        classes = [(i * classes_per_partition + j) % num_classes for j in range(classes_per_partition)]
        partition_classes.append(classes)
    
    # 현재 파티션에 해당하는 데이터 선택
    current_partition_indices = []
    for class_idx in partition_classes[partition_id]:
        # 해당 클래스의 인덱스를 현재 파티션에 추가
        current_partition_indices.extend(class_indices[class_idx])
    
    # 파티션의 학습 데이터셋 생성
    partition_trainset = Subset(trainset, current_partition_indices)
    
    # 데이터로더 생성 - num_workers=0으로 설정하여 세그멘테이션 폴트 방지
    trainloader = DataLoader(
        partition_trainset, batch_size=batch_size, shuffle=True, 
        num_workers=0  # 멀티프로세싱 비활성화
    )
    
    testloader = DataLoader(
        testset, batch_size=batch_size, shuffle=False, 
        num_workers=0  # 멀티프로세싱 비활성화
    )
    
    print(f"Partition {partition_id} has {len(partition_trainset)} training samples "
          f"from classes {partition_classes[partition_id]}")
    
    return trainloader, testloader

# 대체 데이터셋 로드 함수 (CIFAR-10 로드에 실패했을 경우 사용)
def load_mnist_partition(partition_id=0, num_partitions=5, batch_size=32):
    """MNIST 데이터셋을 로드하고, 파티션으로 분할 (CIFAR-10 대체용)"""
    transform = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.1307,), (0.3081,))
    ])
    
    # 데이터 디렉토리 확인
    data_dir = './data'
    os.makedirs(data_dir, exist_ok=True)
    
    # 이미 다운로드된 데이터셋 사용 시도
    try:
        trainset = torchvision.datasets.MNIST(
            root=data_dir, train=True, download=False, transform=transform
        )
        
        testset = torchvision.datasets.MNIST(
            root=data_dir, train=False, download=False, transform=transform
        )
    except Exception:
        print("MNIST 데이터셋을 다운로드합니다...")
        trainset = torchvision.datasets.MNIST(
            root=data_dir, train=True, download=True, transform=transform
        )
        
        testset = torchvision.datasets.MNIST(
            root=data_dir, train=False, download=True, transform=transform
        )
    
    # 비IID 데이터 분할
    num_classes = 10
    
    # 각 클래스별 인덱스 그룹화
    class_indices = [[] for _ in range(num_classes)]
    for idx, (_, label) in enumerate(trainset):
        class_indices[label].append(idx)
    
    # 파티션별 클래스 할당 (비IID 설정)
    classes_per_partition = 2  # 파티션별 클래스 수
    partition_classes = []
    
    for i in range(num_partitions):
        # 파티션마다 다른 클래스 조합 할당
        classes = [(i * classes_per_partition + j) % num_classes for j in range(classes_per_partition)]
        partition_classes.append(classes)
    
    # 현재 파티션에 해당하는 데이터 선택
    current_partition_indices = []
    for class_idx in partition_classes[partition_id]:
        # 해당 클래스의 인덱스를 현재 파티션에 추가
        current_partition_indices.extend(class_indices[class_idx])
    
    # 파티션의 학습 데이터셋 생성
    partition_trainset = Subset(trainset, current_partition_indices)
    
    # 데이터로더 생성 - num_workers=0으로 설정하여 세그멘테이션 폴트 방지
    trainloader = DataLoader(
        partition_trainset, batch_size=batch_size, shuffle=True, num_workers=0
    )
    
    testloader = DataLoader(
        testset, batch_size=batch_size, shuffle=False, num_workers=0
    )
    
    print(f"Partition {partition_id} has {len(partition_trainset)} training samples "
          f"from classes {partition_classes[partition_id]} (MNIST)")
    
    return trainloader, testloader