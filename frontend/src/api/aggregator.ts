//import { interceptors } from "undici-types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// 백엔드 models.Aggregator 타입 정의 (실제 API 응답)
export interface BackendAggregator {
  id: string;
  user_id: number;
  name: string;
  status: string;
  algorithm: string;
  cloud_provider: string;
  project_name: string;
  region: string;
  zone: string;
  instance_type: string;
  storage_specs: string;
  instance_id?: string;
  public_ip?: string;
  private_ip?: string;
  created_at: string;
  updated_at: string;
}

// 백엔드 TrainingRound 타입 정의
export interface BackendTrainingRound {
  id: number;
  aggregator_id: string;
  round_number: number;
  accuracy: number;
  loss: number;
  participants_count: number;
  duration: number;
  created_at: string;
}

// UI용 확장된 AggregatorInstance 타입
export interface AggregatorInstance {
  id: string;
  name: string;
  status:
    | "running"
    | "completed"
    | "error"
    | "pending"
    | "paused"
    | "stopped"
    | "creating"
    | "failed";
  algorithm: string;
  federatedLearningId: string;
  federatedLearningName: string;
  cloudProvider: string;
  region: string;
  instanceType: string;
  createdAt: string;
  lastUpdated: string;
  participants: number;
  rounds: number;
  currentRound: number;
  accuracy?: number;
  cost?: {
    current: number;
    estimated: number;
  };
  specs: {
    cpu: string;
    memory: string;
    storage: string;
  };
  metrics: {
    cpuUsage: number;
    memoryUsage: number;
    networkUsage: number;
  };
}

// 통계 타입 정의 (백엔드에서 map[string]interface{} 형태로 반환)
export interface AggregatorStats {
  totalAggregators: number;
  runningAggregators: number;
  completedAggregators: number;
  totalCost: number;
}

// 학습 히스토리 타입 정의
export interface TrainingHistory {
  round: number;
  accuracy: number;
  loss: number;
  timestamp: string;
  duration?: number;
  participantsCount?: number;
}

export interface TerraformOutput {
  instanceId?: string;
  publicIp?: string;
  privateIp?: string;
  region?: string;
  [key: string]: unknown;
}

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

export interface AggregatorOptimizeConfig {
  maxBudget: number;
  maxLatency: number;
}

export interface AggregatorConfig {
  cloudProvider: string;
  region: string;
  instanceType: string;
  memory: number;
  projectId?: string;
}

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
  modelFileSize?: number;
}

export interface CreateAggregatorRequest {
  name: string;
  algorithm: string;
  region: string;
  storage: string;
  instanceType: string;
  cloudProvider: string;
  projectId?: string;
}

export interface CreateAggregatorResponse {
  aggregatorId: string;
  status: string;
  terraformStatus?: string;
}

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

export const createAggregator = async (
  federatedLearningData: FederatedLearningData,
  aggregatorConfig: AggregatorConfig
): Promise<CreateAggregatorResponse> => {
  try {
    const derivedStorageGB = Math.max(
      20,
      Math.ceil(
        (federatedLearningData.modelFileSize || 0) / (1024 * 1024 * 1024) + 5
      )
    );

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
      cloudProvider: aggregatorConfig.cloudProvider || "aws",
      projectId: aggregatorConfig.projectId,
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

export const getAggregatorStatus = async (
  aggregatorId: string
): Promise<{
  status: string;
  progress?: number;
  message?: string;
  terraformStatus?: string;
}> => {
  try {
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

// 공통 fetch 함수
const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const token = localStorage.getItem("authToken");

  const response = await fetch(`${API_URL}${url}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      Authorization: token ? `Bearer ${token}` : "",
      ...options.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
};

// 백엔드 데이터를 UI 형태로 변환하는 함수
const transformBackendToUI = (
  backend: BackendAggregator
): AggregatorInstance => {
  // instanceType에서 스펙 추정
  const getSpecsFromInstanceType = (instanceType: string) => {
    // TODO: 인스턴스 타입별 스펙 정보는 백엔드 API를 통해 동적으로 받아오는 것이 이상적입니다.
    const specs: { [key: string]: { cpu: string; memory: string } } = {
      "t3.large": { cpu: "2 vCPUs", memory: "8 GB" },
      "t3.xlarge": { cpu: "4 vCPUs", memory: "16 GB" },
      "c5.xlarge": { cpu: "4 vCPUs", memory: "8 GB" },
      "n1-standard-4": { cpu: "4 vCPUs", memory: "15 GB" },
    };
    return specs[instanceType] || { cpu: "2 vCPUs", memory: "8 GB" };
  };

  const specs = getSpecsFromInstanceType(backend.instance_type);

  return {
    id: backend.id,
    name: backend.name,
    status: backend.status as AggregatorInstance["status"],
    algorithm: backend.algorithm,
    // TODO: federatedLearningId는 백엔드에서 관련 프로젝트 ID를 받아와야 합니다. 현재는 aggregator id를 임시로 사용합니다.
    federatedLearningId: backend.id,
    // TODO: federatedLearningName은 백엔드에서 관련 프로젝트 이름을 받아와야 합니다. 현재는 aggregator 이름을 기반으로 임시 생성합니다.
    federatedLearningName: `FL-${backend.name}`,
    cloudProvider: backend.cloud_provider,
    region: backend.region,
    instanceType: backend.instance_type,
    createdAt: backend.created_at,
    lastUpdated: backend.updated_at,
    // TODO: participants는 실제 학습에 참여하는 클라이언트 수를 백엔드에서 받아와야 합니다.
    participants: Math.floor(Math.random() * 5) + 3, // 임시값
    // TODO: rounds는 프로젝트 설정에서 정의된 총 라운드 수를 백엔드에서 받아와야 합니다.
    rounds: 10,
    // TODO: currentRound는 백엔드에서 실제 진행 중인 라운드 정보를 받아와야 합니다.
    currentRound:
      backend.status === "running" ? Math.floor(Math.random() * 10) : 0,
    // TODO: accuracy는 백엔드에서 최근 학습 라운드의 실제 정확도를 받아와야 합니다.
    accuracy:
      backend.status === "running" ? Math.random() * 30 + 70 : undefined,
    // TODO: cost 관련 데이터는 백엔드에서 실제 비용 정보를 받아와야 합니다.
    cost: {
      current: Math.random() * 50 + 10,
      estimated: Math.random() * 100 + 50,
    },
    specs: {
      cpu: specs.cpu,
      memory: specs.memory,
      storage: backend.storage_specs,
    },
    // TODO: metrics 관련 데이터는 백엔드 모니터링 시스템에서 실제 사용량 데이터를 받아와야 합니다.
    metrics: {
      cpuUsage: backend.status === "running" ? Math.random() * 60 + 20 : 0,
      memoryUsage: backend.status === "running" ? Math.random() * 50 + 30 : 0,
      networkUsage: backend.status === "running" ? Math.random() * 40 + 20 : 0,
    },
  };
};

// 백엔드 TrainingRound를 UI 형태로 변환
const transformTrainingHistory = (
  rounds: BackendTrainingRound[]
): TrainingHistory[] => {
  return rounds.map((round) => ({
    round: round.round_number,
    accuracy: round.accuracy,
    loss: round.loss,
    timestamp: round.created_at,
    duration: round.duration,
    participantsCount: round.participants_count,
  }));
};

// Aggregator 목록 조회
export const getAggregators = async (): Promise<BackendAggregator[]> => {
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

    // 백엔드는 배열을 직접 반환
    const data = await response.json();
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error("Aggregator 목록 조회에 실패했습니다:", error);
    throw error;
  }
};

// UI용 Aggregator 목록 조회
export const getAggregatorsExtended = async (): Promise<
  AggregatorInstance[]
> => {
  const backendAggregators = await getAggregators();
  return backendAggregators.map(transformBackendToUI);
};

// Aggregator 통계 조회
export const getAggregatorStats = async (): Promise<AggregatorStats> => {
  const stats = await fetchWithAuth("/api/aggregators/stats");

  // 백엔드에서 map[string]interface{} 형태로 반환하므로 적절히 변환
  return {
    totalAggregators: stats.totalAggregators || stats.total || 0,
    runningAggregators: stats.runningAggregators || stats.running || 0,
    completedAggregators: stats.completedAggregators || stats.completed || 0,
    totalCost: stats.totalCost || stats.cost || 0,
  };
};

// 특정 Aggregator 조회
export const getAggregator = async (
  id: string
): Promise<AggregatorInstance> => {
  const response = await fetch(`${API_URL}/api/aggregators/${id}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const backendAggregator = await response.json();
  return transformBackendToUI(backendAggregator);
};

// Aggregator 상태 업데이트
export const updateAggregatorStatus = async (
  id: string,
  status: string
): Promise<void> => {
  await fetchWithAuth(`/api/aggregators/${id}/status`, {
    method: "PUT",
    body: JSON.stringify({ status }),
  });
};

// Aggregator 메트릭 업데이트
export const updateAggregatorMetrics = async (
  id: string,
  metrics: {
    cpuUsage: number;
    memoryUsage: number;
    networkUsage: number;
  }
): Promise<void> => {
  await fetchWithAuth(`/api/aggregators/${id}/metrics`, {
    method: "PUT",
    body: JSON.stringify({
      cpu_usage: metrics.cpuUsage,
      memory_usage: metrics.memoryUsage,
      network_usage: metrics.networkUsage,
    }),
  });
};

// Aggregator 학습 히스토리 조회
export const getTrainingHistory = async (
  id: string
): Promise<TrainingHistory[]> => {
  const rounds: BackendTrainingRound[] = await fetchWithAuth(
    `/api/aggregators/${id}/training-history`
  );
  return transformTrainingHistory(rounds);
};

// Aggregator 삭제
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

// API 클래스
class AggregatorAPI {
  getAggregators = getAggregatorsExtended;
  getAggregatorStats = getAggregatorStats;
  getAggregator = getAggregator;
  createAggregator = createAggregator;
  updateAggregatorStatus = updateAggregatorStatus;
  updateAggregatorMetrics = updateAggregatorMetrics;
  getTrainingHistory = getTrainingHistory;
  deleteAggregator = deleteAggregator;
  optimizeAggregatorPlacement = optimizeAggregatorPlacement;
}

export const aggregatorAPI = new AggregatorAPI();
