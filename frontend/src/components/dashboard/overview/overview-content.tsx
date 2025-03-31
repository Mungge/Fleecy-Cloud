import React from "react";
import StatusCards from "./status-cards";
import TrainingProgressChart from "./training-progress-chart";
import CloudDistributionChart from "./cloud-distribution-chart";
import ActiveModelsList from "./active-models-list";
import ResourceUsageChart from "./resource-usage-chart";
import RecentActivities from "./recent-activities";

// 필요한 데이터 타입
import {
	trainingProgress,
	resourceUsage,
	clusterData,
	activeModels,
	recentActivities,
} from "../mock_data";

const OverviewContent: React.FC = () => {
	return (
		<div className="space-y-4">
			{/* 주요 지표 요약 */}
			<StatusCards />

			{/* 차트와 그래프 */}
			<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<TrainingProgressChart data={trainingProgress} />
				<CloudDistributionChart data={clusterData} />
			</div>

			{/* 활성 모델 목록 */}
			<ActiveModelsList models={activeModels} />

			{/* 자원 사용량 및 최근 활동 */}
			<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<ResourceUsageChart data={resourceUsage} />
				<RecentActivities activities={recentActivities} />
			</div>
		</div>
	);
};

export default OverviewContent;
