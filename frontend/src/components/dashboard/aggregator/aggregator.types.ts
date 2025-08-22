import { OptimizationResponse, AggregatorOption } from "@/api/aggregator";
import { AggregatorOptimizeConfig, AggregatorConfig } from "@/api/aggregator";
import {FederatedLearningData } from "@/api/aggregator";
import { create } from "zustand";
export interface ModelFileInfo {
  fileName: string;
  fileSize: number; // bytes
  fileSizeInMB: number; // MB
}

export interface ExtendedAggregatorOptimizeConfig extends AggregatorOptimizeConfig {
  minMemoryRequirement?: number; // GB 단위
}
export interface CreationStatus {
  step: "creating" | "selecting" | "deploying" | "completed" | "error";
  message: string;
  progress?: number;
}

export interface AggregatorSelectionModalProps {
  results: OptimizationResponse;
  onSelect: (option: AggregatorOption) => void;
  onCancel: () => void;
}

export interface CreationStatusDisplayProps {
  status: CreationStatus | null;
}

export interface ProgressStepsProps {
  creationStatus: CreationStatus | null;
  isLoading: boolean;
}

export type AggregatorCreationPayload = {
	federatedLearningData: FederatedLearningData;
	aggregatorConfig: AggregatorConfig;
	// 선택된 옵션을 배포 UI에서 보여줘야 한다면 같이 보관
	selectedOption: {
	  rank: number;
	  region: string;
	  instanceType: string;
	  cloudProvider: string;
	  estimatedMonthlyCost: number;
	  estimatedHourlyPrice: number;
	  avgLatency: number;
	  maxLatency: number;
	  vcpu: number;
	  memory: number; // MB
	  recommendationScore: number;
	};
	// aggregator 생성 완료 후 저장되는 ID
	aggregatorId?: string;
  };

type State = {
payload: AggregatorCreationPayload | null;
setPayload: (p: AggregatorCreationPayload) => void;
clearPayload: () => void;
};

export const useAggregatorCreationStore = create<State>((set) => ({
	payload: null,
	setPayload: (p) => set({ payload: p }),
	clearPayload: () => set({ payload: null }),
  }));

// Re-export types from API for convenience
export type {
  OptimizationResponse,
  AggregatorOption,
  FederatedLearningData,
  AggregatorOptimizeConfig,
  AggregatorConfig
};