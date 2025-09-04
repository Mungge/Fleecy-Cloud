import Cookies from "js-cookie";
import { Participant } from "@/types/participant";
import {
  VMMonitoringInfo,
  VMHealthCheckResult,
  OpenStackVMInstance,
  VirtualMachine,
  OptimalVM,
} from "@/types/virtual-machine";

export interface OptimalVMRequest {
  min_vcpus?: number;
  min_ram?: number;
  min_disk?: number;
  max_cpu_usage?: number;
  max_memory_usage?: number;
}

// 기본 API URL
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// API 헤더 생성 함수
const getAuthHeaders = () => {
  const token = Cookies.get("token");
  return {
    "Content-Type": "application/json",
    Authorization: token ? `Bearer ${token}` : "",
  };
};

// 파일 업로드용 헤더 생성 함수
const getAuthHeadersForFile = () => {
  const token = Cookies.get("token");
  return {
    Authorization: token ? `Bearer ${token}` : "",
  };
};

// 모든 참여자 조회
export async function getParticipants(): Promise<Participant[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants`, {
      method: "GET",
      headers: getAuthHeaders(),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data || [];
  } catch (error) {
    console.error("참여자 목록 조회 실패:", error);
    throw error;
  }
}

// 사용 가능한 참여자 조회
export async function getAvailableParticipants(): Promise<Participant[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants/available`, {
      method: "GET",
      headers: getAuthHeaders(),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data || [];
  } catch (error) {
    console.error("사용 가능한 참여자 조회 실패:", error);
    throw error;
  }
}

// 특정 참여자 조회
export async function getParticipant(id: string): Promise<Participant> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants/${id}`, {
      method: "GET",
      headers: getAuthHeaders(),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("참여자 조회 실패:", error);
    throw error;
  }
}

// 참여자 생성
export async function createParticipant(
  formData: FormData
): Promise<Participant> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants`, {
      method: "POST",
      headers: getAuthHeadersForFile(),
      credentials: "include",
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("참여자 생성 실패:", error);
    throw error;
  }
}

// 참여자 업데이트
export async function updateParticipant(
  id: string,
  formData: FormData
): Promise<Participant> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants/${id}`, {
      method: "PUT",
      headers: getAuthHeadersForFile(),
      credentials: "include",
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("참여자 업데이트 실패:", error);
    throw error;
  }
}

// 참여자 삭제
export async function deleteParticipant(id: string): Promise<void> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/participants/${id}`, {
      method: "DELETE",
      headers: getAuthHeaders(),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("참여자 삭제 실패:", error);
    throw error;
  }
}

// OpenStack VM 모니터링 정보 조회
export async function monitorVM(
  participantId: string
): Promise<VMMonitoringInfo> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/monitor`,
      {
        method: "GET",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("VM 모니터링 실패:", error);
    throw error;
  }
}

// OpenStack VM 헬스체크 수행
export async function healthCheckVM(
  participantId: string
): Promise<VMHealthCheckResult> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/health-check`,
      {
        method: "POST",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("VM 헬스체크 실패:", error);
    throw error;
  }
}

// 연합학습 작업 할당
export async function assignFederatedLearningTask(
  participantId: string,
  taskId: string
): Promise<void> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/assign-task`,
      {
        method: "POST",
        headers: getAuthHeaders(),
        credentials: "include",
        body: JSON.stringify({ task_id: taskId }),
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error("연합학습 작업 할당 실패:", error);
    throw error;
  }
}

// ===== VM 관련 API 함수들 =====

// 참여자의 VM 목록 조회
export async function getOpenStackVMs(
  participantId: string
): Promise<OpenStackVMInstance[]> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/vms/all`,
      {
        method: "GET",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data || [];
  } catch (error) {
    console.error("OpenStack VM 목록 조회 실패:", error);
    throw error;
  }
}

// 참여자의 최적 VM 선택
export async function getOptimalVM(
  participantId: string,
  criteria?: OptimalVMRequest
): Promise<OptimalVM> {
  try {
    // TODO: 기본 조건 설정
    const defaultCriteria: OptimalVMRequest = {
      min_vcpus: 1,
      min_ram: 512,
      min_disk: 5,
      max_cpu_usage: 100,
      max_memory_usage: 70.0,
    };

    const requestBody = { ...defaultCriteria, ...criteria };

    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/vms/select`,
      {
        method: "POST",
        headers: getAuthHeaders(),
        credentials: "include",
        body: JSON.stringify(requestBody),
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result;
  } catch (error) {
    console.error("Optimal VM 선택 실패:", error);
    throw error;
  }
}

// DB에 저장된 VM 목록 조회
export async function getVirtualMachines(
  participantId: string
): Promise<VirtualMachine[]> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/vms`,
      {
        method: "GET",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data || [];
  } catch (error) {
    console.error("VM 목록 조회 실패:", error);
    throw error;
  }
}

// 특정 VM 세부 정보 조회 (OpenStack에서 직접)
export async function getOpenStackVMDetails(
  participantId: string,
  instanceId: string
): Promise<OpenStackVMInstance> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/vms/openstack/${instanceId}`,
      {
        method: "GET",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("OpenStack VM 세부 정보 조회 실패:", error);
    throw error;
  }
}

// 특정 VM 모니터링 (OpenStack에서 직접)
export async function monitorOpenStackVM(
  participantId: string,
  instanceId: string
): Promise<VMMonitoringInfo> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/participants/${participantId}/vms/openstack/${instanceId}/monitor`,
      {
        method: "GET",
        headers: getAuthHeaders(),
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    return result.data;
  } catch (error) {
    console.error("OpenStack VM 모니터링 실패:", error);
    throw error;
  }
}
