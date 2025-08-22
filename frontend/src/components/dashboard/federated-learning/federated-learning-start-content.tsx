// @/components/dashboard/federated-learning/federated-learning-start-content.tsx
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAggregatorCreationStore } from "../aggregator/aggregator.types";
import { startFederatedLearning, getFirstActiveCloudConnection } from "@/api/federatedLearning";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
	Play,
	Server,
	HardDrive,
	Users,
	FileText,
	Layers,
	Clock,
	ArrowLeft,
	CheckCircle,
} from "lucide-react";

const FederatedLearningStartContent = () => {
	const router = useRouter();
	const payload = useAggregatorCreationStore((s) => s.payload);
	const [isStarting, setIsStarting] = useState(false);

	// payload가 없으면 이전 페이지로 리다이렉트
	useEffect(() => {
		if (!payload) {
			router.replace("/dashboard/federated-learning");
		}
	}, [payload, router]);

	const selectedOption = payload?.selectedOption;
	const federatedLearningData = payload?.federatedLearningData;

	const handleStartFederatedLearning = async (): Promise<void> => {
		setIsStarting(true);

		try {
			// payload 데이터를 사용하여 연합학습 시작 API 호출
			if (!payload || !selectedOption || !federatedLearningData) {
				throw new Error("필요한 데이터가 없습니다. 다시 시도해주세요.");
			}

			// 사용자의 첫 번째 활성 클라우드 연결 가져오기
			const cloudConnectionId = await getFirstActiveCloudConnection();

			// aggregatorId 확인
			if (!payload.aggregatorId) {
				throw new Error("Aggregator ID가 없습니다. 먼저 집계자를 배포해주세요.");
			}

			const result = await startFederatedLearning({
				aggregatorId: payload.aggregatorId,
				cloudConnectionId,
				federatedLearningData: {
					name: federatedLearningData.name,
					description: federatedLearningData.description || "",
					modelType: federatedLearningData.modelType || "CNN",
					algorithm: federatedLearningData.algorithm,
					rounds: federatedLearningData.rounds,
					participants: federatedLearningData.participants,
					modelFileName: federatedLearningData.modelFileName || undefined,
				},
			});

			console.log("연합학습 시작 성공:", result);
			
			toast.success("연합학습이 성공적으로 시작되었습니다!", {
				description: `연합학습 ID: ${result.federatedLearningId}`,
				duration: 5000,
			});
			
			// 성공 후 대시보드 또는 모니터링 페이지로 이동
			router.push("/dashboard/federated-learning");
		} catch (error) {
			console.error("연합학습 시작 실패:", error);
			
			const errorMessage = error instanceof Error 
				? error.message 
				: "연합학습 시작에 실패했습니다.";
			
			toast.error("연합학습 시작 실패", {
				description: errorMessage,
				duration: 5000,
			});
		} finally {
			setIsStarting(false);
		}
	};

	const handleGoBack = () => {
		router.back();
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
			{/* 헤더 */}
			<div className="flex items-center justify-between">
				<div className="space-y-2">
					<div className="flex items-center gap-3">
						<Play className="h-8 w-8 text-green-600" />
						<h1 className="text-3xl font-bold text-green-600">
							연합학습 시작 준비
						</h1>
					</div>
					<p className="text-muted-foreground">
						모든 설정이 완료되었습니다. 연합학습을 시작하시겠습니까?
					</p>
				</div>
				<Button variant="outline" onClick={handleGoBack}>
					<ArrowLeft className="h-4 w-4 mr-2" />
					이전으로
				</Button>
			</div>

			{/* 배포된 집계자 정보 */}
			<Card className="border-2 border-green-200 bg-green-50/30">
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<CheckCircle className="h-5 w-5 text-green-500" />
						배포된 집계자 정보
					</CardTitle>
					<CardDescription>성공적으로 배포된 집계자 인스턴스 정보</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-3">
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
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">상태</span>
							<Badge className="bg-green-500">활성</Badge>
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
									{((selectedOption.memory || 0) / 1024).toFixed(1)}GB
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

			{/* 연합학습 작업 정보 */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Layers className="h-5 w-5 text-blue-500" />
						연합학습 작업 정보
					</CardTitle>
					<CardDescription>실행할 연합학습 작업의 세부 정보</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="grid gap-4 md:grid-cols-2">
						<div className="space-y-3">
							<div className="flex items-center justify-between">
								<span className="text-sm font-medium">작업 이름</span>
								<span className="text-sm font-semibold">{federatedLearningData.name}</span>
							</div>
							<div className="flex items-center justify-between">
								<span className="text-sm font-medium">알고리즘</span>
								<Badge variant="outline">{federatedLearningData.algorithm}</Badge>
							</div>
							<div className="flex items-center justify-between">
								<span className="text-sm font-medium">라운드 수</span>
								<div className="flex items-center gap-1">
									<Clock className="h-3 w-3" />
									<span className="text-sm">{federatedLearningData.rounds}회</span>
								</div>
							</div>
						</div>
						<div className="space-y-3">
							<div className="flex items-center justify-between">
								<span className="text-sm font-medium">참여자 수</span>
								<div className="flex items-center gap-1">
									<Users className="h-3 w-3" />
									<span className="text-sm">
										{federatedLearningData.participants.length}명
									</span>
								</div>
							</div>
							{federatedLearningData.modelFileName && (
								<div className="flex items-center justify-between">
									<span className="text-sm font-medium">모델 파일</span>
									<div className="flex items-center gap-1">
										<FileText className="h-3 w-3" />
										<span className="text-sm font-mono">
											{federatedLearningData.modelFileName}
										</span>
									</div>
								</div>
							)}
							{federatedLearningData.description && (
								<div className="space-y-1">
									<span className="text-sm font-medium">설명</span>
									<p className="text-sm text-muted-foreground">
										{federatedLearningData.description}
									</p>
								</div>
							)}
						</div>
					</div>
				</CardContent>
			</Card>

			{/* 참여자 목록 */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Users className="h-5 w-5 text-purple-500" />
						참여자 목록
					</CardTitle>
					<CardDescription>
						연합학습에 참여할 {federatedLearningData.participants.length}명의 참여자
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
						{federatedLearningData.participants.map((participant, index) => (
							<div
								key={index}
								className="flex items-center justify-between p-3 border rounded-lg bg-muted/30"
							>
								<div className="flex items-center gap-2">
									<div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
										<span className="text-xs font-medium text-purple-600">
											{index + 1}
										</span>
									</div>
									<div>
										<p className="text-sm font-medium">{participant.name || `참여자 ${index + 1}`}</p>
										<p className="text-xs text-muted-foreground">
											{participant.id}
										</p>
									</div>
								</div>
								<Badge variant="secondary" className="text-xs">
									준비됨
								</Badge>
							</div>
						))}
					</div>
				</CardContent>
			</Card>

			{/* 시작 버튼 */}
			<div className="flex justify-center gap-4">
				<Button
					onClick={handleStartFederatedLearning}
					disabled={isStarting}
					size="lg"
					className="min-w-[200px] bg-green-600 hover:bg-green-700"
				>
					{isStarting ? (
						<>
							<div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
							연합학습 저장 중...
						</>
					) : (
						<>
							<Play className="h-4 w-4 mr-2" />
							연합학습 시작 및 저장
						</>
					)}
				</Button>
			</div>
		</div>
	);
};

export default FederatedLearningStartContent;
