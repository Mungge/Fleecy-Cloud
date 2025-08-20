// @/components/dashboard/aggregator/aggregator-deploy.tsx
"use client";

import { useState, useEffect } from "react";
import { useAggregatorCreation } from "./hooks/useAggregatorCreation";
import { useAggregatorCreationStore } from "./aggregator.types";
import { useRouter } from "next/navigation";
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
	Server,
	HardDrive,
	Zap,
} from "lucide-react";

const AggregatorDeploy = () => {
	const router = useRouter();
	const { creationStatus, handleCreateAggregator, resetCreation } =
		useAggregatorCreation();
	const payload = useAggregatorCreationStore((s) => s.payload);
	const [hasStartedDeployment, setHasStartedDeployment] = useState(false);

	// payload에서 실제 선택된 옵션과 연합학습 데이터 사용
	const selectedOption = payload?.selectedOption;
	const federatedLearningData = payload?.federatedLearningData;

	// (선택) 없을 때의 안전장치: 사용자가 URL로 직접 들어온 경우 등
	// 간단한 가드 + 안내(필요 시 대시보드로 리다이렉트)
	useEffect(() => {
		if (!payload) {
			console.warn(
				"배포 페이로드가 없습니다. 이전 단계에서 구성을 완료해주세요."
			);
			// router.replace("/dashboard/aggregator"); // 필요하다면 활성화
		}
	}, [payload]);

	// 컴포넌트 마운트 시 자동으로 배포 시작 (한 번만 실행)
	useEffect(() => {
		if (!payload || hasStartedDeployment) {
			return;
		}

		setHasStartedDeployment(true);

		handleCreateAggregator(
			payload.selectedOption,
			payload.federatedLearningData,
			() => {
				// 배포 성공 처리
			},
			() => {
				// 배포 실패 처리 - 재시도하지 않도록 hasStartedDeployment를 true로 유지
			}
		);
	}, [payload, handleCreateAggregator, hasStartedDeployment]);

	useEffect(() => {
		if (!payload) {
			// 이전 단계 없이 직접 진입 시 되돌리기
			router.replace("/dashboard/federated-learning");
		}
	}, [payload, router]);

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

	// payload가 없으면 로딩 표시
	if (!payload || !selectedOption || !federatedLearningData) {
		return (
			<div className="flex justify-center items-center min-h-screen">
				<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
			</div>
		);
	}

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
					선택하신 {selectedOption?.cloudProvider || "클라우드"} 인스턴스로
					집계자를 배포하고 있습니다
				</p>
			</div>

			{/* 선택된 스펙 카드 */}
			<Card className="border-2 border-blue-200 bg-blue-50/30">
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Zap className="h-5 w-5 text-blue-500" />
						선택된 배포 스펙
					</CardTitle>
					<CardDescription>최적화에서 선택하신 인스턴스 구성</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-3">
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">순위</span>
							<Badge variant="secondary">#{selectedOption?.rank}</Badge>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">클라우드 제공자</span>
							<Badge className="bg-orange-500">
								{selectedOption?.cloudProvider}
							</Badge>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">리전</span>
							<Badge variant="outline">{selectedOption?.region}</Badge>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">인스턴스 타입</span>
							<div className="flex items-center gap-1">
								<Server className="h-3 w-3" />
								<span className="text-sm font-mono">
									{selectedOption?.instanceType}
								</span>
							</div>
						</div>
					</div>
					<div className="space-y-3">
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">vCPU</span>
							<span className="text-sm">{selectedOption?.vcpu} 코어</span>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">메모리</span>
							<div className="flex items-center gap-1">
								<HardDrive className="h-3 w-3" />
								<span className="text-sm">
									{((selectedOption?.memory || 0) / 1024).toFixed(1)}GB
								</span>
							</div>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">지연시간</span>
							<span className="text-sm text-green-600">
								{selectedOption?.avgLatency}ms
							</span>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">월 예상 비용</span>
							<span className="text-sm font-semibold">
								₩{selectedOption?.estimatedMonthlyCost.toLocaleString()}
							</span>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* 연합학습 정보 카드 */}
			<Card>
				<CardHeader>
					<CardTitle>연합학습 정보</CardTitle>
					<CardDescription>배포할 연합학습 작업 정보</CardDescription>
				</CardHeader>
				<CardContent className="space-y-3">
					<div className="flex items-center justify-between">
						<span className="text-sm font-medium">작업 이름</span>
						<span className="text-sm">{federatedLearningData.name}</span>
					</div>
					<div className="flex items-center justify-between">
						<span className="text-sm font-medium">알고리즘</span>
						<Badge variant="outline">{federatedLearningData.algorithm}</Badge>
					</div>
					<div className="flex items-center justify-between">
						<span className="text-sm font-medium">라운드 수</span>
						<span className="text-sm">{federatedLearningData.rounds}회</span>
					</div>
					<div className="flex items-center justify-between">
						<span className="text-sm font-medium">참여자 수</span>
						<span className="text-sm">
							{federatedLearningData.participants.length}명
						</span>
					</div>
					{federatedLearningData.modelFileName && (
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">모델 파일</span>
							<span className="text-sm font-mono">
								{federatedLearningData.modelFileName}
							</span>
						</div>
					)}
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
						onClick={() => {
							setHasStartedDeployment(false);
							window.location.reload();
						}}
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
};

export default AggregatorDeploy;
