// 샘플 데이터
export const trainingProgress = [
	{ round: 1, accuracy: 0.65, loss: 0.75 },
	{ round: 2, accuracy: 0.72, loss: 0.62 },
	{ round: 3, accuracy: 0.78, loss: 0.51 },
	{ round: 4, accuracy: 0.82, loss: 0.43 },
	{ round: 5, accuracy: 0.85, loss: 0.38 },
	{ round: 6, accuracy: 0.87, loss: 0.34 },
	{ round: 7, accuracy: 0.89, loss: 0.31 },
	{ round: 8, accuracy: 0.91, loss: 0.28 },
];

export const clusterData = [
	{ name: "AWS", value: 45, color: "#FF9900" },
	{ name: "Azure", value: 30, color: "#0078D4" },
	{ name: "GCP", value: 25, color: "#4285F4" },
];

export const resourceUsage = [
	{ name: "클러스터 1", cpu: 78, memory: 65, network: 45 },
	{ name: "클러스터 2", cpu: 45, memory: 72, network: 63 },
	{ name: "클러스터 3", cpu: 82, memory: 51, network: 38 },
	{ name: "클러스터 4", cpu: 56, memory: 48, network: 72 },
];

export const activeModels: {
	id: number;
	name: string;
	status: "학습 중" | "완료됨" | "대기 중" | "오류";
	progress: number;
	participatingNodes: number;
	lastUpdate: string;
}[] = [
	{
		id: 1,
		name: "이미지 분류 모델",
		status: "학습 중",
		progress: 72,
		participatingNodes: 12,
		lastUpdate: "10분 전",
	},
	{
		id: 2,
		name: "자연어 처리 모델",
		status: "완료됨",
		progress: 100,
		participatingNodes: 8,
		lastUpdate: "2시간 전",
	},
	{
		id: 3,
		name: "시계열 예측 모델",
		status: "대기 중",
		progress: 0,
		participatingNodes: 15,
		lastUpdate: "30분 전",
	},
	{
		id: 4,
		name: "추천 시스템",
		status: "학습 중",
		progress: 45,
		participatingNodes: 10,
		lastUpdate: "5분 전",
	},
];

export const recentActivities = [
	{
		id: 1,
		action: "새 모델 배포",
		user: "김관리자",
		time: "10분 전",
		details: "이미지 분류 모델 v2",
	},
	{
		id: 2,
		action: "클러스터 확장",
		user: "이개발자",
		time: "1시간 전",
		details: "AWS 클러스터 5개 노드 추가",
	},
	{
		id: 3,
		action: "학습 완료",
		user: "시스템",
		time: "2시간 전",
		details: "자연어 처리 모델 (정확도: 92%)",
	},
	{
		id: 4,
		action: "데이터 업데이트",
		user: "박데이터",
		time: "3시간 전",
		details: "이미지 데이터셋 추가 (2,500개)",
	},
];
