"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Badge } from "@/components/ui/badge";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Check, ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { createAggregator, optimizeAggregatorPlacement, AggregatorConfig } from "@/api/aggregator";


// 연합학습 데이터 타입 정의
interface FederatedLearningData {
	name: string;
	description: string;
	modelType: string;
	algorithm: string;
	rounds: number;
	participants: Array<{
		id: string;
		name: string;
		status: string;
		openstack_endpoint?: string;
	}>;
	modelFileName?: string | null;
}

const AggregatorCreateContent = () => {
	const router = useRouter();
	const [federatedLearningData, setFederatedLearningData] =
		useState<FederatedLearningData | null>(null);
	const [aggregatorConfig, setAggregatorConfig] = useState<AggregatorConfig>({
		region: "ap-northeast-2",
		storage: "20",
		instanceType: "m1.medium",
		maxBudget: 500000,
		maxLatency: 100,
	});
	const [isLoading, setIsLoading] = useState(false);
	const [creationStatus, setCreationStatus] = useState<{
		step: "creating" | "deploying" | "completed" | "error";
		message: string;
		progress?: number;
	} | null>(null);

	// 페이지 로드 시 sessionStorage에서 데이터 가져오기
	useEffect(() => {
		const savedData = sessionStorage.getItem("federatedLearningData");
		if (savedData) {
			try {
				const parsedData = JSON.parse(savedData);
				setFederatedLearningData(parsedData);
			} catch (error) {
				console.error("데이터 파싱 실패:", error);
				toast.error("저장된 연합학습 정보를 불러올 수 없습니다.");
				router.push("/dashboard/federated-learning");
			}
		} else {
			toast.error("연합학습 정보가 없습니다. 다시 시도해주세요.");
			router.push("/dashboard/federated-learning");
		}
	}, [router]);

	// 이전 단계로 돌아가기
	const handleGoBack = () => {
		router.push("/dashboard/federated-learning");
	};

	// Aggregator 생성 및 연합학습 생성
	const handleCreateAggregator = async () => {
		if (!federatedLearningData) {
			toast.error("연합학습 정보가 없습니다.");
			return;
		}

		setIsLoading(true);
		setCreationStatus({
			step: "creating",
			message: "Aggregator 배치 최적화를 시작합니다.",
			//message: "Aggregator 설정을 생성하고 있습니다...",
			progress: 10,
		});

		try {
			//0단계: Aggregator 배치 최적화
			toast.info("집계자 배치 최적화를 실행합니다...");
			const optimizationResult = await optimizeAggregatorPlacement(
				federatedLearningData,
				{
					maxBudget: aggregatorConfig.maxBudget,
					maxLatency: aggregatorConfig.maxLatency
				}
			);
			setCreationStatus({
				step: "deploying",
				message: "최적화 완료! 최적의 집계자 배치를 찾았습니다.",
				progress: 80,
			});
			toast.success(`최적화 완료! ${optimizationResult.optimizationResults.length}개의 최적해를 찾았습니다.`);
			// 최적화 결과를 사용자에게 표시하거나 다음 단계 진행
			// 여기서는 임시로 첫 번째 최적해 정보를 표시
			const bestOption = optimizationResult.optimizationResults[0];
			if (bestOption) {
				toast.info(`추천 배치: ${bestOption.cloudProvider} ${bestOption.region} (${bestOption.instanceType}) - 비용: ${bestOption.estimatedCost.toLocaleString()}원, 지연시간: ${bestOption.estimatedLatency}ms`);
			}
			setCreationStatus({
				step: "completed",
				message: "집계자 배치 최적화가 성공적으로 완료되었습니다!",
				progress: 100,
			});
			// 3단계 완료 상태로 업데이트 후 페이지 이동
			setTimeout(() => {
				// sessionStorage 정리
				sessionStorage.removeItem("federatedLearningData");
				sessionStorage.removeItem("modelFileName");
				router.push("/dashboard/federated-learning");
			}, 3000); // 결과를 좀 더 오래 보여주기 위해 3초로 변경
		} catch (error: unknown) {
			console.error("집계자 배치 최적화 실패:", error);
			const errorMessage =
				error instanceof Error ? error.message : "알 수 없는 오류";
			
			setCreationStatus({
				step: "error",
				message: errorMessage || "집계자 배치 최적화에 실패했습니다.",
				progress: 0,
			});	
			toast.error(`집계자 배치 최적화에 실패했습니다: ${errorMessage}`);
			} finally {
				setIsLoading(false);
			}

			
		// 	// 1단계: Aggregator 생성 요청
		// 	//toast.info("Aggregator 생성을 시작합니다...");

		// 	const response = await createAggregator(
		// 		federatedLearningData,
		// 		aggregatorConfig
		// 	);

		// 	setCreationStatus({
		// 		step: "deploying",
		// 		message: "Terraform을 이용하여 인프라를 배포하고 있습니다...",
		// 		progress: 50,
		// 	});

		// 	toast.info("Terraform으로 인프라를 배포 중입니다...");

		// 	// 2단계: 배포 상태 모니터링 (실제로는 polling이나 WebSocket으로 구현)
		// 	// 여기서는 시뮬레이션으로 처리
		// 	await new Promise((resolve) => setTimeout(resolve, 3000));

		// 	setCreationStatus({
		// 		step: "completed",
		// 		message: "Aggregator가 성공적으로 생성되었습니다!",
		// 		progress: 100,
		// 	});

		// 	// 성공 메시지 표시
		// 	toast.success(
		// 		`Aggregator가 성공적으로 생성되었습니다! (ID: ${response.aggregatorId})`
		// 	);

		// 	// 3단계 완료 상태로 Progress bar 업데이트 후 페이지 이동
		// 	setTimeout(() => {
		// 		// sessionStorage 정리
		// 		sessionStorage.removeItem("federatedLearningData");
		// 		sessionStorage.removeItem("modelFileName");

		// 		// 연합학습 목록 페이지로 이동
		// 		router.push("/dashboard/federated-learning");
		// 	}, 2000);
		// } catch (error: unknown) { 	 	
		// 	console.error("Aggregator 생성 실패:", error);

		// 	const errorMessage =
		// 		error instanceof Error ? error.message : "알 수 없는 오류";

		// 	setCreationStatus({
		// 		step: "error",
		// 		message: errorMessage || "Aggregator 생성에 실패했습니다.",
		// 		progress: 0,
		// 	});

		// 	toast.error(`Aggregator 생성에 실패했습니다: ${errorMessage}`);
		// } finally {
		// 	setIsLoading(false);
		// }
	};	
	if (!federatedLearningData) {
		return (
			<div className="flex justify-center items-center min-h-screen">
				<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* 헤더 */}
			<div className="flex items-center justify-between">
				<div className="flex items-center space-x-4">
					<Button variant="outline" onClick={handleGoBack}>
						<ArrowLeft className="mr-2 h-4 w-4" />
						이전 단계
					</Button>
					<div>
						<h2 className="text-3xl font-bold tracking-tight">
							연합학습 집계자 생성
						</h2>
						<p className="text-muted-foreground">
							연합학습을 위한 집계자 설정을 완료하세요.
						</p>
					</div>
				</div>
			</div>

			{/* Progress Steps */}
			<Card>
				<CardContent className="pt-6">
					<div className="w-full py-4">
						<div className="flex items-center justify-between max-w-2xl mx-auto">
							{/* Step 1: 정보 입력 (완료) */}
							<div className="flex flex-col items-center">
								<div className="flex items-center justify-center w-12 h-12 rounded-full bg-green-500 text-white text-lg font-medium shadow-lg">
									<Check className="w-6 h-6" />
								</div>
								<span className="mt-3 text-base font-medium text-green-600">
									정보 입력
								</span>
								<span className="mt-1 text-sm text-gray-500">
									연합학습 정보 설정
								</span>
							</div>

							{/* Connector Line (완료) */}
							<div className="flex-1 h-1 bg-green-500 mx-6 rounded-full"></div>

							{/* Step 2: 집계자 생성 (현재/완료) */}
							<div className="flex flex-col items-center">
								<div
									className={`flex items-center justify-center w-12 h-12 rounded-full text-white text-lg font-medium shadow-lg ${
										creationStatus?.step === "completed"
											? "bg-green-500"
											: "bg-blue-500"
									}`}
								>
									{creationStatus?.step === "completed" ? (
										<Check className="w-6 h-6" />
									) : isLoading ? (
										<div className="animate-spin rounded-full h-6 w-6 border-2 border-white border-t-transparent"></div>
									) : (
										"2"
									)}
								</div>
								<span
									className={`mt-3 text-base font-medium ${
										creationStatus?.step === "completed"
											? "text-green-600"
											: "text-blue-600"
									}`}
								>
									집계자 생성
								</span>
								<span className="mt-1 text-sm text-gray-500">집계자 설정</span>
							</div>

							{/* Connector Line */}
							<div
								className={`flex-1 h-1 mx-6 rounded-full ${
									creationStatus?.step === "completed"
										? "bg-green-500"
										: "bg-gray-200"
								}`}
							></div>

							{/* Step 3: 연합학습 생성 */}
							<div className="flex flex-col items-center">
								<div
									className={`flex items-center justify-center w-12 h-12 rounded-full text-lg font-medium ${
										creationStatus?.step === "completed"
											? "bg-green-500 text-white shadow-lg"
											: "bg-gray-200 text-gray-400"
									}`}
								>
									{creationStatus?.step === "completed" ? (
										<Check className="w-6 h-6" />
									) : (
										"3"
									)}
								</div>
								<span
									className={`mt-3 text-base ${
										creationStatus?.step === "completed"
											? "text-green-600 font-medium"
											: "text-gray-400"
									}`}
								>
									연합학습 생성
								</span>
								<span className="mt-1 text-sm text-gray-400">
									최종 생성 완료
								</span>
							</div>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* 생성 상태 표시 */}
			{creationStatus && (
				<Card>
					<CardContent className="pt-6">
						<div className="space-y-4">
							<div className="flex items-center justify-between">
								<h3 className="text-lg font-medium">배포 진행 상황</h3>
								<span className="text-sm text-gray-500">
									{creationStatus.progress}%
								</span>
							</div>

							{/* Progress Bar */}
							<div className="w-full bg-gray-200 rounded-full h-2">
								<div
									className={`h-2 rounded-full transition-all duration-500 ${
										creationStatus.step === "error"
											? "bg-red-500"
											: "bg-blue-500"
									}`}
									style={{ width: `${creationStatus.progress || 0}%` }}
								></div>
							</div>

							<p
								className={`text-sm ${
									creationStatus.step === "error"
										? "text-red-600"
										: "text-gray-600"
								}`}
							>
								{creationStatus.message}
							</p>
						</div>
					</CardContent>
				</Card>
			)}

			<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
				{/* 연합학습 정보 요약 */}
				<Card>
					<CardHeader>
						<CardTitle>연합학습 정보 요약</CardTitle>
						<CardDescription>
							이전 단계에서 설정한 연합학습 정보를 확인하세요.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">이름:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.name}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">설명:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.description || "-"}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">모델 유형:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.modelType}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">알고리즘:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.algorithm}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">라운드 수:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.rounds}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">참여자:</div>
							<div className="text-sm col-span-2">
								{federatedLearningData.participants.length}명
							</div>
						</div>
						{federatedLearningData.modelFileName && (
							<div className="grid grid-cols-3 gap-2">
								<div className="text-sm font-medium">모델 파일:</div>
								<div className="text-sm col-span-2">
									{federatedLearningData.modelFileName}
								</div>
							</div>
						)}

						{/* 참여자 목록 */}
						<div className="space-y-2">
							<div className="text-sm font-medium">참여자 목록:</div>
							<div className="space-y-1">
								{federatedLearningData.participants.map((participant) => (
									<div
										key={participant.id}
										className="flex items-center justify-between p-2 bg-gray-50 rounded"
									>
										<span className="text-sm">{participant.name}</span>
										<Badge
											variant={
												participant.status === "active"
													? "default"
													: "secondary"
											}
										>
											{participant.status === "active" ? "활성" : "비활성"}
										</Badge>
									</div>
								))}
							</div>
						</div>
					</CardContent>
				</Card>

				{/* Aggregator 설정 */}
				<Card>
					<CardHeader>
						<CardTitle>연합학습 집계자 설정</CardTitle>
						<CardDescription>
							연합학습을 위한 집계자의 리소스를 설정하세요.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="space-y-2">
							<Label htmlFor="region">리전</Label>
							<Select
								value={aggregatorConfig.region}
								onValueChange={(value) =>
									setAggregatorConfig((prev) => ({ ...prev, region: value }))
								}
							>
								<SelectTrigger>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="ap-northeast-2">
										아시아 태평양 (서울)
									</SelectItem>
									<SelectItem value="ap-northeast-1">
										아시아 태평양 (도쿄)
									</SelectItem>
									<SelectItem value="us-east-1">
										미국 동부 (버지니아 북부)
									</SelectItem>
									<SelectItem value="us-west-2">미국 서부 (오레곤)</SelectItem>
									<SelectItem value="eu-west-1">유럽 (아일랜드)</SelectItem>
								</SelectContent>
							</Select>
						</div>

						<div className="space-y-2">
							<Label htmlFor="instanceType">인스턴스 타입</Label>
							<Select
								value={aggregatorConfig.instanceType}
								onValueChange={(value) =>
									setAggregatorConfig((prev) => ({
										...prev,
										instanceType: value,
									}))
								}
							>
								<SelectTrigger>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="m1.small">
										m1.small (1 vCPU, 2GB RAM)
									</SelectItem>
									<SelectItem value="m1.medium">
										m1.medium (2 vCPU, 4GB RAM)
									</SelectItem>
									<SelectItem value="m1.large">
										m1.large (4 vCPU, 8GB RAM)
									</SelectItem>
									<SelectItem value="m1.xlarge">
										m1.xlarge (8 vCPU, 16GB RAM)
									</SelectItem>
								</SelectContent>
							</Select>
						</div>

						<div className="space-y-2">
							<Label htmlFor="storage">스토리지 (GB)</Label>
							<Select
								value={aggregatorConfig.storage}
								onValueChange={(value) =>
									setAggregatorConfig((prev) => ({ ...prev, storage: value }))
								}
							>
								<SelectTrigger>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="10">10 GB</SelectItem>
									<SelectItem value="20">20 GB</SelectItem>
									<SelectItem value="50">50 GB</SelectItem>
									<SelectItem value="100">100 GB</SelectItem>
								</SelectContent>
							</Select>
						</div>
						{/* 제약조건 설정 추가 */}
						<div className="space-y-2">
							<div className="flex justify-between items-center">
								<Label htmlFor="budget">최대 월 예산</Label>
								<span className="text-sm font-medium text-green-600">
									{aggregatorConfig.maxBudget.toLocaleString()}원
								</span>
							</div>
							<Slider
								id="budget"
								value={[aggregatorConfig.maxBudget]}
								onValueChange={([value]) => 
									setAggregatorConfig(prev => ({ ...prev, maxBudget: value }))
								}
								max={2000000}
								min={100000}
								step={100000}
								className="w-full"
							/>
							<div className="flex justify-between text-xs text-muted-foreground">
								<span>10만원</span>
								<span>200만원</span>
							</div>
						</div>

						<div className="space-y-2">
							<div className="flex justify-between items-center">
								<Label htmlFor="latency">최대 허용 지연시간</Label>
								<span className="text-sm font-medium text-blue-600">
									{aggregatorConfig.maxLatency}ms
								</span>
							</div>
							<Slider
								id="latency"
								value={[aggregatorConfig.maxLatency]}
								onValueChange={([value]) => 
									setAggregatorConfig(prev => ({ ...prev, maxLatency: value }))
								}
								max={500}
								min={20}
								step={10}
								className="w-full"
							/>
							<div className="flex justify-between text-xs text-muted-foreground">
								<span>20ms (매우 빠름)</span>
								<span>500ms (여유)</span>
							</div>
						</div>

						{/* 현재 설정 요약 */}
						<div className="mt-4 p-3 bg-gray-50 rounded-md">
							<div className="text-sm text-muted-foreground mb-1">제약조건:</div>
							<div className="text-sm">
								월 최대 <span className="font-medium text-green-600">{aggregatorConfig.maxBudget.toLocaleString()}원</span> 예산으로{" "}
								<span className="font-medium text-blue-600">{aggregatorConfig.maxLatency}ms</span> 이하의 응답속도 보장
							</div>
						</div>

						<div className="pt-4">
							<Button
								onClick={handleCreateAggregator}
								disabled={isLoading || creationStatus?.step === "completed"}
								className="w-full"
								variant={
									creationStatus?.step === "completed" ? "secondary" : "default"
								}
							>
								{isLoading ? (
									<>
										<div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-white mr-2"></div>
										{creationStatus?.message || "생성 중..."}
									</>
								) : creationStatus?.step === "completed" ? (
									<>
										<Check className="mr-2 h-4 w-4" />
										생성 완료
									</>
								) : creationStatus?.step === "error" ? (
									"다시 시도"
								) : (
									"집계자 배치 최적화 실행"
								)}
							</Button>

							{creationStatus?.step === "completed" && (
								<p className="text-sm text-green-600 text-center mt-2">
									잠시 후 연합학습 페이지로 이동합니다...
								</p>
							)}
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	);
};

export default AggregatorCreateContent;
