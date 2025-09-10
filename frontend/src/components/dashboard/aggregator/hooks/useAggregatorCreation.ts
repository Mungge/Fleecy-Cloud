// hooks/useAggregatorCreation.ts
import { useState } from "react";
// import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { createAggregator } from "@/api/aggregator";
import {
  CreationStatus,
  AggregatorOption,
  FederatedLearningData,
  AggregatorConfig,
  useAggregatorCreationStore,
} from "../aggregator.types";

export const useAggregatorCreation = () => {
  // const router = useRouter();
  const [isCreating, setIsCreating] = useState(false);
  const [creationStatus, setCreationStatus] = useState<CreationStatus | null>(
    null
  );
  const { setPayload } = useAggregatorCreationStore.getState();

  const handleCreateAggregator = async (
    selectedOption: AggregatorOption,
    federatedLearningData: FederatedLearningData,
    onSuccess?: () => void,
    onError?: (error: Error) => void
  ) => {
    setIsCreating(true);

    setCreationStatus({
      step: "deploying",
      message: `선택된 집계자를 배포하는 중... (${selectedOption.cloudProvider} ${selectedOption.region})`,
      progress: 50,
    });

    try {
      // API 구조에 맞게 config 생성
      const aggregatorConfig: AggregatorConfig = {
        cloudProvider: selectedOption.cloudProvider,
        region: selectedOption.region,
        instanceType: selectedOption.instanceType,
        memory: selectedOption.memory,
        estimatedCost: selectedOption.estimatedMonthlyCost,
      };

      // createAggregator API 호출
      const result = await createAggregator(
        federatedLearningData,
        aggregatorConfig
      );

      // aggregatorId를 payload에 저장 (한 번만)
      setPayload({
        federatedLearningData,
        aggregatorConfig,
        selectedOption,
        aggregatorId: result.aggregatorId,
      });

      setCreationStatus({
        step: "completed",
        message: "집계자가 성공적으로 생성되었습니다!",
        progress: 100,
      });

      toast.success(
        `집계자 생성이 완료되었습니다! (ID: ${result.aggregatorId})`
      );

      // 성공 콜백 실행
      if (onSuccess) {
        setTimeout(onSuccess, 2000);
      }
    } catch (error: unknown) {
      console.error("집계자 생성 실패:", error);
      const errorObj =
        error instanceof Error ? error : new Error("알 수 없는 오류");
      const errorMessage = errorObj.message;

      setCreationStatus({
        step: "error",
        message: errorMessage || "집계자 생성에 실패했습니다.",
        progress: 0,
      });

      toast.error(`집계자 생성에 실패했습니다: ${errorMessage}`);

      // 에러 콜백 실행
      if (onError) {
        onError(errorObj);
      }
    } finally {
      setIsCreating(false);
    }
  };

  const resetCreation = () => {
    setCreationStatus(null);
    setIsCreating(false);
  };

  return {
    isCreating,
    creationStatus,
    setCreationStatus,
    handleCreateAggregator,
    resetCreation,
  };
};
