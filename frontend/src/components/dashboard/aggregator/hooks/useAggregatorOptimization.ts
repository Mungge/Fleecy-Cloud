// hooks/useAggregatorOptimization.ts
import { useState } from "react";
import { toast } from "sonner";
import {
  optimizeAggregatorPlacement,
  AggregatorOptimizeConfig,
} from "@/api/aggregator";
import {
  CreationStatus,
  OptimizationResponse,
  FederatedLearningData,
  ExtendedAggregatorOptimizeConfig,
} from "../aggregator.types";
import { calculateMinimumMemoryRequirement } from "../utils/modelMemoryCalculator";
import { ModelAnalysis, formatModelSize } from "../utils/modelDefinitionParser";

export const useAggregatorOptimization = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [optimizationStatus, setOptimizationStatus] =
    useState<CreationStatus | null>(null);
  const [optimizationResults, setOptimizationResults] =
    useState<OptimizationResponse | null>(null);
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
      // sessionStorage에서 모델 분석 결과 가져오기
      const modelAnalysisData = sessionStorage.getItem("modelAnalysis");
      let modelAnalysis: ModelAnalysis | null = null;
      let minMemoryRequired = 0;

      if (modelAnalysisData) {
        try {
          modelAnalysis = JSON.parse(modelAnalysisData);
          if (modelAnalysis) {
            // 모델 분석 결과로부터 최소 메모리 요구사항 계산
            minMemoryRequired = calculateMinimumMemoryRequirement(
              modelAnalysis.modelSizeBytes,
              federatedLearningData?.participants?.length ?? 0,
              1.5 // 안전 계수
            );

            const weights = calculateWeights(
              aggregatorOptimizeConfig.weightBalance || 4
            );

            toast.info(
              `모델 분석 기반 최소 메모리 요구사항: ${minMemoryRequired}GB\n` +
                `(${formatModelSize(modelAnalysis.totalParams)} × 참여자 ${
                  federatedLearningData?.participants?.length ?? 0
                }명 × 1.5)` +
                `가중치 설정: 비용 ${(weights.costWeight * 100).toFixed(
                  0
                )}%, 지연시간 ${(weights.latencyWeight * 100).toFixed(0)}%`,
              { duration: 6000 }
            );
          }
        } catch (error) {
          console.error("모델 분석 데이터 파싱 실패:", error);
        }
      }

      // 분석 결과가 없을 경우 기본값 사용
      if (!modelAnalysis) {
        // 기본 모델 크기 (1M 파라미터 = 4MB)
        const defaultModelSizeBytes = 1000000 * 4;
        minMemoryRequired = calculateMinimumMemoryRequirement(
          defaultModelSizeBytes,
          federatedLearningData?.participants?.length ?? 0,
          1.5
        );

        toast.warning(
          `모델 분석 데이터를 찾을 수 없어 기본값을 사용합니다.\n최소 메모리 요구사항: ${minMemoryRequired}GB`,
          { duration: 4000 }
        );
      }

      // 확장된 설정에 메모리 요구사항 추가
      const extendedConfig: ExtendedAggregatorOptimizeConfig = {
        ...aggregatorOptimizeConfig,
        minMemoryRequirement: minMemoryRequired,
      };

      // Aggregator 배치 최적화 실행
      toast.info("집계자 배치 최적화를 실행합니다...");
      const optimizationResult: OptimizationResponse =
        await optimizeAggregatorPlacement(
          federatedLearningData,
          extendedConfig
        );

      if (optimizationResult.status === "error") {
        throw new Error(optimizationResult.message);
      }

      // 메모리 요구사항을 충족하지 못하는 옵션 필터링 (클라이언트 사이드 검증)
      if (
        extendedConfig.minMemoryRequirement &&
        extendedConfig.minMemoryRequirement > 0
      ) {
        const filteredOptions = optimizationResult.optimizedOptions.filter(
          (option) => option.memory >= extendedConfig.minMemoryRequirement!
        );

        if (filteredOptions.length === 0) {
          throw new Error(
            `최소 메모리 요구사항(${extendedConfig.minMemoryRequirement}GB)을 충족하는 인스턴스가 없습니다.`
          );
        }

        // 필터링된 결과로 업데이트
        optimizationResult.optimizedOptions = filteredOptions;
        optimizationResult.summary.feasibleOptions = filteredOptions.length;

        toast.success(
          `${filteredOptions.length}개의 적합한 집계자 옵션을 찾았습니다.`,
          { duration: 3000 }
        );
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
      const errorMessage =
        error instanceof Error ? error.message : "알 수 없는 오류";

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

  const calculateWeights = (weightBalance: number) => {
    const costWeight = weightBalance / 10;
    const latencyWeight = 1 - costWeight;
    return { costWeight, latencyWeight };
  };

  return {
    isLoading,
    optimizationStatus,
    setOptimizationStatus,
    optimizationResults,
    showAggregatorSelection,
    setShowAggregatorSelection,
    handleAggregatorOptimization,
    resetOptimization,
  };
};
