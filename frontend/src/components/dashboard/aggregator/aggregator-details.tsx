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
import { AggregatorInstance } from "./aggregator-content";

interface AggregatorDetailsProps {
	aggregator: AggregatorInstance;
	onBack: () => void;
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
	aggregator,
	onBack,
}) => {
	const [isLoading, setIsLoading] = useState(false);
	const [trainingHistory, setTrainingHistory] = useState<TrainingRound[]>([]);
	const [realTimeMetrics, setRealTimeMetrics] = useState(aggregator.metrics);

	useEffect(() => {
		// Mock training history data
		const mockHistory: TrainingRound[] = Array.from(
			{ length: aggregator.currentRound },
			(_, i) => ({
				round: i + 1,
				accuracy: Math.min(0.5 + i * 0.05 + Math.random() * 0.1, 1.0),
				loss: Math.max(1.0 - i * 0.08 - Math.random() * 0.1, 0.1),
				duration: 120 + Math.random() * 60,
				participantsCount: aggregator.participants,
				timestamp: new Date(
					Date.now() - (aggregator.currentRound - i) * 3600000
				).toISOString(),
			})
		);
		setTrainingHistory(mockHistory);

		// Real-time metrics simulation
		if (aggregator.status === "running") {
			const interval = setInterval(() => {
				setRealTimeMetrics((prev) => ({
					cpuUsage: Math.max(
						20,
						Math.min(90, prev.cpuUsage + (Math.random() - 0.5) * 10)
					),
					memoryUsage: Math.max(
						30,
						Math.min(85, prev.memoryUsage + (Math.random() - 0.5) * 8)
					),
					networkUsage: Math.max(
						10,
						Math.min(70, prev.networkUsage + (Math.random() - 0.5) * 15)
					),
				}));
			}, 3000);

			return () => clearInterval(interval);
		}
	}, [aggregator.currentRound, aggregator.participants, aggregator.status]);

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

	const handleControlAction = async (action: "pause" | "resume" | "stop") => {
		setIsLoading(true);
		// Simulate API call
		setTimeout(() => {
			setIsLoading(false);
			console.log(`Action ${action} performed on aggregator ${aggregator.id}`);
		}, 1000);
	};

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

				{/* Control Buttons */}
				{aggregator.status === "running" && (
					<div className="flex space-x-2">
						<Button
							variant="outline"
							onClick={() => handleControlAction("pause")}
							disabled={isLoading}
						>
							<Pause className="h-4 w-4 mr-2" />
							일시정지
						</Button>
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
								<div className="text-lg font-bold">{aggregator.accuracy}%</div>
							</div>
							<div>
								<span className="font-medium">연합학습:</span>
								<div className="text-sm">
									{aggregator.federatedLearningName}
								</div>
							</div>
							<div>
								<span className="font-medium">예상 완료:</span>
								<div className="text-sm">
									{aggregator.status === "running" ? "2시간 후" : "완료됨"}
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
											<div className="text-sm">
												<span className="font-medium">소요시간:</span>{" "}
												{formatDuration(round.duration)}
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
