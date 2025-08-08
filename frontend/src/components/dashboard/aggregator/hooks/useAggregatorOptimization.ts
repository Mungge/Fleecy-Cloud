// hooks/useAggregatorOptimization.ts
import { useState } from "react";
import { toast } from "sonner";
import { optimizeAggregatorPlacement } from "@/api/aggregator";
import { 
  CreationStatus, 
  OptimizationResponse, 
  FederatedLearningData, 
  AggregatorOptimizeConfig 
} from "../aggregator.types";

export const useAggregatorOptimization = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [optimizationStatus, setOptimizationStatus] = useState<CreationStatus | null>(null);
  const [optimizationResults, setOptimizationResults] = useState<OptimizationResponse | null>(null);
  const [showAggregatorSelection, setShowAggregatorSelection] = useState(false);

  const handleAggregatorOptimization = async (
    federatedLearningData: FederatedLearningData,
    aggregatorOptimizeConfig: AggregatorOptimizeConfig
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
      // Aggregator 배치 최적화 실행
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