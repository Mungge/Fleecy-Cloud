"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { FederatedLearningData, AggregatorOptimizeConfig, CreationStatus, AggregatorOption } from "./aggregator.types";
import { useAggregatorOptimization } from "./hooks/useAggregatorOptimization";
import { useAggregatorCreation } from "./hooks/useAggregatorCreation";
import { AggregatorSelectionModal } from "./components/AggregatorSelectionModal";
import { ProgressSteps } from "./components/ProgressSteps";
import { CreationStatusDisplay } from "./components/CreationStatus";
import { FederatedLearningInfo } from "./components/FederatedLearningInfo";
import { AggregatorSettings } from "./components/AggregatorSettings";
import { useAggregatorCreationStore } from "./aggregator.types";

const AggregatorCreateContent = () => {
	const router = useRouter();
	const [federatedLearningData, setFederatedLearningData] = useState<FederatedLearningData | null>(null);
	const [modelFileSize, setModelFileSize] = useState<number>(0);
	const [aggregatorOptimizeConfig, setAggregatorOptimizeConfig] = useState<AggregatorOptimizeConfig>({
		maxBudget: 100000,
		maxLatency: 150,
	});

	// 최적화 훅
	const {
		isLoading: isOptimizing,
		optimizationStatus,
		setOptimizationStatus,
		optimizationResults,
		showAggregatorSelection,
		setShowAggregatorSelection,
		handleAggregatorOptimization,
		//resetOptimization
	} = useAggregatorOptimization();

	// 생성 훅
	const {
		isCreating,
		creationStatus: actualCreationStatus,
		resetCreation
	} = useAggregatorCreation();

	// 통합 상태 관리 (UI 표시용)
	const isLoading = isOptimizing || isCreating;
	const displayStatus: CreationStatus | null = actualCreationStatus || optimizationStatus;

	// 페이지 로드 시 sessionStorage에서 데이터 가져오기
	useEffect(() => {
		const savedData = sessionStorage.getItem("federatedLearningData");
		const savedFileSize = sessionStorage.getItem("modelFileSize");
		if (savedData) {
			try {
				const parsedData = JSON.parse(savedData);
				setFederatedLearningData(parsedData);
				if (savedFileSize) {
					const fileSizeInBytes = parseInt(savedFileSize, 10);
					setModelFileSize(fileSizeInBytes);
					console.log("모델 파일 크기 (MB):", (fileSizeInBytes / (1024 * 1024)).toFixed(2), "MB");
				}
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

	// 최적화 실행
	const onOptimize = () => {
		if (federatedLearningData) {
			// 이전 상태 초기화
			resetCreation();
			handleAggregatorOptimization(
				federatedLearningData, 
				aggregatorOptimizeConfig,
				);
		}
	};

	// 집계자 선택 후 생성
	const onSelectAggregator = (option: AggregatorOption) => {
		if (!federatedLearningData) return;
		// 최적화 상태를 생성 상태로 전환
		setShowAggregatorSelection(false);

		// 스토어에 페이로드 저장 후 배포 페이지로 이동
		const { setPayload } = useAggregatorCreationStore.getState();
		setPayload({
			federatedLearningData,
			aggregatorConfig: {
				cloudProvider: option.cloudProvider,
				region: option.region,
				instanceType: option.instanceType,
				memory: option.memory,
			},
			selectedOption: option,
		});

		// 세션 정리 및 이동
		sessionStorage.removeItem("federatedLearningData");
		sessionStorage.removeItem("modelFileName");
		sessionStorage.removeItem("modelFileSize");
		router.push("/dashboard/aggregator/deploy");
	};

	// 모달 취소 시
	const onCancelSelection = () => {
		setShowAggregatorSelection(false);
		// 최적화 상태 유지하되 선택 단계만 닫기
		setOptimizationStatus({
			step: "selecting",
			message: "집계자를 다시 선택하려면 버튼을 클릭하세요.",
			progress: 10,
		});
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
			<ProgressSteps creationStatus={displayStatus} isLoading={isLoading} />

			{/* 생성 상태 표시 */}
			<CreationStatusDisplay status={displayStatus} />

			<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
				{/* 연합학습 정보 요약 */}
				<FederatedLearningInfo data={federatedLearningData} />

				{/* Aggregator 설정 */}
				<AggregatorSettings
					config={aggregatorOptimizeConfig}
					onConfigChange={setAggregatorOptimizeConfig}
					onOptimize={onOptimize}
					isLoading={isLoading}
					creationStatus={displayStatus}
					//modelFileSize={modelFileSize}
					participantCount={federatedLearningData.participants.length}
				/>
			</div>

			{/* 집계자 선택 모달 */}
			{showAggregatorSelection && optimizationResults && (
				<AggregatorSelectionModal
					results={optimizationResults}
					onSelect={onSelectAggregator}
					onCancel={onCancelSelection}
				/>
			)}
		</div>
	);
};

export default AggregatorCreateContent;