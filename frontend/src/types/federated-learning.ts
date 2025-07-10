// 연합학습 작업 관련 타입들

export interface FederatedLearningJob {
	id: string;
	user_id: number;
	name: string;
	description?: string;
	status: string;
	participants: number;
	completed_at: string | null;
	accuracy: number;
	recall: number;
	precision: number;
	f1score: number;
	rounds: number;
	algorithm: string;
	model_type: string;
	created_at: string;
	updated_at: string;
}

export interface CloudParticipant {
	id: string;
	name: string;
	region: string;
	status: "active" | "inactive";
}
