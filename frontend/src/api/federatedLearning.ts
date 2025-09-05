import { FederatedLearningJob } from "@/types/federated-learning";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Cloud Connection 타입 정의
interface CloudConnection {
  id: string;
  provider: string;
  name: string;
  region: string;
  status: string;
}

// 사용자의 첫 번째 활성 클라우드 연결을 가져오는 함수
export const getFirstActiveCloudConnection = async (): Promise<string> => {
  try {
    const response = await fetch(`${API_URL}/api/clouds`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    const clouds: CloudConnection[] = Array.isArray(data) ? data : data.data;

    // 활성 상태인 첫 번째 클라우드 연결 찾기
    const activeCloud = clouds.find((cloud) => cloud.status === "active");

    if (!activeCloud) {
      throw new Error("활성 상태인 클라우드 연결이 없습니다.");
    }

    return activeCloud.id;
  } catch (error) {
    console.error("클라우드 연결 조회에 실패했습니다:", error);
    throw error;
  }
};

export const getFederatedLearnings = async (): Promise<
  FederatedLearningJob[]
> => {
  try {
    const response = await fetch(`${API_URL}/api/federated-learning`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error("연합학습 목록을 가져오는데 실패했습니다:", error);
    throw error;
  }
};

export const getFederatedLearningById = async (
  id: string
): Promise<FederatedLearningJob> => {
  try {
    const response = await fetch(`${API_URL}/api/federated-learning/${id}`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error(`연합학습 ID ${id}를 가져오는데 실패했습니다:`, error);
    throw error;
  }
};

export const updateFederatedLearning = async (
  id: string,
  payload: Partial<FederatedLearningJob>
): Promise<FederatedLearningJob> => {
  try {
    const response = await fetch(`${API_URL}/api/federated-learning/${id}`, {
      method: "PUT",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data.data;
  } catch (error) {
    console.error(`연합학습 ID ${id} 업데이트에 실패했습니다:`, error);
    throw error;
  }
};

export const deleteFederatedLearning = async (id: string): Promise<void> => {
  try {
    const response = await fetch(`${API_URL}/api/federated-learning/${id}`, {
      method: "DELETE",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error(`연합학습 ID ${id} 삭제에 실패했습니다:`, error);
    throw error;
  }
};

// 연합학습 시작을 위한 데이터 저장 함수
export const startFederatedLearning = async (payload: {
  aggregatorId: string;
  cloudConnectionId: string;
  federatedLearningData: {
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
    modelFileName?: string;
  };
}): Promise<{
  federatedLearningId: string;
  aggregatorId: string;
  status: string;
}> => {
  try {
    const requestBody = {
      aggregatorId: payload.aggregatorId,
      cloudConnectionId: payload.cloudConnectionId,
      name: payload.federatedLearningData.name,
      description: payload.federatedLearningData.description,
      modelType: payload.federatedLearningData.modelType,
      algorithm: payload.federatedLearningData.algorithm,
      rounds: payload.federatedLearningData.rounds,
      participants: payload.federatedLearningData.participants,
      modelFileName: payload.federatedLearningData.modelFileName,
    };

    const response = await fetch(`${API_URL}/api/federated-learning`, {
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
    console.error("연합학습 시작에 실패했습니다:", error);
    throw error;
  }
};
