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
import { toast } from "sonner";
import {
  Eye,
  Monitor,
  DollarSign,
  Settings,
  Activity,
  Server,
  RefreshCw,
} from "lucide-react";
import AggregatorDetails from "@/components/dashboard/aggregator/aggregator-details";
import {
  AggregatorInstance,
  AggregatorStats,
  aggregatorAPI,
} from "@/api/aggregator";

const AggregatorManagementContent: React.FC = () => {
  const [aggregators, setAggregators] = useState<AggregatorInstance[]>([]);
  const [stats, setStats] = useState<AggregatorStats | null>(null);
  const [selectedAggregator, setSelectedAggregator] =
    useState<AggregatorInstance | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showDetails, setShowDetails] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const [aggregatorsData, statsData] = await Promise.all([
        aggregatorAPI.getAggregators(),
        aggregatorAPI.getAggregatorStats(),
      ]);

      setAggregators(aggregatorsData);
      setStats(statsData);
    } catch (error) {
      console.error("Failed to fetch aggregator data:", error);
      toast.error("집계자 데이터를 불러오는데 실패했습니다.");
    } finally {
      setIsLoading(false);
    }
  };

  const refreshData = async () => {
    try {
      setIsRefreshing(true);
      await fetchData();
      toast.success("최신 집계자 정보를 불러왔습니다.");
    } catch (error) {
      toast.error("데이터를 새로고침하는데 실패했습니다.");
    } finally {
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();

    // 실시간 업데이트를 위한 폴링 (30초마다)
    const interval = setInterval(fetchData, 30000);

    return () => clearInterval(interval);
  }, []);

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
      default:
        return "알 수 없음";
    }
  };

  const handleViewDetails = (aggregator: AggregatorInstance) => {
    setSelectedAggregator(aggregator);
    setShowDetails(true);
  };

  const handleDeleteAggregator = async (id: string) => {
    try {
      await aggregatorAPI.deleteAggregator(id);
      await fetchData(); // 데이터 다시 로드
      toast.success("집계자가 성공적으로 삭제되었습니다.");
    } catch (error) {
      toast.error("집계자 삭제에 실패했습니다.");
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

  if (showDetails && selectedAggregator) {
    return (
      <AggregatorDetails
        aggregator={selectedAggregator}
        onBack={() => setShowDetails(false)}
        onUpdate={fetchData}
      />
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">집계자 관리</h1>
          <p className="text-muted-foreground mt-2">
            연합학습 집계자 인스턴스를 관리하고 모니터링합니다
          </p>
        </div>
        <Button
          variant="outline"
          onClick={refreshData}
          disabled={isRefreshing}
          className="flex items-center gap-2"
        >
          <RefreshCw
            className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
          />
          새로고침
        </Button>
      </div>

      {/* 통계 카드 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">총 집계자</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats ? stats.totalAggregators : aggregators.length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">실행 중</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats
                ? stats.runningAggregators
                : aggregators.filter((a) => a.status === "running").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">완료됨</CardTitle>
            <Badge className="h-4 w-4 rounded-full bg-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats
                ? stats.completedAggregators
                : aggregators.filter((a) => a.status === "completed").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">총 비용</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats
                ? formatCurrency(stats.totalCost)
                : formatCurrency(
                    aggregators.reduce(
                      (sum, a) => sum + (a.cost?.current || 0),
                      0
                    )
                  )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Aggregator 목록 */}
      <Card>
        <CardHeader>
          <CardTitle>집계자 인스턴스</CardTitle>
          <CardDescription>
            활성화된 연합학습 집계자 인스턴스 목록
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center items-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
            </div>
          ) : aggregators.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Server className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>실행 중인 Aggregator가 없습니다.</p>
            </div>
          ) : (
            <div className="space-y-4">
              {aggregators.map((aggregator) => (
                <div
                  key={aggregator.id}
                  className="border rounded-lg p-4 hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="space-y-2 flex-1">
                      <div className="flex items-center space-x-3">
                        <h3 className="font-semibold text-lg">
                          {aggregator.name}
                        </h3>
                        <Badge className={getStatusColor(aggregator.status)}>
                          {getStatusText(aggregator.status)}
                        </Badge>
                        <Badge variant="outline">{aggregator.algorithm}</Badge>
                      </div>

                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm text-muted-foreground">
                        <div>
                          <span className="font-medium">연합학습:</span>
                          <br />
                          {aggregator.federatedLearningName}
                        </div>
                        <div>
                          <span className="font-medium">클라우드:</span>
                          <br />
                          {aggregator.cloudProvider} ({aggregator.region})
                        </div>
                        <div>
                          <span className="font-medium">진행률:</span>
                          <br />
                          {aggregator.currentRound}/{aggregator.rounds} 라운드
                        </div>
                        <div>
                          <span className="font-medium">참여자:</span>
                          <br />
                          {aggregator.participants}개
                        </div>
                      </div>

                      {aggregator.status === "running" && (
                        <div className="mt-2">
                          <div className="flex items-center space-x-4 text-sm">
                            <div>
                              <span className="font-medium">CPU:</span>{" "}
                              {(aggregator.metrics?.cpuUsage ?? 0).toFixed(2)}%
                            </div>
                            <div>
                              <span className="font-medium">메모리:</span>{" "}
                              {(aggregator.metrics?.memoryUsage ?? 0).toFixed(
                                2
                              )}
                              %
                            </div>
                            {aggregator.accuracy && (
                              <div>
                                <span className="font-medium">정확도:</span>{" "}
                                {aggregator.accuracy.toFixed(2)}%
                              </div>
                            )}
                          </div>
                        </div>
                      )}

                      <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                        <div>생성일: {formatDate(aggregator.createdAt)}</div>
                        <div>
                          마지막 업데이트: {formatDate(aggregator.lastUpdated)}
                        </div>
                        {aggregator.cost && (
                          <div className="font-medium text-foreground">
                            비용: {formatCurrency(aggregator.cost.current)}
                            {aggregator.status === "running" && (
                              <span className="text-muted-foreground">
                                / 예상{" "}
                                {formatCurrency(aggregator.cost.estimated)}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="flex flex-col space-y-2 ml-4">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleViewDetails(aggregator)}
                      >
                        <Eye className="h-4 w-4 mr-2" />
                        상세 보기
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={aggregator.status !== "running"}
                      >
                        <Monitor className="h-4 w-4 mr-2" />
                        모니터링
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={aggregator.status !== "running"}
                      >
                        <Settings className="h-4 w-4 mr-2" />
                        설정
                      </Button>
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

export default AggregatorManagementContent;
