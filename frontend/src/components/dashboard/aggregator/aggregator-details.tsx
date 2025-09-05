"use client";

import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";

interface TrainingHistoryResponse {
  round: number;
  accuracy: number;
  loss: number;
  timestamp: string;
  participants: number;
}

interface AggregatorDetailsProps {
  aggregator: {
    id: string;
    name: string;
    status: "running" | "completed" | "error" | "pending" | "creating";
    algorithm: string;
    federatedLearningId?: string;
    federatedLearningName: string;
    cloudProvider: string;
    region: string;
    instanceType: string;
    createdAt: string;
    lastUpdated: string;
    participants: number;
    rounds: number;
    currentRound: number;
    accuracy: number;
    cost?: {
      current: number;
      estimated: number;
    };
    specs: {
      cpu: string;
      memory: string;
      storage: string;
    };
    metrics: {
      cpuUsage: number;
      memoryUsage: number;
      networkUsage: number;
    };
    mlflowExperimentName?: string;
    mlflowExperimentId?: string;
  };
  onBack: () => void;
}

const AggregatorDetails: React.FC<AggregatorDetailsProps> = ({
  aggregator,
  onBack,
}) => {
  const realTimeMetrics = {
    cpuUsage: aggregator.metrics.cpuUsage,
    memoryUsage: aggregator.metrics.memoryUsage,
    networkUsage: aggregator.metrics.networkUsage,
    accuracy: aggregator.accuracy,
    loss: 0.1234,
    participantsConnected: aggregator.participants,
    lastUpdated: new Date().toISOString(),
  };

  const trainingHistory: TrainingHistoryResponse[] = Array.from(
    { length: 8 },
    (_, i) => ({
      round: i + 1,
      accuracy: Math.min(0.6 + i * 0.05 + Math.random() * 0.1, 0.95),
      loss: Math.max(2.0 - i * 0.15 - Math.random() * 0.2, 0.1),
      timestamp: new Date(Date.now() - (8 - i) * 600000).toISOString(),
      participants: aggregator.participants,
    })
  );

  const isLoading = false;

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
      case "completed":
        return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300";
      case "error":
        return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
      case "creating":
        return "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300";
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "running":
        return "실행 중";
      case "completed":
        return "완료됨";
      case "error":
        return "오류";
      case "pending":
        return "대기 중";
      case "creating":
        return "생성 중";
      default:
        return "알 수 없음";
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString("ko-KR");
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR", {
      style: "currency",
      currency: "USD",
    }).format(amount);
  };

  const progressPercentage =
    (aggregator.currentRound / aggregator.rounds) * 100;

  const getMetricColor = (value: number) => {
    if (value >= 80) return "bg-red-500";
    if (value >= 60) return "bg-yellow-500";
    return "bg-green-500";
  };

  return (
    <div className="space-y-6">
      {/* 헤더 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="outline" onClick={onBack}>
            ← 뒤로가기
          </Button>
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-3xl font-bold">{aggregator.name}</h1>
              <Badge className={getStatusColor(aggregator.status)}>
                {getStatusText(aggregator.status)}
              </Badge>
            </div>
            <p className="text-muted-foreground mt-1">
              Aggregator 상세 정보 및 실시간 모니터링
            </p>
          </div>
        </div>
        <div className="flex space-x-2">
          <Button variant="outline">새로고침</Button>
          {aggregator.status === "running" && (
            <Button variant="destructive">중지</Button>
          )}
        </div>
      </div>

      {/* 기본 정보 카드 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>기본 정보</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">ID</p>
                <p className="font-mono text-sm">{aggregator.id}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  알고리즘
                </p>
                <p>{aggregator.algorithm}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  연합학습
                </p>
                <p>{aggregator.federatedLearningName}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  클라우드 제공자
                </p>
                <p>{aggregator.cloudProvider}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  리전
                </p>
                <p>{aggregator.region}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  인스턴스 타입
                </p>
                <p>{aggregator.instanceType}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  생성일
                </p>
                <p>{formatDate(aggregator.createdAt)}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  마지막 업데이트
                </p>
                <p>{formatDate(realTimeMetrics.lastUpdated)}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>하드웨어 사양</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">CPU</p>
                <p>{aggregator.specs.cpu}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  메모리
                </p>
                <p>{aggregator.specs.memory}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  스토리지
                </p>
                <p>{aggregator.specs.storage}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 학습 진행 상황 */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <Card className="xl:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              학습 진행 상황
              <Badge variant="outline" className="text-xs">
                라운드 {aggregator.currentRound}/{aggregator.rounds}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              <div className="relative">
                <Progress value={progressPercentage} className="h-3" />
                <div className="flex justify-between text-xs text-muted-foreground mt-1">
                  <span>시작</span>
                  <span className="font-medium">
                    {progressPercentage.toFixed(1)}% 완료
                  </span>
                  <span>완료</span>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="bg-muted p-4 rounded-lg">
                  <div className="text-2xl font-bold">
                    {realTimeMetrics.accuracy.toFixed(2)}%
                  </div>
                  <div className="text-sm text-muted-foreground font-medium">
                    현재 정확도
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    +2.3% vs 이전 라운드
                  </div>
                </div>
                <div className="bg-muted p-4 rounded-lg">
                  <div className="text-2xl font-bold">
                    {realTimeMetrics.loss.toFixed(4)}
                  </div>
                  <div className="text-sm text-muted-foreground font-medium">
                    현재 손실
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    -0.0123 vs 이전 라운드
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>실시간 상태</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="text-center py-2">
              <div className="text-3xl font-bold text-green-600">
                {realTimeMetrics.participantsConnected}
              </div>
              <div className="text-sm text-muted-foreground">연결된 참여자</div>
              <div className="flex items-center justify-center mt-2">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse mr-2"></div>
                <span className="text-xs text-green-600">활성 상태</span>
              </div>
            </div>

            <div className="space-y-3 pt-4 border-t">
              <div className="text-xs text-muted-foreground uppercase tracking-wide">
                최근 활동
              </div>
              <div className="space-y-2">
                <div className="text-sm">
                  <span className="font-medium">
                    라운드 {aggregator.currentRound}
                  </span>{" "}
                  진행 중
                </div>
                <div className="text-xs text-muted-foreground">
                  예상 완료: 약 15분 후
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 시스템 메트릭 */}
      <Card>
        <CardHeader>
          <CardTitle>시스템 메트릭</CardTitle>
          <CardDescription>실시간 시스템 리소스 사용량</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium flex items-center">
                  {/* FIX 5: getMetricColor 호출 시 불필요한 두 번째 인자 제거 */}
                  <div
                    className={`w-2 h-2 rounded-full mr-2 ${getMetricColor(
                      realTimeMetrics.cpuUsage
                    )}`}
                  ></div>
                  CPU 사용률
                </span>
                <span className="text-sm font-mono">
                  {realTimeMetrics.cpuUsage}%
                </span>
              </div>
              <Progress value={realTimeMetrics.cpuUsage} className="h-2" />
              <div className="text-xs text-muted-foreground">
                {realTimeMetrics.cpuUsage < 60
                  ? "정상"
                  : realTimeMetrics.cpuUsage < 80
                  ? "주의"
                  : "위험"}
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium flex items-center">
                  {/* FIX 5: getMetricColor 호출 시 불필요한 두 번째 인자 제거 */}
                  <div
                    className={`w-2 h-2 rounded-full mr-2 ${getMetricColor(
                      realTimeMetrics.memoryUsage
                    )}`}
                  ></div>
                  메모리 사용률
                </span>
                <span className="text-sm font-mono">
                  {realTimeMetrics.memoryUsage}%
                </span>
              </div>
              <Progress value={realTimeMetrics.memoryUsage} className="h-2" />
              <div className="text-xs text-muted-foreground">
                {realTimeMetrics.memoryUsage < 60
                  ? "정상"
                  : realTimeMetrics.memoryUsage < 80
                  ? "주의"
                  : "위험"}
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium flex items-center">
                  {/* FIX 5: getMetricColor 호출 시 불필요한 두 번째 인자 제거 */}
                  <div
                    className={`w-2 h-2 rounded-full mr-2 ${getMetricColor(
                      realTimeMetrics.networkUsage
                    )}`}
                  ></div>
                  네트워크 사용률
                </span>
                <span className="text-sm font-mono">
                  {realTimeMetrics.networkUsage}%
                </span>
              </div>
              <Progress value={realTimeMetrics.networkUsage} className="h-2" />
              <div className="text-xs text-muted-foreground">
                {realTimeMetrics.networkUsage < 60
                  ? "정상"
                  : realTimeMetrics.networkUsage < 80
                  ? "주의"
                  : "위험"}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 비용 정보 */}
      {aggregator.cost && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle>현재 사용 비용</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">
                {formatCurrency(aggregator.cost.current)}
              </div>
              <div className="text-sm text-muted-foreground mt-1">
                시간당 평균: {formatCurrency(aggregator.cost.current / 24)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle>예상 총 비용</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">
                {formatCurrency(aggregator.cost.estimated)}
              </div>
              <div className="text-sm text-muted-foreground mt-1">
                남은 예산:{" "}
                {formatCurrency(
                  aggregator.cost.estimated - aggregator.cost.current
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 학습 히스토리 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            학습 히스토리
            <Badge variant="outline" className="text-xs">
              최근 {Math.min(trainingHistory.length, 10)}개 라운드
            </Badge>
          </CardTitle>
          <CardDescription>라운드별 성능 변화 추이</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center items-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-primary"></div>
            </div>
          ) : trainingHistory.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>학습 히스토리가 없습니다.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {trainingHistory.slice(-6).map((history, index) => {
                const isLatest = index === trainingHistory.slice(-6).length - 1;
                const accuracyChange =
                  index > 0
                    ? (
                        (history.accuracy -
                          trainingHistory.slice(-6)[index - 1].accuracy) *
                        100
                      ).toFixed(2)
                    : "0.00";
                const isImproving = parseFloat(accuracyChange) > 0;

                return (
                  <div
                    key={index}
                    className={`flex items-center justify-between p-4 rounded-lg border-2 transition-all duration-200 ${
                      isLatest
                        ? "border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950"
                        : "border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950"
                    }`}
                  >
                    <div className="flex items-center space-x-4">
                      <div
                        className={`flex items-center justify-center w-10 h-10 rounded-full ${
                          isLatest
                            ? "bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-400"
                            : "bg-gray-100 text-gray-600 dark:bg-gray-900 dark:text-gray-400"
                        }`}
                      >
                        <span className="text-sm font-bold">
                          {history.round}
                        </span>
                      </div>
                      <div>
                        <div className="font-medium">
                          라운드 {history.round}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {new Date(history.timestamp).toLocaleString("ko-KR", {
                            month: "short",
                            day: "numeric",
                            hour: "2-digit",
                            minute: "2-digit",
                          })}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-6">
                      <div className="text-right">
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium">
                            {(history.accuracy * 100).toFixed(2)}%
                          </span>
                          {index > 0 && (
                            <span
                              className={`text-xs px-1.5 py-0.5 rounded ${
                                isImproving
                                  ? "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300"
                                  : "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300"
                              }`}
                            >
                              {isImproving ? "+" : ""}
                              {accuracyChange}%
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          정확도
                        </div>
                      </div>

                      <div className="text-right">
                        <div className="text-sm font-medium">
                          {history.loss.toFixed(4)}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          손실
                        </div>
                      </div>

                      <div className="text-right">
                        <div className="text-sm font-medium">
                          {history.participants}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          참여자
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* MLflow 정보 */}
      {aggregator.mlflowExperimentName && (
        <Card className="bg-gradient-to-br from-purple-50 to-indigo-50 dark:from-purple-950 dark:to-indigo-950 border-purple-200 dark:border-purple-800">
          <CardHeader>
            <CardTitle className="text-purple-700 dark:text-purple-300 flex items-center">
              <div className="w-2 h-2 bg-purple-500 rounded-full mr-2"></div>
              MLflow 실험 정보
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div className="bg-white dark:bg-gray-900 p-3 rounded-lg border">
                <p className="text-sm font-medium text-muted-foreground">
                  실험 이름
                </p>
                <p className="font-mono text-sm">
                  {aggregator.mlflowExperimentName}
                </p>
              </div>
              {aggregator.mlflowExperimentId && (
                <div className="bg-white dark:bg-gray-900 p-3 rounded-lg border">
                  <p className="text-sm font-medium text-muted-foreground">
                    실험 ID
                  </p>
                  <p className="font-mono text-sm">
                    {aggregator.mlflowExperimentId}
                  </p>
                </div>
              )}
            </div>
            <div className="flex space-x-2">
              <Button variant="outline" className="flex-1">
                MLflow에서 보기
              </Button>
              <Button variant="outline">실험 복사</Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default AggregatorDetails;