import React, { useState, useEffect } from "react";

interface AggregatorInstance {
	id: string;
	name: string;
	status: string;
	algorithm: string;
	cloudProvider: string;
	participants: number;
	currentRound: number;
	rounds: number;
	accuracy: number;
	federatedLearningName: string;
	metrics: Record<string, any>;
}

interface AggregatorDetailsProps {
	aggregator: AggregatorInstance;
	onBack: () => void;
}

interface TrainingRound {
	round: number;
	accuracy: number;
	loss: number;
	f1_score: number;
	precision: number;
	recall: number;
	duration: number;
	participantsCount: number;
	timestamp: string;
}

interface RealTimeMetrics {
	accuracy: number;
	loss: number;
	currentRound: number;
	status: string;
	f1_score: number;
	precision: number;
	recall: number;
}

// Mock data for demo purposes
const mockAggregator: AggregatorInstance = {
	id: "demo-aggregator-1",
	name: "Demo Federated Learning Aggregator",
	status: "running",
	algorithm: "FedAvg",
	cloudProvider: "AWS",
	participants: 3,
	currentRound: 5,
	rounds: 10,
	accuracy: 85.2,
	federatedLearningName: "CIFAR-10 Classification",
	metrics: {}
};

const mockTrainingHistory: TrainingRound[] = [
	{
		round: 1,
		accuracy: 0.72,
		loss: 0.85,
		f1_score: 0.70,
		precision: 0.71,
		recall: 0.69,
		duration: 125,
		participantsCount: 3,
		timestamp: new Date(Date.now() - 4 * 3600000).toISOString()
	},
	{
		round: 2,
		accuracy: 0.78,
		loss: 0.65,
		f1_score: 0.76,
		precision: 0.77,
		recall: 0.75,
		duration: 118,
		participantsCount: 3,
		timestamp: new Date(Date.now() - 3 * 3600000).toISOString()
	},
	{
		round: 3,
		accuracy: 0.82,
		loss: 0.52,
		f1_score: 0.81,
		precision: 0.83,
		recall: 0.80,
		duration: 122,
		participantsCount: 3,
		timestamp: new Date(Date.now() - 2 * 3600000).toISOString()
	},
	{
		round: 4,
		accuracy: 0.84,
		loss: 0.41,
		f1_score: 0.83,
		precision: 0.85,
		recall: 0.82,
		duration: 115,
		participantsCount: 3,
		timestamp: new Date(Date.now() - 1 * 3600000).toISOString()
	},
	{
		round: 5,
		accuracy: 0.852,
		loss: 0.38,
		f1_score: 0.847,
		precision: 0.861,
		recall: 0.833,
		duration: 120,
		participantsCount: 3,
		timestamp: new Date().toISOString()
	}
];

const AggregatorDetails: React.FC<AggregatorDetailsProps> = (props) => {
	const { aggregator: propAggregator, onBack } = props;
	
	// Use mock data if no aggregator is provided or if it's incomplete
	const aggregator: AggregatorInstance = propAggregator && propAggregator.id ? propAggregator : mockAggregator;
	
	const [isLoading, setIsLoading] = useState<boolean>(false);
	const [trainingHistory, setTrainingHistory] = useState<TrainingRound[]>(mockTrainingHistory);
	const [realTimeMetrics, setRealTimeMetrics] = useState<RealTimeMetrics | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [lastUpdated, setLastUpdated] = useState<string>(new Date().toLocaleString("ko-KR"));

	const getAuthToken = () => {
		// 1. document.cookie는 "key1=value1; key2=value2; ..." 형태의 문자열을 반환합니다.
		const cookies = document.cookie.split(';');
	
		// 2. 모든 쿠키를 순회하며 'token'을 찾습니다.
		for (let i = 0; i < cookies.length; i++) {
			let cookie = cookies[i].trim(); // 각 쿠키의 앞뒤 공백 제거
	
			// 3. 'token='으로 시작하는 쿠키를 찾습니다.
			if (cookie.startsWith('token=')) {
				// 4. '=' 뒷부분의 토큰 값만 잘라서 반환합니다.
				return cookie.substring('token='.length, cookie.length);
			}
		}
	
		// 5. 'token' 쿠키를 찾지 못하면 빈 문자열을 반환합니다.
		return '';
	};

	// API 호출을 위한 공통 함수
	const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
		const token = getAuthToken();
		
		const defaultOptions: RequestInit = {
			headers: {
				'Content-Type': 'application/json',
				'Authorization': `Bearer ${token}`,
			},
		};

		return fetch(url, {
			...defaultOptions,
			...options,
			headers: {
				...defaultOptions.headers,
				...options.headers,
			},
		});
	};
	
	// MLflow에서 학습 히스토리 조회
	const fetchTrainingHistory = async (): Promise<void> => {
		try {
			setError(null);
			const response = await fetchWithAuth(`http://localhost:8080/api/aggregators/${aggregator.id}/training-history`);
			
			if (response.ok) {
				const data: TrainingRound[] = await response.json();
				setTrainingHistory(data && data.length > 0 ? data : mockTrainingHistory);
			} else {
				// API 실패시 mock 데이터 사용
				setTrainingHistory(mockTrainingHistory);
			}
			setLastUpdated(new Date().toLocaleString("ko-KR"));
		} catch (error: unknown) {
			console.error('학습 히스토리 조회 실패:', error);
			setTrainingHistory(mockTrainingHistory);
			setError('API 서버에 연결할 수 없어 데모 데이터를 표시합니다.');
		}
	};

	// 실시간 메트릭 조회
	const fetchRealTimeMetrics = async (): Promise<void> => {
		try {
			const response = await fetchWithAuth(`http://localhost:8080/api/aggregators/${aggregator.id}/realtime-metrics`);
			
			if (response.ok) {
				const data: RealTimeMetrics = await response.json();
				setRealTimeMetrics(data);
			}
		} catch (error: unknown) {
			console.error('실시간 메트릭 조회 실패:', error);
			// 실패해도 기본값 사용하므로 에러 표시 안함
		}
	};

	// 초기 데이터 로딩
	useEffect(() => {
		fetchTrainingHistory();
		fetchRealTimeMetrics();
	}, []);

	// 실시간 메트릭 주기적 업데이트 (30초마다)
	useEffect(() => {
		const interval = setInterval(() => {
			if (aggregator.status === "running") {
				fetchRealTimeMetrics();
			}
		}, 30000);

		return () => clearInterval(interval);
	}, [aggregator.status]);

	const handleRefresh = async (): Promise<void> => {
		setIsLoading(true);
		try {
			await Promise.all([
				fetchTrainingHistory(),
				fetchRealTimeMetrics()
			]);
		} finally {
			setIsLoading(false);
		}
	};

	const getStatusColor = (status: string): string => {
		const statusColors: Record<string, string> = {
			running: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
			completed: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
			error: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
			pending: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
		};
		return statusColors[status] || "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300";
	};

	const getStatusText = (status: string): string => {
		const statusMap: Record<string, string> = {
			running: "실행 중",
			completed: "완료됨",
			error: "오류",
			pending: "대기 중",
		};
		return statusMap[status] || "알 수 없음";
	};

	const formatDate = (dateString: string): string => {
		return new Date(dateString).toLocaleString("ko-KR");
	};

	const formatDuration = (seconds: number): string => {
		const minutes = Math.floor(seconds / 60);
		const remainingSeconds = Math.floor(seconds % 60);
		return `${minutes}분 ${remainingSeconds}초`;
	};

	// 현재 메트릭 (실시간 데이터 우선, 없으면 기본값)
	const currentMetrics: RealTimeMetrics = realTimeMetrics || {
		accuracy: aggregator.accuracy || 85.2,
		currentRound: aggregator.currentRound || 5,
		status: aggregator.status || "running",
		loss: 0.38,
		f1_score: 0.847,
		precision: 0.861,
		recall: 0.833,
	};

	const progressPercentage: number = aggregator.rounds ? (currentMetrics.currentRound / aggregator.rounds) * 100 : 50;

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex items-center justify-between">
				<div className="flex items-center space-x-4">
					<button 
						className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
						onClick={onBack}
						type="button"
					>
						<span className="mr-2">←</span>
						뒤로 가기
					</button>
					<div>
						<h1 className="text-3xl font-bold">{aggregator.name}</h1>
						<div className="flex items-center space-x-2 mt-2">
							<span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(currentMetrics.status)}`}>
								{getStatusText(currentMetrics.status)}
							</span>
							<span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border border-gray-200">
								{aggregator.algorithm}
							</span>
							<span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
								{aggregator.cloudProvider}
							</span>
						</div>
					</div>
				</div>

				{/* Control Buttons */}
				<div className="flex space-x-2">
					<button
						className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
						onClick={handleRefresh}
						disabled={isLoading}
						type="button"
					>
						<span className={`mr-2 ${isLoading ? 'animate-spin' : ''}`}>↻</span>
						새로고침
					</button>
				</div>
			</div>

			{/* Error Alert */}
			{error && (
				<div className="border border-red-200 bg-red-50 rounded-lg">
					<div className="p-6">
						<div className="flex items-center space-x-2 text-red-800">
							<span>⚠️</span>
							<span>{error}</span>
						</div>
					</div>
				</div>
			)}

			{/* Progress Card */}
			<div className="bg-white rounded-lg border border-gray-200 shadow-sm">
				<div className="p-6">
					<div className="flex items-center justify-between mb-4">
						<div className="flex items-center space-x-2">
							<span>📊</span>
							<span className="text-lg font-semibold">학습 진행 상황</span>
						</div>
						{lastUpdated && (
							<span className="text-sm text-gray-500">
								마지막 업데이트: {lastUpdated}
							</span>
						)}
					</div>
					<div className="space-y-4">
						<div className="flex items-center justify-between">
							<span className="text-sm font-medium">진행률</span>
							<span className="text-sm text-gray-500">
								{currentMetrics.currentRound} / {aggregator.rounds} 라운드
							</span>
						</div>
						<div className="w-full bg-gray-200 rounded-full h-3">
							<div 
								className="bg-blue-600 h-3 rounded-full transition-all duration-300" 
								style={{ width: `${progressPercentage}%` }}
							></div>
						</div>
						<div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
							<div>
								<span className="font-medium">참여자:</span>
								<div className="text-lg font-bold">{aggregator.participants}</div>
							</div>
							<div>
								<span className="font-medium">현재 정확도:</span>
								<div className="text-lg font-bold">
									{currentMetrics.accuracy.toFixed(2)}%
								</div>
							</div>
							<div>
								<span className="font-medium">현재 손실:</span>
								<div className="text-lg font-bold">
									{currentMetrics.loss.toFixed(4)}
								</div>
							</div>
							<div>
								<span className="font-medium">F1 Score:</span>
								<div className="text-lg font-bold">
									{currentMetrics.f1_score.toFixed(3)}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* Training History */}
			<div className="bg-white rounded-lg border border-gray-200 shadow-sm">
				<div className="p-6">
					<div className="flex items-center space-x-2 mb-4">
						<span>🕒</span>
						<span className="text-lg font-semibold">학습 히스토리</span>
					</div>
					<p className="text-sm text-gray-600 mb-4">각 라운드별 학습 결과 및 성능 지표</p>
					
					<div className="space-y-2 max-h-96 overflow-y-auto">
						{trainingHistory.map((round: TrainingRound) => (
							<div
								key={`${round.timestamp}-${round.round}`}
								className="border rounded-lg p-3 hover:bg-gray-50 transition-colors"
							>
								<div className="flex items-center justify-between">
									<div className="flex items-center space-x-4">
										<span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border border-gray-200">
											Round {round.round}
										</span>
										<div className="text-sm">
											<span className="font-medium">정확도:</span>{" "}
											{(round.accuracy * 100).toFixed(2)}%
										</div>
										<div className="text-sm">
											<span className="font-medium">손실:</span>{" "}
											{round.loss.toFixed(4)}
										</div>
										<div className="text-sm">
											<span className="font-medium">F1:</span>{" "}
											{round.f1_score.toFixed(3)}
										</div>
										<div className="text-sm">
											<span className="font-medium">소요시간:</span>{" "}
											{formatDuration(round.duration)}
										</div>
									</div>
									<div className="text-xs text-gray-500">
										{formatDate(round.timestamp)}
									</div>
								</div>
							</div>
						))}
					</div>
				</div>
			</div>

			{/* Performance Chart Placeholder */}
			<div className="bg-white rounded-lg border border-gray-200 shadow-sm">
				<div className="p-6">
					<div className="flex items-center space-x-2 mb-4">
						<span>📈</span>
						<span className="text-lg font-semibold">성능 지표 변화</span>
					</div>
					<p className="text-sm text-gray-600 mb-4">라운드별 정확도 및 손실 변화</p>
					
					<div className="h-80 w-full bg-gray-50 rounded-lg flex items-center justify-center">
						<div className="text-center text-gray-500">
							<span className="text-4xl mb-2 block">📊</span>
							<p>차트는 recharts 라이브러리가 로드되면 표시됩니다</p>
							<div className="mt-4 text-sm">
								<p>현재 데이터:</p>
								<div className="mt-2 space-y-1">
									{trainingHistory.slice(-3).map((round: TrainingRound) => (
										<div key={`${round.timestamp}-${round.round}`} className="flex justify-center space-x-4">
											<span>Round {round.round}:</span>
											<span>정확도 {(round.accuracy * 100).toFixed(1)}%</span>
											<span>손실 {round.loss.toFixed(3)}</span>
										</div>
									))}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
};

export default AggregatorDetails;