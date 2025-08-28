"use client";

import React, { useState, useEffect } from "react";
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
import { toast } from "sonner";
import {
  ArrowLeft,
  Pause,
  Square,
  RefreshCw,
  Server,
  Activity,
  Clock,
  Cpu,
  HardDrive,
  Network,
} from "lucide-react";
import {
  AggregatorInstance,
  TrainingHistory,
  aggregatorAPI,
} from "@/api/aggregator";

interface AggregatorDetailsProps {
  aggregator: AggregatorInstance;
  onBack: () => void;
  onUpdate?: () => void;
}

interface TrainingRound {
  round: number;
  accuracy: number;
  loss: number;
  duration: number;
  participantsCount: number;
  timestamp: string;
}

const AggregatorDetails: React.FC<AggregatorDetailsProps> = ({
  aggregator: initialAggregator,
  onBack,
  onUpdate,
}) => {
  const [aggregator, setAggregator] =
    useState<AggregatorInstance>(initialAggregator);
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [trainingHistory, setTrainingHistory] = useState<TrainingHistory[]>([]);
  const [realTimeMetrics, setRealTimeMetrics] = useState(
    aggregator.metrics || {
      cpuUsage: 0,
      memoryUsage: 0,
      networkUsage: 0,
    }
  );

  const fetchAggregatorDetails = async () => {
    try {
      setIsRefreshing(true);
      const [updatedAggregator, history] = await Promise.all([
        aggregatorAPI.getAggregator(aggregator.id),
        aggregatorAPI.getTrainingHistory(aggregator.id),
      ]);

      setAggregator(updatedAggregator);
      setTrainingHistory(history);
      setRealTimeMetrics(
        updatedAggregator.metrics || {
          cpuUsage: 0,
          memoryUsage: 0,
          networkUsage: 0,
        }
      );
    } catch (error) {
      console.error("Failed to fetch aggregator details:", error);
      toast.error("집계자 상세 정보를 불러오는데 실패했습니다.");
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleStatusUpdate = async (newStatus: string) => {
    try {
      setIsLoading(true);
      await aggregatorAPI.updateAggregatorStatus(aggregator.id, newStatus);
      await fetchAggregatorDetails(); // 상태 업데이트 후 최신 데이터 로드
      if (onUpdate) onUpdate(); // 부모 컴포넌트 업데이트

      toast.success(
        `집계자가 ${newStatus === "paused" ? "일시정지" : "중지"}되었습니다.`
      );
    } catch (error) {
      toast.error("집계자 상태 변경에 실패했습니다.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleControlAction = async (action: "pause" | "resume" | "stop") => {
    const statusMap = {
      pause: "paused",
      resume: "running",
      stop: "stopped",
    };

    await handleStatusUpdate(statusMap[action]);
  };

  useEffect(() => {
    fetchAggregatorDetails();

    // 실시간 메트릭 업데이트를 위한 폴링 (실행 중인 경우에만)
    let interval: NodeJS.Timeout;
    if (aggregator.status === "running") {
      interval = setInterval(async () => {
        try {
          const updatedAggregator = await aggregatorAPI.getAggregator(
            aggregator.id
          );
          setRealTimeMetrics(
            updatedAggregator.metrics || {
              cpuUsage: 0,
              memoryUsage: 0,
              networkUsage: 0,
            }
          );
          setAggregator(updatedAggregator);
        } catch (error) {
          console.error("Failed to update real-time metrics:", error);
        }
      }, 5000); // 5초마다 업데이트
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [aggregator.id, aggregator.status]);

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
      case "paused":
        return "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300";
      case "stopped":
        return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300";
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
      case "paused":
        return "일시정지";
      case "stopped":
        return "중지됨";
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

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}분 ${remainingSeconds}초`;
  };

  const progressPercentage =
    (aggregator.currentRound / aggregator.rounds) * 100;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="outline" onClick={onBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            뒤로 가기
          </Button>
          <div>
            <h1 className="text-3xl font-bold">{aggregator.name}</h1>
            <div className="flex items-center space-x-2 mt-2">
              <Badge className={getStatusColor(aggregator.status)}>
                {getStatusText(aggregator.status)}
              </Badge>
              <Badge variant="outline">{aggregator.algorithm}</Badge>
              <Badge variant="secondary">{aggregator.cloudProvider}</Badge>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            onClick={fetchAggregatorDetails}
            disabled={isRefreshing}
          >
            <RefreshCw
              className={`h-4 w-4 mr-2 ${isRefreshing ? "animate-spin" : ""}`}
            />
            새로고침
          </Button>

          {/* Control Buttons */}
          {(aggregator.status === "running" ||
            aggregator.status === "paused") && (
            <div className="flex space-x-2">
              {aggregator.status === "running" && (
                <Button
                  variant="outline"
                  onClick={() => handleControlAction("pause")}
                  disabled={isLoading}
                >
                  <Pause className="h-4 w-4 mr-2" />
                  일시정지
                </Button>
              )}
              {aggregator.status === "paused" && (
                <Button
                  variant="outline"
                  onClick={() => handleControlAction("resume")}
                  disabled={isLoading}
                >
                  <Activity className="h-4 w-4 mr-2" />
                  재시작
                </Button>
              )}
              <Button
                variant="destructive"
                onClick={() => handleControlAction("stop")}
                disabled={isLoading}
              >
                <Square className="h-4 w-4 mr-2" />
                중지
              </Button>
            </div>
          )}
        </div>
      </div>

      {/* Progress Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5" />
            <span>학습 진행 상황</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">진행률</span>
              <span className="text-sm text-muted-foreground">
                {aggregator.currentRound} / {aggregator.rounds} 라운드
              </span>
            </div>
            <Progress value={progressPercentage} className="h-3" />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <span className="font-medium">참여자:</span>
                <div className="text-lg font-bold">
                  {aggregator.participants}
                </div>
              </div>
              <div>
                <span className="font-medium">현재 정확도:</span>
                <div className="text-lg font-bold">
                  {aggregator.accuracy
                    ? `${aggregator.accuracy.toFixed(2)}%`
                    : "N/A"}
                </div>
              </div>
              <div>
                <span className="font-medium">연합학습:</span>
                <div className="text-sm">
                  {aggregator.federatedLearningName}
                </div>
              </div>
              <div>
                <span className="font-medium">상태:</span>
                <div className="text-sm">
                  {getStatusText(aggregator.status)}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Real-time Metrics */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Server className="h-5 w-5" />
                <span>실시간 메트릭</span>
              </div>
              {aggregator.status === "running" && (
                <RefreshCw className="h-4 w-4 animate-spin text-green-500" />
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    <Cpu className="h-4 w-4" />
                    <span className="text-sm font-medium">CPU 사용률</span>
                  </div>
                  <span className="text-sm font-bold">
                    {realTimeMetrics.cpuUsage.toFixed(1)}%
                  </span>
                </div>
                <Progress value={realTimeMetrics.cpuUsage} className="h-2" />
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    <HardDrive className="h-4 w-4" />
                    <span className="text-sm font-medium">메모리 사용률</span>
                  </div>
                  <span className="text-sm font-bold">
                    {realTimeMetrics.memoryUsage.toFixed(1)}%
                  </span>
                </div>
                <Progress value={realTimeMetrics.memoryUsage} className="h-2" />
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-2">
                    <Network className="h-4 w-4" />
                    <span className="text-sm font-medium">네트워크 사용률</span>
                  </div>
                  <span className="text-sm font-bold">
                    {realTimeMetrics.networkUsage.toFixed(1)}%
                  </span>
                </div>
                <Progress
                  value={realTimeMetrics.networkUsage}
                  className="h-2"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Instance Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Server className="h-5 w-5" />
              <span>인스턴스 정보</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="font-medium">인스턴스 타입:</span>
                <span>{aggregator.instanceType}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">리전:</span>
                <span>{aggregator.region}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">CPU:</span>
                <span>{aggregator.specs.cpu}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">메모리:</span>
                <span>{aggregator.specs.memory}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">스토리지:</span>
                <span>{aggregator.specs.storage}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">생성일:</span>
                <span>{formatDate(aggregator.createdAt)}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">마지막 업데이트:</span>
                <span>{formatDate(aggregator.lastUpdated)}</span>
              </div>
              {aggregator.cost && (
                <>
                  <div className="flex justify-between">
                    <span className="font-medium">현재 비용:</span>
                    <span className="font-bold">
                      {formatCurrency(aggregator.cost.current)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="font-medium">예상 총 비용:</span>
                    <span className="font-bold">
                      {formatCurrency(aggregator.cost.estimated)}
                    </span>
                  </div>
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Training History */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Clock className="h-5 w-5" />
            <span>학습 히스토리</span>
          </CardTitle>
          <CardDescription>각 라운드별 학습 결과 및 성능 지표</CardDescription>
        </CardHeader>
        <CardContent>
          {trainingHistory.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Clock className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>아직 학습 히스토리가 없습니다.</p>
            </div>
          ) : (
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {trainingHistory.map((round) => (
                <div
                  key={round.round}
                  className="border rounded-lg p-3 hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <Badge variant="outline">Round {round.round}</Badge>
                      <div className="text-sm">
                        <span className="font-medium">정확도:</span>{" "}
                        {(round.accuracy * 100).toFixed(2)}%
                      </div>
                      <div className="text-sm">
                        <span className="font-medium">손실:</span>{" "}
                        {round.loss.toFixed(4)}
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {formatDate(round.timestamp)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default AggregatorDetails;
