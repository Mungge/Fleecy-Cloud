// frontend/src/lib/api-service.ts
/**
 * 백엔드 API와의 통신을 담당하는 서비스
 */

import {
	ResourceEstimate,
	Instance,
} from "@/components/dashboard/aggregator/aggregator-content";

// 런타임에 환경 변수 접근을 위한 타입 확장
declare global {
	interface Window {
		ENV?: {
			BACKEND_URL?: string;
		};
	}
}

// 백엔드 URL 가져오기 (Docker 환경을 위한 설정)
const getBackendUrl = (): string => {
	// 브라우저에서 실행 중인지 확인
	if (typeof window !== "undefined" && window.ENV && window.ENV.BACKEND_URL) {
		return window.ENV.BACKEND_URL;
	}
	// 서버 사이드 렌더링 또는 환경 변수가 없을 경우 기본값
	return process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";
};

// 사용자 입력 타입
export interface UserInput {
	max_clients: number;
	avg_model_size_mb: number;
	flops: number;
	upload_freq_min: number;
	aggregation_type: string;
}

// 리소스 추정 API 호출
export const estimateResources = async (
	input: UserInput
): Promise<ResourceEstimate> => {
	try {
		const response = await fetch(`${getBackendUrl()}/aggregator/estimate`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(input),
		});

		if (!response.ok) {
			throw new Error(`API 요청 실패: ${response.status}`);
		}

		return await response.json();
	} catch (error) {
		console.error("리소스 추정 API 호출 오류:", error);
		throw error;
	}
};

// 인스턴스 추천 API 호출
export const recommendInstances = async (
	input: UserInput
): Promise<{
	estimate: ResourceEstimate;
	recommendations: Instance[];
}> => {
	try {
		const response = await fetch(`${getBackendUrl()}/aggregator/recommend`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(input),
		});

		if (!response.ok) {
			throw new Error(`API 요청 실패: ${response.status}`);
		}

		return await response.json();
	} catch (error) {
		console.error("인스턴스 추천 API 호출 오류:", error);
		throw error;
	}
};
