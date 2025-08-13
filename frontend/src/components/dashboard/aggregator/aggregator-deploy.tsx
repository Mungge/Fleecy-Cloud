// @/components/dashboard/aggregator/aggregator-deploy.tsx
"use client";

import { useState, useEffect } from "react";
import { useAggregatorCreation } from "./hooks/useAggregatorCreation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  CheckCircle,
  XCircle,
  Loader2,
  Cloud,
  Server,
  HardDrive,
  Zap,
} from "lucide-react";

export default function AggregatorDeploy() {
  const { isCreating, creationStatus, handleCreateAggregator, resetCreation } =
    useAggregatorCreation();

  // 팝업에서 선택된 옵션 (실제로는 props로 받아올 것)
  const [selectedOption] = useState({
    rank: 1,
    region: "af-south-1",
    instanceType: "m6gd.medium",
    cloudProvider: "AWS",
    estimatedMonthlyCost: 56.16,
    estimatedHourlyPrice: 0.06,
    avgLatency: 5.0,
    maxLatency: 5.0,
    vcpu: 1,
    memory: 4096, // 4GB
    recommendationScore: 4.3,
  });

  const [federatedLearningData] = useState({
    name: "Neural Network Federated Learning",
    description: "분산 신경망 모델을 이용한 연합 학습 프로젝트",
    modelType: "neural_network",
    algorithm: "FedAvg",
    rounds: 10,
    participants: [
      {
        id: "participant_1",
        name: "Medical Center A",
        status: "active",
        openstackEndpoint: "https://openstack-a.example.com",
      },
      {
        id: "participant_2",
        name: "Medical Center B",
        status: "active",
        openstackEndpoint: "https://openstack-b.example.com",
      },
      {
        id: "participant_3",
        name: "Medical Center C",
        status: "pending",
      },
    ],
    modelFileName: "neural_network_model.h5",
  });

  // 컴포넌트 마운트 시 자동으로 배포 시작
  useEffect(() => {
    const timer = setTimeout(() => {
      handleCreateAggregator(
        selectedOption,
        federatedLearningData,
        () => {
          console.log("배포 성공!");
        },
        (error) => {
          console.error("배포 실패:", error);
        }
      );
    }, 1000); // 1초 후 자동 시작

    return () => clearTimeout(timer);
  }, []);

  const getStatusIcon = () => {
    if (!creationStatus)
      return <Loader2 className="h-5 w-5 animate-spin text-blue-500" />;

    switch (creationStatus.step) {
      case "deploying":
        return <Loader2 className="h-5 w-5 animate-spin text-blue-500" />;
      case "completed":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "error":
        return <XCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Loader2 className="h-5 w-5 animate-spin text-blue-500" />;
    }
  };

  const getMainStatusText = () => {
    if (!creationStatus) return "배포 준비 중...";

    switch (creationStatus.step) {
      case "deploying":
        return "선택한 스펙으로 배포 중...";
      case "completed":
        return "배포 완료!";
      case "error":
        return "배포 실패";
      default:
        return "배포 준비 중...";
    }
  };

  const getMainStatusColor = () => {
    if (!creationStatus) return "text-blue-600";

    switch (creationStatus.step) {
      case "deploying":
        return "text-blue-600";
      case "completed":
        return "text-green-600";
      case "error":
        return "text-red-600";
      default:
        return "text-blue-600";
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 메인 상태 헤더 */}
      <div className="text-center space-y-4">
        <div className="flex items-center justify-center gap-3">
          {getStatusIcon()}
          <h1 className={`text-3xl font-bold ${getMainStatusColor()}`}>
            {getMainStatusText()}
          </h1>
        </div>
        <p className="text-muted-foreground">
          선택하신 AWS 인스턴스로 집계자를 배포하고 있습니다
        </p>
      </div>

      {/* 선택된 스펙 카드 */}
      <Card className="border-2 border-blue-200 bg-blue-50/30">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-blue-500" />
            선택된 배포 스펙
          </CardTitle>
          <CardDescription>
            팝업에서 선택하신 최적화된 인스턴스 구성
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">순위</span>
              <Badge variant="secondary">#{selectedOption.rank}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">클라우드 제공자</span>
              <Badge className="bg-orange-500">
                {selectedOption.cloudProvider}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">리전</span>
              <Badge variant="outline">{selectedOption.region}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">인스턴스 타입</span>
              <div className="flex items-center gap-1">
                <Server className="h-3 w-3" />
                <span className="text-sm font-mono">
                  {selectedOption.instanceType}
                </span>
              </div>
            </div>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">vCPU</span>
              <span className="text-sm">{selectedOption.vcpu} 코어</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">메모리</span>
              <div className="flex items-center gap-1">
                <HardDrive className="h-3 w-3" />
                <span className="text-sm">
                  {(selectedOption.memory / 1024).toFixed(1)}GB
                </span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">지연시간</span>
              <span className="text-sm text-green-600">
                {selectedOption.avgLatency}ms
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">월 예상 비용</span>
              <span className="text-sm font-semibold">
                ₩{selectedOption.estimatedMonthlyCost.toLocaleString()}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 배포 진행 상황 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {getStatusIcon()}
            배포 진행 상황
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {creationStatus && (
            <>
              <Alert
                className={`${
                  creationStatus.step === "error"
                    ? "border-red-200 bg-red-50"
                    : creationStatus.step === "completed"
                    ? "border-green-200 bg-green-50"
                    : "border-blue-200 bg-blue-50"
                }`}
              >
                <AlertDescription className="flex items-center gap-2">
                  {getStatusIcon()}
                  {creationStatus.message}
                </AlertDescription>
              </Alert>

              {creationStatus.progress !== undefined && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>배포 진행률</span>
                    <span>{creationStatus.progress}%</span>
                  </div>
                  <Progress
                    value={creationStatus.progress}
                    className="w-full"
                  />
                </div>
              )}
            </>
          )}

          {!creationStatus && (
            <div className="flex items-center gap-2 text-blue-600">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>배포 시작 준비 중...</span>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 액션 버튼 */}
      <div className="flex justify-center gap-4">
        {creationStatus?.step === "completed" && (
          <Button size="lg" className="min-w-[150px]">
            대시보드로 이동
          </Button>
        )}

        {creationStatus?.step === "error" && (
          <Button
            variant="outline"
            onClick={() => window.location.reload()}
            size="lg"
          >
            다시 시도
          </Button>
        )}

        {(creationStatus?.step === "completed" ||
          creationStatus?.step === "error") && (
          <Button variant="outline" onClick={resetCreation} size="lg">
            새로 시작
          </Button>
        )}
      </div>
    </div>
  );
}
