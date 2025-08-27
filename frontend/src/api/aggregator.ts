//import { interceptors } from "undici-types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Terraform 출력 타입 정의
export interface TerraformOutput {
  instanceId?: string;
  publicIp?: string;
  privateIp?: string;
  region?: string;
  [key: string]: unknown;
}

// Aggregator 정보 타입 정의
export interface AggregatorInfo {
  id: string;
  federatedLearningId: string;
  status: string;
  region: string;
  instanceType: string;
  storage: string;
  createdAt: string;
  updatedAt: string;
  terraformOutput?: TerraformOutput;
}

// Aggregator 배치 최적화설정 타입 정의
export interface AggregatorOptimizeConfig {
  maxBudget: number;
  maxLatency: number;
}

export interface AggregatorConfig {
  cloudProvider: string;
  region: string;
  instanceType: string;
  memory: number;
  projectId?: string; // GCP용 프로젝트 ID (선택적)
}

// 연합학습 데이터 타입 정의
export interface FederatedLearningData {
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
  modelFileSize?: number; // 모델 파일 크기 (바이트 단위)
}

// Aggregator 생성 요청 타입
export interface CreateAggregatorRequest {
  name: string;
  algorithm: string;
  region: string;
  storage: string; // GB as string
  instanceType: string;
  cloudProvider: string; // "aws" | "gcp"
  projectId?: string; // GCP용 프로젝트 ID (선택적)
}

// Aggregator 생성 응답 타입
export interface CreateAggregatorResponse {
  aggregatorId: string;
  status: string;
  terraformStatus?: string;
}

// Aggregator 생성 함수
export const createAggregator = async (
  federatedLearningData: FederatedLearningData,
  aggregatorConfig: AggregatorConfig
): Promise<CreateAggregatorResponse> => {
  try {
    // Derive simple storage size in GB (fallback to 20GB)
    const derivedStorageGB = Math.max(
      20,
      Math.ceil(
        (federatedLearningData.modelFileSize || 0) / (1024 * 1024 * 1024) + 5
      )
    );

    // 중복 방지를 위해 타임스탬프 추가
    const timestamp = new Date()
      .toISOString()
      .replace(/[:.]/g, "-")
      .slice(0, -5);
    const uniqueName = `${federatedLearningData.name}-${timestamp}`;

    const requestBody: CreateAggregatorRequest = {
      name: uniqueName,
      algorithm: federatedLearningData.algorithm,
      region: aggregatorConfig.region,
      storage: String(derivedStorageGB),
      instanceType: aggregatorConfig.instanceType,
      cloudProvider: aggregatorConfig.cloudProvider || "aws", // 기본값 AWS
      projectId: aggregatorConfig.projectId, // GCP용
    };

    const response = await fetch(`${API_URL}/api/aggregators`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      console.error("API 호출 실패:", errorData);

      // 중복 생성 에러에 대한 특별한 처리
      if (
        response.status === 400 &&
        errorData.error &&
        errorData.error.includes("동일한 이름의 집계자가 이미 존재합니다")
      ) {
        throw new Error(
          "동일한 이름의 집계자가 이미 존재합니다. 다른 이름을 사용해주세요."
        );
      }

      throw new Error(
        errorData.error || `HTTP error! status: ${response.status}`
      );
    }

    const data = await response.json();
    console.log("API 응답 성공:", data);
    return data.data;
  } catch (error) {
    console.error("Aggregator 생성에 실패했습니다:", error);
    throw error;
  }
};

// Aggregator 상태 조회 함수
export const getAggregatorStatus = async (
  aggregatorId: string
): Promise<{
  status: string;
  progress?: number;
  message?: string;
  terraformStatus?: string;
}> => {
  try {
    // Backend exposes GET /api/aggregators/{id}
    const response = await fetch(`${API_URL}/api/aggregators/${aggregatorId}`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    const agg = data?.data ?? data;
    return {
      status: agg.status,
      terraformStatus: agg.terraformStatus,
    };
  } catch (error) {
    console.error("Aggregator 상태 조회에 실패했습니다:", error);
    throw error;
  }
};

// Aggregator 목록 조회 함수
export const getAggregators = async (): Promise<AggregatorInfo[]> => {
  try {
    const response = await fetch(`${API_URL}/api/aggregators`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return Array.isArray(data) ? data : data.data;
  } catch (error) {
    console.error("Aggregator 목록 조회에 실패했습니다:", error);
    throw error;
  }
};

// Aggregator 삭제 함수
export const deleteAggregator = async (aggregatorId: string): Promise<void> => {
  try {
    const response = await fetch(`${API_URL}/api/aggregators/${aggregatorId}`, {
      method: "DELETE",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.error || `HTTP error! status: ${response.status}`
      );
    }
  } catch (error) {
    console.error("Aggregator 삭제에 실패했습니다:", error);
    throw error;
  }
};

// 집계자 배치 최적화 응답 타입
export interface AggregatorOption {
  rank: number;
  region: string;
  instanceType: string;
  cloudProvider: string;
  estimatedMonthlyCost: number;
  estimatedHourlyPrice: number;
  avgLatency: number;
  maxLatency: number;
  vcpu: number;
  memory: number;
  recommendationScore: number;
}
export interface OptimizationResponse {
  status: string;
  summary: {
    totalParticipants: number;
    participantRegions: string[];
    totalCandidateOptions: number;
    feasibleOptions: number;
    constraints: {
      maxBudget: number;
      maxLatency: number;
    };
    modelInfo: {
      modelType: string;
      algorithm: string;
      rounds: number;
    };
  };
  optimizedOptions: AggregatorOption[];
  message: string;
}

// 집계자 배치 최적화 함수
export const optimizeAggregatorPlacement = async (
  federatedLearningData: FederatedLearningData,
  constraints: {
    maxBudget: number;
    maxLatency: number;
    minMemoryRequirement?: number;
  }
): Promise<OptimizationResponse> => {
  try {
    const requestBody = {
      federatedLearning: federatedLearningData,
      aggregatorConfig: constraints,
    };

    const response = await fetch(`${API_URL}/api/aggregators/optimization`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(
        errorData.error || `HTTP error! status: ${response.status}`
      );
    }

    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error("집계자 배치 최적화에 실패했습니다:", error);
    throw error;
  }
};
