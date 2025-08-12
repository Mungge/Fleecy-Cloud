import { OptimizationResponse, AggregatorOption } from "@/api/aggregator";
import { AggregatorOptimizeConfig, AggregatorConfig } from "@/api/aggregator";
import {FederatedLearningData } from "@/api/aggregator";

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

// Re-export types from API for convenience
export type {
  OptimizationResponse,
  AggregatorOption,
  FederatedLearningData,
  AggregatorOptimizeConfig,
  AggregatorConfig
};