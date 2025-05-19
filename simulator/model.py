import torch.nn as nn
from torchvision import models

class ResNet152Model(nn.Module):
    """ResNet152"""
    def __init__(self, num_classes=10):
        super(ResNet152Model, self).__init__()
        # 사전 학습된 ResNet152 모델 로드
        self.model = models.resnet152(weights=None, num_classes=100)
        
        # 마지막 fully connected 레이어 변경
        in_features = self.model.fc.in_features
        self.model.fc = nn.Linear(in_features, num_classes)
    
    def forward(self, x):
        return self.model(x)