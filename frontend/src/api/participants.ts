import Cookies from "js-cookie";
import {
	Participant,
	CreateParticipantRequest,
	VMMonitoringInfo,
	VMHealthCheckResult,
} from "@/types/federatedLearning";

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

// 참여자 업데이트 요청 인터페이스
export interface UpdateParticipantRequest {
	name?: string;
	type?: string;
	status?: string;
	metadata?: string;

	// OpenStack 클라우드 관련 필드
	openstack_endpoint?: string;
	openstack_username?: string;
	openstack_password?: string;
	openstack_project_name?: string;
	openstack_domain_name?: string;
	openstack_region?: string;
}

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
	participantData: CreateParticipantRequest
): Promise<Participant> {
	try {
		const response = await fetch(`${API_BASE_URL}/api/participants`, {
			method: "POST",
			headers: getAuthHeaders(),
			credentials: "include",
			body: JSON.stringify(participantData),
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
	participantData: UpdateParticipantRequest
): Promise<Participant> {
	try {
		const response = await fetch(`${API_BASE_URL}/api/participants/${id}`, {
			method: "PUT",
			headers: getAuthHeaders(),
			credentials: "include",
			body: JSON.stringify(participantData),
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

// OpenStack VM 전원 제어 (시작/중지/재부팅)
export async function vmPowerAction(
	participantId: string,
	action: "start" | "stop" | "reboot"
): Promise<void> {
	try {
		const response = await fetch(
			`${API_BASE_URL}/api/participants/${participantId}/power`,
			{
				method: "POST",
				headers: getAuthHeaders(),
				credentials: "include",
				body: JSON.stringify({ action }),
			}
		);

		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}
	} catch (error) {
		console.error(`VM ${action} 실패:`, error);
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
