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
import { toast } from "sonner";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import AggregatorDetails from "@/components/dashboard/aggregator/aggregator-details";
import { Eye, Monitor, Zap, CheckCircle2, Wallet, Trash2 } from "lucide-react";

export interface AggregatorInstance {
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
	// MLflow 관련 필드
	mlflowExperimentName?: string;
	mlflowExperimentId?: string;
}

// API 응답 타입 정의
interface ApiAggregatorResponse {
	id: string;
	name: string;
	status: "running" | "completed" | "error" | "pending" | "creating";
	algorithm: string;
	cloud_provider: string;
	region: string;
	instance_type: string;
	created_at: string;
	updated_at: string;
	participant_count?: number;
	current_round?: number;
	accuracy?: number;
	current_cost?: number;
	estimated_cost?: number;
	cpu_specs?: string;
	memory_specs?: string;
	storage_specs?: string;
	cpu_usage?: number;
	memory_usage?: number;
	network_usage?: number;
	mlflow_experiment_name?: string;
	mlflow_experiment_id?: string;
}

const AggregatorManagementContent: React.FC = () => {
	const [aggregators, setAggregators] = useState<AggregatorInstance[]>([]);
	const [selectedAggregator, setSelectedAggregator] =
		useState<AggregatorInstance | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [showDetails, setShowDetails] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// 인증 토큰 가져오기 (실제 구현에 맞게 수정 필요)
	const getAuthToken = () => {
		// 1. document.cookie는 "key1=value1; key2=value2; ..." 형태의 문자열을 반환합니다.
		const cookies = document.cookie.split(";");

		// 2. 모든 쿠키를 순회하며 'accessToken'을 찾습니다.
		for (let i = 0; i < cookies.length; i++) {
			const cookie = cookies[i].trim(); // 각 쿠키의 앞뒤 공백 제거

			// 3. 'accessToken='으로 시작하는 쿠키를 찾습니다.
			if (cookie.startsWith("token=")) {
				// 4. '=' 뒷부분의 토큰 값만 잘라서 반환합니다.
				return cookie.substring("token=".length, cookie.length);
			}
		}

		// 5. 'accessToken' 쿠키를 찾지 못하면 빈 문자열을 반환합니다.
		return "";
	};

	// API 호출을 위한 공통 함수
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

	// Aggregator 목록 조회 - useCallback으로 감싸서 의존성 문제 해결
	const fetchAggregators = useCallback(async () => {
		setIsLoading(true);
		setError(null);

		try {
			const response = await fetchWithAuth(
				"http://localhost:8080/api/aggregators"
			);

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: ApiAggregatorResponse[] = await response.json();

			// API 응답을 프론트엔드 인터페이스에 맞게 변환
			const transformedAggregators: AggregatorInstance[] = data.map(
				(agg: ApiAggregatorResponse) => ({
					id: agg.id,
					name: agg.name,
					status: agg.status,
					algorithm: agg.algorithm,
					federatedLearningName: agg.name, // 또는 별도 필드가 있으면 사용
					cloudProvider: agg.cloud_provider,
					region: agg.region,
					instanceType: agg.instance_type,
					createdAt: agg.created_at,
					lastUpdated: agg.updated_at,
					participants: agg.participant_count || 3, // 기본값
					rounds: 10, // 기본값 (실제로는 연합학습 설정에서 가져와야 함)
					currentRound: agg.current_round || 0,
					accuracy: agg.accuracy,
					cost: {
						current: agg.current_cost || 0,
						estimated: agg.estimated_cost || 0,
					},
					specs: {
						cpu: agg.cpu_specs || "2 vCPUs",
						memory: agg.memory_specs || "8 GB",
						storage: agg.storage_specs || "20 GB SSD",
					},
					metrics: {
						cpuUsage: agg.cpu_usage || 0,
						memoryUsage: agg.memory_usage || 0,
						networkUsage: agg.network_usage || 0,
					},
					mlflowExperimentName: agg.mlflow_experiment_name,
					mlflowExperimentId: agg.mlflow_experiment_id,
				})
			);

			setAggregators(transformedAggregators);
		} catch (error) {
			console.error("Aggregator 목록 조회 실패:", error);
			setError(
				"Aggregator 목록을 불러오는데 실패했습니다. 네트워크 연결을 확인해주세요."
			);
			setAggregators([]);
		} finally {
			setIsLoading(false);
		}
	}, [fetchWithAuth]); // fetchWithAuth를 dependency에 추가

	// 컴포넌트 마운트 시 데이터 로드
	useEffect(() => {
		fetchAggregators();
	}, [fetchAggregators]);

	// 주기적으로 데이터 새로고침 (실행 중인 aggregator가 있을 때)
	useEffect(() => {
		const hasRunningAggregators = aggregators.some(
			(agg) => agg.status === "running"
		);

		if (hasRunningAggregators) {
			const interval = setInterval(() => {
				fetchAggregators();
			}, 30000); // 30초마다 갱신

			return () => clearInterval(interval);
		}
	}, [aggregators, fetchAggregators]);

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

	const handleViewDetails = (aggregator: AggregatorInstance) => {
		setSelectedAggregator(aggregator);
		setShowDetails(true);
	};

	const handleRefresh = async () => {
		await fetchAggregators();
	};

	const handleDeleteAggregator = useCallback(
		async (aggregatorId: string, aggregatorName: string) => {
			try {
				const response = await fetchWithAuth(
					`http://localhost:8080/api/aggregators/${aggregatorId}`,
					{
						method: "DELETE",
					}
				);

				if (response.ok) {
					toast.success(
						`"${aggregatorName}" 집계자가 성공적으로 삭제되었습니다.`
					);
					// 목록을 다시 불러와서 UI 업데이트
					await fetchAggregators();
				} else {
					const errorData = await response.json();
					throw new Error(errorData.error || "삭제에 실패했습니다.");
				}
			} catch (error) {
				console.error("집계자 삭제 실패:", error);
				toast.error(
					`삭제 실패: ${
						error instanceof Error ? error.message : "알 수 없는 오류"
					}`
				);
			}
		},
		[fetchWithAuth, fetchAggregators]
	);

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleString("ko-KR");
	};

	const formatCurrency = (amount: number) => {
		return new Intl.NumberFormat("ko-KR", {
			style: "currency",
			currency: "USD",
		}).format(amount);
	};

	// 상세보기 모드
	if (showDetails && selectedAggregator) {
		const aggregatorWithAccuracy = {
			...selectedAggregator,
			accuracy:
				selectedAggregator.accuracy !== undefined
					? selectedAggregator.accuracy
					: 0,
		};

		return (
			<AggregatorDetails
				aggregator={aggregatorWithAccuracy}
				onBack={() => setShowDetails(false)}
			/>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<div>
					<h1 className="text-3xl font-bold">연합학습 집계자 관리</h1>
					<p className="text-muted-foreground mt-2">
						연합학습 집계자 인스턴스를 관리하고 모니터링합니다
					</p>
				</div>
				<Button onClick={handleRefresh} disabled={isLoading}>
					{isLoading ? "새로고침 중..." : "새로고침"}
				</Button>
			</div>

			{/* 통계 카드 */}
			<div className="grid grid-cols-1 md:grid-cols-4 gap-4">
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">총 Aggregator</CardTitle>
						<span className="h-4 w-4 text-muted-foreground">
							<Monitor className="h-4 w-4 text-muted-foreground" />
						</span>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{aggregators.length}</div>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">실행 중</CardTitle>
						<Zap className="h-4 w-4 text-muted-foreground" />
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
						<CheckCircle2 className="h-4 w-4 text-blue-500" />
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
						<Wallet className="h-4 w-4 text-muted-foreground" />
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">
							{formatCurrency(
								aggregators.reduce(
									(total, agg) => total + (agg.cost?.current || 0),
									0
								)
							)}
						</div>
					</CardContent>
				</Card>
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

			{/* Aggregator 목록 */}
			<Card>
				<CardHeader>
					<CardTitle>연합학습 집계자가 인스턴스</CardTitle>
					<CardDescription>
						활성화된 연합학습 집계자 인스턴스 목록
					</CardDescription>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<div className="flex justify-center items-center py-12">
							<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
						</div>
					) : aggregators.length === 0 && !error ? (
						<div className="text-center py-8 text-muted-foreground">
							<span className="mx-auto h-12 w-12 mb-4 opacity-50 text-4xl block">
								🖥️
							</span>
							<p>실행 중인 Aggregator가 없습니다.</p>
							<p className="text-sm mt-2">새로운 Aggregator를 생성해보세요.</p>
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
												{aggregator.mlflowExperimentName && (
													<Badge variant="secondary" className="text-xs">
														MLflow: {aggregator.mlflowExperimentName}
													</Badge>
												)}
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
														{aggregator.accuracy && (
															<div>
																<span className="font-medium">정확도:</span>{" "}
																{aggregator.accuracy}%
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
											<AlertDialog>
												<AlertDialogTrigger asChild>
													<Button variant="destructive" size="sm">
														<Trash2 className="h-4 w-4 mr-2" />
														삭제
													</Button>
												</AlertDialogTrigger>
												<AlertDialogContent>
													<AlertDialogHeader>
														<AlertDialogTitle>
															집계자 삭제 확인
														</AlertDialogTitle>
														<AlertDialogDescription>
															정말로 &ldquo;{aggregator.name}&rdquo; 집계자를
															삭제하시겠습니까? 이 작업은 되돌릴 수 없으며,
															관련된 모든 데이터가 영구적으로 삭제됩니다.
														</AlertDialogDescription>
													</AlertDialogHeader>
													<AlertDialogFooter>
														<AlertDialogCancel>취소</AlertDialogCancel>
														<AlertDialogAction
															onClick={() =>
																handleDeleteAggregator(
																	aggregator.id,
																	aggregator.name
																)
															}
															className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
														>
															삭제
														</AlertDialogAction>
													</AlertDialogFooter>
												</AlertDialogContent>
											</AlertDialog>
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
