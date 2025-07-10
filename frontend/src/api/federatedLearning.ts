import { FederatedLearningJob } from "@/types/federated-learning";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

export const createFederatedLearning = async (
	formData: FormData
): Promise<FederatedLearningJob> => {
	try {
		const response = await fetch(`${API_URL}/api/federated-learning`, {
			method: "POST",
			credentials: "include",
			body: formData,
		});

		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}

		const data = await response.json();
		return data.data;
	} catch (error) {
		console.error("연합학습 생성에 실패했습니다:", error);
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
