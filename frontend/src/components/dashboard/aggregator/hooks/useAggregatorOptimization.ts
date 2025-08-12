// hooks/useAggregatorOptimization.ts
import { useState } from "react";
import { toast } from "sonner";
import { optimizeAggregatorPlacement, AggregatorOptimizeConfig } from "@/api/aggregator";
import { 
  CreationStatus, 
  OptimizationResponse, 
  FederatedLearningData, 
  ExtendedAggregatorOptimizeConfig 
} from "../aggregator.types";
import { calculateMinimumMemoryRequirement } from "../utils/modelMemoryCalculator";

export const useAggregatorOptimization = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [optimizationStatus, setOptimizationStatus] = useState<CreationStatus | null>(null);
  const [optimizationResults, setOptimizationResults] = useState<OptimizationResponse | null>(null);
  const [showAggregatorSelection, setShowAggregatorSelection] = useState(false);

  const handleAggregatorOptimization = async (
    federatedLearningData: FederatedLearningData,
    aggregatorOptimizeConfig: AggregatorOptimizeConfig,
    modelFileSize?: number
  ) => {
    if (!federatedLearningData) {
      toast.error("연합학습 정보가 없습니다.");
      return;
    }
    
    setIsLoading(true);
    setOptimizationStatus({
      step: "creating",
      message: "Aggregator 배치 최적화를 시작합니다.",
      progress: 5,
    });

    try {
      // 모델 파일 크기 기반 최소 메모리 요구사항 계산
      let extendedConfig: ExtendedAggregatorOptimizeConfig = {
        ...aggregatorOptimizeConfig
      };

      if (modelFileSize) {
        const minMemoryRequired = calculateMinimumMemoryRequirement(
          modelFileSize,
          federatedLearningData.participants.length,
          1.5 // 안전 계수
        );

        extendedConfig.minMemoryRequirement = minMemoryRequired;
        toast.info(
          `최소 메모리 요구사항: ${minMemoryRequired}GB (모델 크기 × 참여자 ${federatedLearningData.participants.length}명 × 1.5)`,
          { duration: 5000 }
        );
      }

      // Aggregator 배치 최적화 실행
      toast.info("집계자 배치 최적화를 실행합니다...");
      const optimizationResult: OptimizationResponse = await optimizeAggregatorPlacement(
        federatedLearningData,
        extendedConfig
      );
      
      if (optimizationResult.status === 'error') {
        throw new Error(optimizationResult.message);
      }

      // 메모리 요구사항을 충족하지 못하는 옵션 필터링 (클라이언트 사이드 검증)
      if (extendedConfig.minMemoryRequirement) {
        const filteredOptions = optimizationResult.optimizedOptions.filter(
          option => option.memory >= extendedConfig.minMemoryRequirement!
        );

        if (filteredOptions.length === 0) {
          throw new Error(
            `최소 메모리 요구사항(${extendedConfig.minMemoryRequirement}GB)을 충족하는 인스턴스가 없습니다.`
          );
        }

        // 필터링된 결과로 업데이트
        optimizationResult.optimizedOptions = filteredOptions;
        optimizationResult.summary.feasibleOptions = filteredOptions.length;
      }

      setOptimizationStatus({
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
      
      setOptimizationStatus({
        step: "error",
        message: errorMessage || "집계자 배치 최적화에 실패했습니다.",
        progress: 0,
      });
      toast.error(`집계자 배치 최적화에 실패했습니다: ${errorMessage}`);
    } finally {
      setIsLoading(false);
    }
  };

  const resetOptimization = () => {
    setOptimizationStatus(null);
    setOptimizationResults(null);
    setShowAggregatorSelection(false);
  };

  return {
    isLoading,
    optimizationStatus,
    setOptimizationStatus,
    optimizationResults,
    showAggregatorSelection,
    setShowAggregatorSelection,
    handleAggregatorOptimization,
    resetOptimization
  };
};