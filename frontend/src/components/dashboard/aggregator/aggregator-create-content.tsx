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
// import {
// 	Select,
// 	SelectContent,
// 	SelectItem,
// 	SelectTrigger,
// 	SelectValue,
// } from "@/components/ui/select";
import { Check, ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { OptimizationResponse, AggregatorOption } from "@/api/aggregator";
import { optimizeAggregatorPlacement, AggregatorOptimizeConfig, AggregatorConfig } from "@/api/aggregator";
import {FederatedLearningData } from "@/api/aggregator";
import { createAggregator } from "@/api/aggregator";

const AggregatorCreateContent = () => {
	const router = useRouter();
	const [federatedLearningData, setFederatedLearningData] =
		useState<FederatedLearningData | null>(null);
	const [aggregatorOptimizeConfig, setAggregatorOptimizeConfig] = useState<AggregatorOptimizeConfig>({
		maxBudget: 500000,
		maxLatency: 150,
	});
	// const [aggregatorConfig, setAggregatorConfig] = useState<AggregatorConfig>({
	// 	cloudProvider: "aws",
	// 	region: "ap-northeast-2",
	// 	instanceType: "t3.medium",
	// 	memory: 4,
	// });
	const [isLoading, setIsLoading] = useState(false);
	const [creationStatus, setCreationStatus] = useState<{
		step: "creating" | "selecting" | "deploying" | "completed" | "error";
		message: string;
		progress?: number;
	} | null>(null);

	const [optimizationResults, setOptimizationResults] = useState<OptimizationResponse | null>(null);
	const [showAggregatorSelection, setShowAggregatorSelection] = useState(false);

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

	const handleAggregatorOptimization = async () => {
		if (!federatedLearningData) {
			toast.error("연합학습 정보가 없습니다.");
			return;
		}
		setIsLoading(true);
		setCreationStatus({
			step: "creating",
			message: "Aggregator 배치 최적화를 시작합니다.",
			progress: 5,
		});

		try {
			// 0단계: Aggregator 배치 최적화
			toast.info("집계자 배치 최적화를 실행합니다...");
			const optimizationResult: OptimizationResponse = await optimizeAggregatorPlacement(
			  federatedLearningData,
			  {
				maxBudget: aggregatorOptimizeConfig.maxBudget,
				maxLatency: aggregatorOptimizeConfig.maxLatency
			  }
			);
			
			if (optimizationResult.status === 'error') {
				throw new Error(optimizationResult.message);
			}

			setCreationStatus({
				step: "selecting",
				message: "최적화 완료! 집계자를 선택해주세요.",
				progress: 15,
			  });

			toast.success(optimizationResult.message);

			// 최적화 결과가 있는 경우 선택 단계로 이동
			if (optimizationResult.optimizedOptions.length > 0) {
				setOptimizationResults(optimizationResult);
				setShowAggregatorSelection(true);
			  } else {
				throw new Error("사용 가능한 집계자 옵션이 없습니다.");
			  }
		} catch (error: unknown) {
			console.error("집계자 배치 최적화 실패:", error);
			const errorMessage = error instanceof Error ? error.message : "알 수 없는 오류";
			
			setCreationStatus({
			step: "error",
			message: errorMessage || "집계자 배치 최적화에 실패했습니다.",
			progress: 0,
			});
			toast.error(`집계자 배치 최적화에 실패했습니다: ${errorMessage}`);
		} finally {
			setIsLoading(false);
		}
  	};
	
	// 집계자 선택 컴포넌트
	const AggregatorSelectionModal = ({ 
		results, 
		onSelect, 
		onCancel 
	}: {
		results: OptimizationResponse;
		onSelect: (option: AggregatorOption) => void;
		onCancel: () => void;
	}) => {
		const [selectedOption, setSelectedOption] = useState<AggregatorOption | null>(null);
	
		return (
		<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
			<div className="bg-white rounded-lg p-6 max-w-6xl max-h-[80vh] overflow-y-auto">
			<h2 className="text-2xl font-bold mb-4">집계자 선택</h2>
			
			{/* 요약 정보 */}
			<div className="mb-6 p-4 bg-gray-100 rounded-lg">
				<h3 className="font-semibold mb-2">최적화 요약</h3>
				<div className="grid grid-cols-2 gap-4 text-sm">
				<div>참여자 수: {results.summary.totalParticipants}명</div>
				<div>참여자 지역: {results.summary.participantRegions.join(', ')}</div>
				<div>후보 옵션: {results.summary.totalCandidateOptions}개</div>
				<div>조건 만족 옵션: {results.summary.feasibleOptions}개</div>
				</div>
			</div>
	
			{/* 옵션 리스트 */}
			<div className="space-y-3 mb-6">
				{results.optimizedOptions.map((option) => (
				<div
					key={`${option.region}-${option.instanceType}`}
					className={`p-4 border rounded-lg cursor-pointer transition-colors ${
					selectedOption?.rank === option.rank
						? 'border-blue-500 bg-blue-50'
						: 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
					}`}
					onClick={() => setSelectedOption(option)}
				>
					<div className="flex justify-between items-start mb-2">
					<div className="flex items-center space-x-2">
						<span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-sm font-medium">
						#{option.rank}
						</span>
						<span className="font-semibold text-lg">
						{option.cloudProvider} {option.region}
						</span>
						<span className="bg-green-100 text-green-800 px-2 py-1 rounded text-sm">
						추천도: {option.recommendationScore}%
						</span>
					</div>
					<div className="text-right">
						<div className="text-2xl font-bold text-blue-600">
						₩{option.estimatedMonthlyCost.toLocaleString()}
						</div>
						<div className="text-sm text-gray-500">월 예상 비용</div>
					</div>
					</div>
	
					<div className="grid grid-cols-4 gap-4 mt-3">
					<div>
						<div className="text-sm text-gray-600">인스턴스</div>
						<div className="font-medium">{option.instanceType}</div>
						<div className="text-xs text-gray-500">
						{option.vcpu}vCPU, {option.memory}GB
						</div>
					</div>
					<div>
						<div className="text-sm text-gray-600">평균 지연시간</div>
						<div className="font-medium text-orange-600">{option.avgLatency}ms</div>
					</div>
					<div>
						<div className="text-sm text-gray-600">최대 지연시간</div>
						<div className="font-medium text-red-600">{option.maxLatency}ms</div>
					</div>
					<div>
						<div className="text-sm text-gray-600">시간당 비용</div>
						<div className="font-medium">${option.estimatedHourlyPrice}</div>
					</div>
					</div>
				</div>
				))}
			</div>
	
			{/* 버튼 */}
			<div className="flex justify-end space-x-3">
				<button
				onClick={onCancel}
				className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
				>
				취소
				</button>
				<button
				onClick={() => selectedOption && onSelect(selectedOption)}
				disabled={!selectedOption}
				className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
				>
				선택한 집계자로 생성
				</button>
			</div>
			</div>
		</div>
		);
	};
  

	// Aggregator 생성 및 연합학습 생성
	const handleCreateAggregator = async (selectedOption: AggregatorOption) => {
		setShowAggregatorSelection(false);
		setIsLoading(true);
		
		setCreationStatus({
		  step: "deploying",
		  message: `선택된 집계자를 배포하는 중... (${selectedOption.cloudProvider} ${selectedOption.region})`,
		  progress: 50,
		});
	  
		try {
		  // 🔥 기존 API 구조에 맞게 수정
		  const aggregatorConfig: AggregatorConfig = {
			cloudProvider: selectedOption.cloudProvider,
			region: selectedOption.region,
			instanceType: selectedOption.instanceType,
			memory: selectedOption.memory
		  };
	  
		  // 기존 createAggregator API 사용
		  const result = await createAggregator(
			federatedLearningData!,
			aggregatorConfig
		  );
		  
		  setCreationStatus({
			step: "completed",
			message: "집계자가 성공적으로 생성되었습니다!",
			progress: 100,
		  });
	  
		  toast.success(`집계자 생성이 완료되었습니다! (ID: ${result.aggregatorId})`);
	  
		  // 결과 표시 후 페이지 이동
		  setTimeout(() => {
			sessionStorage.removeItem("federatedLearningData");
			sessionStorage.removeItem("modelFileName");
			router.push("/dashboard/federated-learning");
		  }, 2000);
	  
		} catch (error: unknown) {
		  console.error("집계자 생성 실패:", error);
		  const errorMessage = error instanceof Error ? error.message : "알 수 없는 오류";
		  
		  setCreationStatus({
			step: "error",
			message: errorMessage || "집계자 생성에 실패했습니다.",
			progress: 0,
		  });
		  toast.error(`집계자 생성에 실패했습니다: ${errorMessage}`);
		} finally {
		  setIsLoading(false);
		}
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
						{/* 제약조건 설정 */}
						<div className="space-y-2">
							<div className="flex justify-between items-center">
								<Label htmlFor="budget">최대 월 예산 제약조건</Label>
								<span className="text-sm font-medium text-green-600">
									{aggregatorOptimizeConfig.maxBudget.toLocaleString()}원
								</span>
							</div>
							<Slider
								id="budget"
								value={[aggregatorOptimizeConfig.maxBudget]}
								onValueChange={([value]) => 
									setAggregatorOptimizeConfig(prev => ({ ...prev, maxBudget: value }))
								}
								max={1000000}
								min={50000}
								step={10000}
								className="w-full"
							/>
							<div className="flex justify-between text-xs text-muted-foreground">
								<span>10만원</span>
								<span>200만원</span>
							</div>
						</div>

						<div className="space-y-2">
							<div className="flex justify-between items-center">
								<Label htmlFor="latency">최대 허용 지연시간 제약조건</Label>
								<span className="text-sm font-medium text-blue-600">
									{aggregatorOptimizeConfig.maxLatency}ms
								</span>
							</div>
							<Slider
								id="latency"
								value={[aggregatorOptimizeConfig.maxLatency]}
								onValueChange={([value]) => 
									setAggregatorOptimizeConfig(prev => ({ ...prev, maxLatency: value }))
								}
								max={500}
								min={20}
								step={5}
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
								월 최대 <span className="font-medium text-green-600">{aggregatorOptimizeConfig.maxBudget.toLocaleString()}원</span> 예산으로{" "}
								<span className="font-medium text-blue-600">{aggregatorOptimizeConfig.maxLatency}ms</span> 이하의 응답속도 보장
							</div>
						</div>

						<div className="pt-4">
							<Button
								onClick={handleAggregatorOptimization}
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
			{showAggregatorSelection && optimizationResults && (
			<AggregatorSelectionModal
				results={optimizationResults}
				onSelect={handleCreateAggregator}
				onCancel={() => {
				setShowAggregatorSelection(false);
				setCreationStatus(null);
				}}
			/>
			)}
		</div>
	);
};

export default AggregatorCreateContent;
