"use client";
import React, { useState, useEffect, useCallback } from "react";
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
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";

// 타입 정의 추가
interface RealTimeMetricsResponse {
	cpu_usage: number;
	memory_usage: number;
	network_usage: number;
	accuracy?: number;
	loss?: number;
	participants_connected: number;
	current_round: number;
	timestamp: string;
}

interface TrainingHistoryResponse {
	round: number;
	accuracy: number;
	loss: number;
	timestamp: string;
	f1_score: number;
	precision: number;
	recall: number;
	duration: number;
	participantsCount: number;
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
		accuracy: number; // undefined가 아님을 보장
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
	onDelete?: (aggregatorId: string) => void;
}

const AggregatorDetails: React.FC<AggregatorDetailsProps> = ({
	aggregator,
	onBack,
	onDelete,
}) => {
	const [realTimeMetrics, setRealTimeMetrics] = useState({
		cpuUsage: aggregator.metrics.cpuUsage,
		memoryUsage: aggregator.metrics.memoryUsage,
		networkUsage: aggregator.metrics.networkUsage,
		accuracy: aggregator.accuracy,
		loss: 0,
		participantsConnected: aggregator.participants,
		lastUpdated: new Date().toISOString(),
	});

	const [trainingHistory, setTrainingHistory] = useState<
		TrainingHistoryResponse[]
	>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

	// 인증 토큰 가져오기
	const getAuthToken = () => {
		const cookies = document.cookie.split(";");

		for (let i = 0; i < cookies.length; i++) {
			const cookie = cookies[i].trim(); // const로 수정

			if (cookie.startsWith("token=")) {
				return cookie.substring("token=".length, cookie.length);
			}
		}

		return "";
	};

	// API 호출을 위한 공통 함수 - useCallback으로 감싸기
	const fetchWithAuth = useCallback(
		async (url: string, options: RequestInit = {}) => {
			const token = getAuthToken();

			const defaultOptions: RequestInit = {
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${token}`,
				},
			};

			return fetch(url, {
				...defaultOptions,
				...options,
				headers: {
					...defaultOptions.headers,
					...options.headers,
				},
			});
		},
		[]
	);

	// 실시간 메트릭 조회 - useCallback으로 감싸기
	const fetchRealTimeMetrics = useCallback(async () => {
		if (aggregator.status !== "running") return;

		try {
			const response = await fetchWithAuth(
				`http://localhost:8080/api/aggregators/${aggregator.id}/metrics`
			);

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: RealTimeMetricsResponse = await response.json();

			setRealTimeMetrics({
				cpuUsage: data.cpu_usage || 0,
				memoryUsage: data.memory_usage || 0,
				networkUsage: data.network_usage || 0,
				accuracy: data.accuracy || aggregator.accuracy,
				loss: data.loss || 0,
				participantsConnected:
					data.participants_connected || aggregator.participants,
				lastUpdated: data.timestamp || new Date().toISOString(),
			});
		} catch (error) {
			console.error("실시간 메트릭 조회 실패:", error);
			setError("실시간 메트릭을 불러오는데 실패했습니다.");
		}
	}, [
		aggregator.id,
		aggregator.status,
		aggregator.accuracy,
		aggregator.participants,
		fetchWithAuth,
	]);

	// 학습 히스토리 조회 - useCallback으로 감싸기
	const fetchTrainingHistory = useCallback(async () => {
		setIsLoading(true);
		setError(null);

		try {
			const response = await fetchWithAuth(
				`http://localhost:8080/api/aggregators/${aggregator.id}/training-history`
			);

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: TrainingHistoryResponse[] = await response.json();
			setTrainingHistory(data);
		} catch (error) {
			console.error("학습 히스토리 조회 실패:", error);
		} finally {
			setIsLoading(false);
		}
	}, [aggregator.id, fetchWithAuth]);

	// Aggregator 삭제 함수
	const handleDelete = useCallback(async () => {
		try {
			const response = await fetchWithAuth(
				`http://localhost:8080/api/aggregators/${aggregator.id}`,
				{
					method: "DELETE",
				}
			);

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			toast.success(`"${aggregator.name}" 집계자가 성공적으로 삭제되었습니다.`);

			// 삭제 후 콜백 실행 (목록으로 돌아가기)
			if (onDelete) {
				onDelete(aggregator.id);
			}
			setDeleteDialogOpen(false);
			onBack();
		} catch (error) {
			console.error("Aggregator 삭제 실패:", error);
			toast.error(`"${aggregator.name}" 집계자 삭제에 실패했습니다.`);
		}
	}, [aggregator.id, aggregator.name, fetchWithAuth, onDelete, onBack]);

	// 컴포넌트 마운트 시 초기 데이터 로드
	useEffect(() => {
		fetchRealTimeMetrics();
		fetchTrainingHistory();
	}, [fetchRealTimeMetrics, fetchTrainingHistory]);

	// 실행 중인 경우 주기적으로 실시간 메트릭 업데이트
	useEffect(() => {
		if (aggregator.status === "running") {
			const interval = setInterval(() => {
				fetchRealTimeMetrics();
			}, 5000); // 5초마다 업데이트

			return () => clearInterval(interval);
		}
	}, [aggregator.status, fetchRealTimeMetrics]);

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
							연합학습 집계자 상세 정보 및 실시간 모니터링
						</p>
					</div>
				</div>
				<div className="flex space-x-2">
					<Button variant="outline" onClick={fetchRealTimeMetrics}>
						새로고침
					</Button>
					{aggregator.status === "running" && (
						<Button variant="destructive">중지</Button>
					)}
					<Button
						variant="destructive"
						onClick={() => setDeleteDialogOpen(true)}
						disabled={aggregator.status === "running"}
					>
						<Trash2 className="h-4 w-4 mr-2" />
						삭제
					</Button>
				</div>
			</div>

			{/* 에러 표시 */}
			{error && (
				<Card className="border-red-200 bg-red-50">
					<CardContent className="pt-6">
						<div className="flex items-center space-x-2 text-red-800">
							<span>⚠️</span>
							<span>{error}</span>
						</div>
					</CardContent>
				</Card>
			)}

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
								<p>{formatDate(new Date().toISOString())}</p>
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
			<Card>
				<CardHeader>
					<CardTitle>학습 진행 상황</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="space-y-4">
						<div className="flex justify-between items-center">
							<span className="text-sm font-medium">
								진행률: {aggregator.currentRound}/{aggregator.rounds} 라운드
							</span>
							<span className="text-sm text-muted-foreground">
								{progressPercentage.toFixed(1)}%
							</span>
						</div>
						<Progress value={progressPercentage} className="h-2" />

						<div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
							<div className="text-center p-4 bg-muted rounded-lg">
								<div className="text-2xl font-bold">
									{realTimeMetrics.participantsConnected}
								</div>
								<div className="text-sm text-muted-foreground">
									연결된 참여자
								</div>
							</div>
							<div className="text-center p-4 bg-muted rounded-lg">
								<div className="text-2xl font-bold">
									{realTimeMetrics.accuracy.toFixed(2)}%
								</div>
								<div className="text-sm text-muted-foreground">현재 정확도</div>
							</div>
							<div className="text-center p-4 bg-muted rounded-lg">
								<div className="text-2xl font-bold">
									{realTimeMetrics.loss.toFixed(4)}
								</div>
								<div className="text-sm text-muted-foreground">현재 손실</div>
							</div>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* 실시간 시스템 메트릭 */}
			<Card>
				<CardHeader>
					<CardTitle>시스템 메트릭</CardTitle>
					<CardDescription>실시간 시스템 리소스 사용량</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="space-y-6">
						<div className="space-y-2">
							<div className="flex justify-between">
								<span className="text-sm font-medium">CPU 사용률</span>
								<span className="text-sm text-muted-foreground">
									{realTimeMetrics.cpuUsage}%
								</span>
							</div>
							<Progress value={realTimeMetrics.cpuUsage} className="h-2" />
						</div>

						<div className="space-y-2">
							<div className="flex justify-between">
								<span className="text-sm font-medium">메모리 사용률</span>
								<span className="text-sm text-muted-foreground">
									{realTimeMetrics.memoryUsage}%
								</span>
							</div>
							<Progress value={realTimeMetrics.memoryUsage} className="h-2" />
						</div>

						<div className="space-y-2">
							<div className="flex justify-between">
								<span className="text-sm font-medium">네트워크 사용률</span>
								<span className="text-sm text-muted-foreground">
									{realTimeMetrics.networkUsage}%
								</span>
							</div>
							<Progress value={realTimeMetrics.networkUsage} className="h-2" />
						</div>
					</div>
				</CardContent>
			</Card>

			{/* 비용 정보 */}
			{aggregator.cost && (
				<Card>
					<CardHeader>
						<CardTitle>비용 정보</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div className="p-4 bg-muted rounded-lg">
								<div className="text-2xl font-bold text-green-600">
									{formatCurrency(aggregator.cost.current)}
								</div>
								<div className="text-sm text-muted-foreground">
									현재 사용 비용
								</div>
							</div>
							<div className="p-4 bg-muted rounded-lg">
								<div className="text-2xl font-bold text-blue-600">
									{formatCurrency(aggregator.cost.estimated)}
								</div>
								<div className="text-sm text-muted-foreground">
									예상 총 비용
								</div>
							</div>
						</div>
					</CardContent>
				</Card>
			)}

			{/* 학습 히스토리 */}
			<Card>
				<CardHeader>
					<CardTitle>학습 히스토리</CardTitle>
					<CardDescription>라운드별 정확도 및 손실 변화</CardDescription>
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
						<div className="space-y-4">
							{trainingHistory.slice(-10).map((history, index) => (
								<div
									key={index}
									className="flex items-center justify-between p-3 border rounded"
								>
									<div className="flex space-x-4">
										<div className="font-medium">라운드 {history.round}</div>
										<div className="text-sm text-muted-foreground">
											{formatDate(history.timestamp)}
										</div>
									</div>
									<div className="flex space-x-4 text-sm">
										<div>
											<span className="font-medium">정확도:</span>{" "}
											{(history.accuracy * 100).toFixed(2)}%
										</div>
										<div>
											<span className="font-medium">손실:</span>{" "}
											{history.loss.toFixed(4)}
										</div>
										<div>
											<span className="font-medium">참여자:</span>{" "}
											{history.participantsCount}
										</div>
									</div>
								</div>
							))}
						</div>
					)}
				</CardContent>
			</Card>

			{/* MLflow 정보 */}
			{aggregator.mlflowExperimentName && (
				<Card>
					<CardHeader>
						<CardTitle>MLflow 실험</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<p className="text-sm font-medium text-muted-foreground">
									실험 이름
								</p>
								<p className="font-mono">{aggregator.mlflowExperimentName}</p>
							</div>
							{aggregator.mlflowExperimentId && (
								<div>
									<p className="text-sm font-medium text-muted-foreground">
										실험 ID
									</p>
									<p className="font-mono">{aggregator.mlflowExperimentId}</p>
								</div>
							)}
						</div>
						<div className="mt-4">
							<Button variant="outline">MLflow에서 보기</Button>
						</div>
					</CardContent>
				</Card>
			)}

			{/* 삭제 확인 다이얼로그 */}
			<AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>집계자 삭제 확인</AlertDialogTitle>
						<AlertDialogDescription>
							정말로 <strong>&quot;{aggregator.name}&quot;</strong> 집계자를
							삭제하시겠습니까?
							<br />이 작업은 되돌릴 수 없으며, 관련된 모든 데이터가 영구적으로
							삭제됩니다.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>취소</AlertDialogCancel>
						<AlertDialogAction
							onClick={handleDelete}
							className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
						>
							삭제
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	);
};

export default AggregatorDetails;
