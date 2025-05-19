import os
import time
import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
import flwr as fl
from collections import OrderedDict
from pathlib import Path

from model import ResNet152Model
from data_loader import load_cifar10_partition

# 환경 변수 가져오기
CLIENT_ID = int(os.environ.get("CLIENT_ID", "0"))
SERVER_ADDRESS = os.environ.get("SERVER_ADDRESS", "localhost:8080")
DATASET_PARTITION = int(os.environ.get("DATASET_PARTITION", "0"))
RESULTS_DIR = os.environ.get("RESULTS_DIR", "./results")

# 결과 디렉토리 생성
Path(RESULTS_DIR).mkdir(parents=True, exist_ok=True)

# 학습 설정
DEVICE = torch.device("cuda:0" if torch.cuda.is_available() else "cpu")
print(f"Using device: {DEVICE}")

BATCH_SIZE = 32
EPOCHS = 1  # 클라이언트당 1 에폭만 학습

def train(net, trainloader, epochs):
    """모델 학습"""
    criterion = nn.CrossEntropyLoss()
    optimizer = torch.optim.SGD(net.parameters(), lr=0.001, momentum=0.9)
    
    net.train()
    for epoch in range(epochs):
        for batch_idx, (images, labels) in enumerate(trainloader):
            images, labels = images.to(DEVICE), labels.to(DEVICE)
            optimizer.zero_grad()
            outputs = net(images)
            loss = criterion(outputs, labels)
            loss.backward()
            optimizer.step()
            
            if batch_idx % 100 == 0:
                print(f"Client {CLIENT_ID}: Train Epoch: {epoch} [{batch_idx*len(images)}/{len(trainloader.dataset)} "
                      f"({100. * batch_idx / len(trainloader):.0f}%)]\tLoss: {loss.item():.6f}")

def test(net, testloader):
    """모델 평가"""
    criterion = nn.CrossEntropyLoss()
    correct = 0
    total = 0
    loss = 0.0
    
    net.eval()
    with torch.no_grad():
        for images, labels in testloader:
            images, labels = images.to(DEVICE), labels.to(DEVICE)
            outputs = net(images)
            loss += criterion(outputs, labels).item()
            _, predicted = torch.max(outputs.data, 1)
            total += labels.size(0)
            correct += (predicted == labels).sum().item()
            
    accuracy = correct / total
    test_loss = loss / len(testloader)
    print(f"Client {CLIENT_ID}: Test set: Average loss: {test_loss:.4f}, "
          f"Accuracy: {correct}/{total} ({accuracy:.4f})")
    
    return loss, accuracy

class FlowerClient(fl.client.NumPyClient):
    def __init__(self, net, trainloader, testloader):
        self.net = net
        self.trainloader = trainloader
        self.testloader = testloader
    
    def get_parameters(self, config):
        """모델 파라미터를 numpy 배열 형태로 반환"""
        return [val.cpu().numpy() for _, val in self.net.state_dict().items()]
    
    def set_parameters(self, parameters):
        """서버로부터 받은 파라미터를 모델에 설정"""
        params_dict = zip(self.net.state_dict().keys(), parameters)
        state_dict = OrderedDict({k: torch.tensor(v) for k, v in params_dict})
        self.net.load_state_dict(state_dict, strict=True)
    
    def fit(self, parameters, config):
        """로컬 데이터로 모델 훈련"""
        self.set_parameters(parameters)
        train(self.net, self.trainloader, epochs=EPOCHS)
        return self.get_parameters(config={}), len(self.trainloader.dataset), {}
    
    def evaluate(self, parameters, config):
        """모델 평가"""
        self.set_parameters(parameters)
        loss, accuracy = test(self.net, self.testloader)
        return float(loss), len(self.testloader.dataset), {"accuracy": float(accuracy)}

def main():
    """Flower 클라이언트 시작"""
    print(f"Starting client {CLIENT_ID} - connecting to server at {SERVER_ADDRESS}")
    
    
    # 데이터셋 로드
    trainloader, testloader = load_cifar10_partition(
        partition_id=DATASET_PARTITION,
        num_partitions=5,  # 전체 클라이언트 수보다 크게 설정하여 각 클라이언트가 다른 데이터를 가지도록 함
        batch_size=BATCH_SIZE
    )
    
    # 모델 초기화
    net = ResNet152Model.to(DEVICE)
    
    # Flower 클라이언트 생성 및 시작
    client = FlowerClient(net, trainloader, testloader)
    fl.client.start_client(
        server_address=SERVER_ADDRESS,
        client=client.to_client(),
        grpc_max_message_length=1024*1024*1024  # 1GB
    )

if __name__ == "__main__":
    main()