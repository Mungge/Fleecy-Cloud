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
import {
  Eye,
  Monitor,
  DollarSign,
  Settings,
  Activity,
  Server,
} from "lucide-react";
import AggregatorDetails from "@/components/dashboard/aggregator/aggregator-details";
import { getAggregators, AggregatorInfo } from "@/api/aggregator"; // aggregator.ts에서 getAggregators 함수 import

export interface AggregatorInstance {
  id: string;
  name: string;
  status: "running" | "completed" | "error" | "pending";
  algorithm: string;
  federatedLearningId: string;
  federatedLearningName: string;
  cloudProvider: string;
  region: string;
  instanceType: string;
  createdAt: string;
  lastUpdated: string;
  participants: number;
  rounds: number;
  currentRound: number;
  accuracy?: number;
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
}

const AggregatorManagementContent: React.FC = () => {
  const [aggregators, setAggregators] = useState<AggregatorInstance[]>([]);
  const [selectedAggregator, setSelectedAggregator] =
    useState<AggregatorInstance | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showDetails, setShowDetails] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // AggregatorInfo를 AggregatorInstance로 변환하는 함수
  const mapAggregatorInfoToInstance = (
    info: AggregatorInfo
  ): AggregatorInstance => {
    return {
      id: info.id,
      name: info.name, // ID 기반으로 이름 생성
      status: mapStatus(info.status),
      algorithm: info.algorithm,
      federatedLearningId: info.federated_learning?.id ?? "",
      federatedLearningName: info.project_name,
      cloudProvider: info.cloud_provider,
      region: info.region,
      instanceType: info.instance_type,
      createdAt: info.created_at,
      lastUpdated: info.updated_at,
      participants: info.participant_count,
      rounds: info.federated_learning?.rounds ?? 0,
      currentRound: info.status === "running" ? info.current_round : 0,
      accuracy: Number(info.federated_learning?.accuracy) ?? 0,
      cost: {
        current: info.current_cost,
        estimated: info.estimated_cost,
      },
      specs: {
        cpu: info.cpu_specs,
        memory: info.memory_specs,
        storage: info.storage_specs,
      },
      metrics: {
        cpuUsage: info.status === "running" ? info.cpu_usage : 0,
        memoryUsage: info.status === "running" ? info.memory_usage : 0,
        networkUsage: info.status === "running" ? info.network_usage : 0,
      },
    };
  };

  // 상태 매핑 함수
  const mapStatus = (
    status: string
  ): "running" | "completed" | "error" | "pending" => {
    const statusMap: {
      [key: string]: "running" | "completed" | "error" | "pending";
    } = {
      running: "running",
      completed: "completed",
      failed: "error",
      error: "error",
      pending: "pending",
      creating: "pending",
    };
    return statusMap[status] || "pending";
  };

  // aggregator.ts의 getAggregators 함수 사용
  useEffect(() => {
    const fetchAggregators = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const data = await getAggregators();

        // data가 배열이 아니거나 null/undefined인 경우 처리
        if (!Array.isArray(data)) {
          console.warn("API response is not an array:", data);
          setAggregators([]);
          return;
        }

        // 실제 API 응답을 AggregatorInstance로 변환
        const mappedData = data.map((item: AggregatorInfo, index) => {
          return mapAggregatorInfoToInstance(item);
        });

        setAggregators(mappedData);
      } catch (err) {
        console.error("Failed to fetch aggregators:", err);
        setError("집계자 정보를 불러오는데 실패했습니다.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchAggregators();
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
      </div>

      {/* 통계 카드 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">총 집계자</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{aggregators.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">실행 중</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {aggregators.filter((a) => a.status === "running").length}
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
              {aggregators.filter((a) => a.status === "completed").length}
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
              {formatCurrency(
                aggregators.reduce((sum, a) => sum + (a.cost?.current || 0), 0)
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
          ) : error ? (
            <div className="text-center py-8 text-red-500">
              <Server className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>{error}</p>
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
                              {aggregator.metrics.cpuUsage}%
                            </div>
                            <div>
                              <span className="font-medium">메모리:</span>{" "}
                              {aggregator.metrics.memoryUsage}%
                            </div>
                            <div>
                              <span className="font-medium">정확도:</span>{" "}
                              {aggregator.accuracy}%
                            </div>
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
