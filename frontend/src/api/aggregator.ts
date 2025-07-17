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

// Aggregator 설정 타입 정의
export interface AggregatorConfig {
	region: string;
	storage: string;
	instanceType: string;
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
}

// Aggregator 생성 요청 타입
export interface CreateAggregatorRequest {
	federatedLearning: FederatedLearningData;
	aggregatorConfig: AggregatorConfig;
}

// Aggregator 생성 응답 타입
export interface CreateAggregatorResponse {
	aggregatorId: string;
	federatedLearningId: string;
	status: string;
	terraformOutput?: TerraformOutput;
}

// Aggregator 생성 함수
export const createAggregator = async (
	federatedLearningData: FederatedLearningData,
	aggregatorConfig: AggregatorConfig
): Promise<CreateAggregatorResponse> => {
	try {
		const requestBody: CreateAggregatorRequest = {
			federatedLearning: federatedLearningData,
			aggregatorConfig: aggregatorConfig,
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
			throw new Error(
				errorData.error || `HTTP error! status: ${response.status}`
			);
		}

		const data = await response.json();
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
	terraformLogs?: string[];
}> => {
	try {
		const response = await fetch(
			`${API_URL}/api/aggregator/${aggregatorId}/status`,
			{
				credentials: "include",
				headers: {
					"Content-Type": "application/json",
				},
			}
		);

		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}

		const data = await response.json();
		return data.data;
	} catch (error) {
		console.error("Aggregator 상태 조회에 실패했습니다:", error);
		throw error;
	}
};

// Aggregator 목록 조회 함수
export const getAggregators = async (): Promise<AggregatorInfo[]> => {
	try {
		const response = await fetch(`${API_URL}/api/aggregator`, {
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
		console.error("Aggregator 목록 조회에 실패했습니다:", error);
		throw error;
	}
};

// Aggregator 삭제 함수
export const deleteAggregator = async (aggregatorId: string): Promise<void> => {
	try {
		const response = await fetch(`${API_URL}/api/aggregator/${aggregatorId}`, {
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
