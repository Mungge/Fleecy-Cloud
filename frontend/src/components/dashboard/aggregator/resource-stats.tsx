"use client";

import React from "react";
import {
	BarChart,
	Bar,
	ResponsiveContainer,
	XAxis,
	YAxis,
	Tooltip,
	Legend,
} from "recharts";
import { ResourceEstimate } from "./aggregator-content";

interface ResourceStatsProps {
	estimationResult: ResourceEstimate;
}

const ResourceStats: React.FC<ResourceStatsProps> = ({ estimationResult }) => {
	const chartData = [
		{
			name: "리소스 요구사항",
			RAM: estimationResult.ram_mb,
			CPU: estimationResult.cpu_percent,
			Network: estimationResult.net_mb_per_second * 10, // 시각화를 위해 확대
		},
	];

	// 리소스 상태에 따른 색상 (상태에 따라 색상을 다르게 할 수 있습니다)
	const getStatusColor = (type: string, value: number) => {
		if (type === "RAM") {
			return value > 8192 ? "text-yellow-500" : "text-green-500";
		} else if (type === "CPU") {
			return value > 80 ? "text-yellow-500" : "text-green-500";
		} else {
			return value > 5 ? "text-yellow-500" : "text-green-500";
		}
	};

	return (
		<div className="space-y-6">
			<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
				{/* RAM 사용량 */}
				<div className="bg-card/50 rounded-lg p-4 border shadow-sm">
					<div className="text-xs uppercase text-muted-foreground mb-1">
						필요한 RAM
					</div>
					<div className="flex items-baseline">
						<span
							className={`text-2xl font-bold ${getStatusColor(
								"RAM",
								estimationResult.ram_mb
							)}`}
						>
							{estimationResult.ram_mb.toLocaleString()}
						</span>
						<span className="text-sm ml-1 text-muted-foreground">MB</span>
					</div>
					<div className="mt-1 text-xs text-muted-foreground">
						{estimationResult.ram_mb > 8192
							? "높은 메모리 요구량"
							: "적절한 메모리 요구량"}
					</div>
				</div>

				{/* CPU 사용량 */}
				<div className="bg-card/50 rounded-lg p-4 border shadow-sm">
					<div className="text-xs uppercase text-muted-foreground mb-1">
						CPU 사용률
					</div>
					<div className="flex items-baseline">
						<span
							className={`text-2xl font-bold ${getStatusColor(
								"CPU",
								estimationResult.cpu_percent
							)}`}
						>
							{estimationResult.cpu_percent}
						</span>
						<span className="text-sm ml-1 text-muted-foreground">%</span>
					</div>
					<div className="mt-1 text-xs text-muted-foreground">
						{estimationResult.cpu_percent > 80
							? "높은 CPU 사용률"
							: "적절한 CPU 사용률"}
					</div>
				</div>

				{/* 네트워크 대역폭 */}
				<div className="bg-card/50 rounded-lg p-4 border shadow-sm">
					<div className="text-xs uppercase text-muted-foreground mb-1">
						네트워크 대역폭
					</div>
					<div className="flex items-baseline">
						<span
							className={`text-2xl font-bold ${getStatusColor(
								"Network",
								estimationResult.net_mb_per_second
							)}`}
						>
							{estimationResult.net_mb_per_second}
						</span>
						<span className="text-sm ml-1 text-muted-foreground">MB/s</span>
					</div>
					<div className="mt-1 text-xs text-muted-foreground">
						{estimationResult.net_mb_per_second > 5
							? "높은 네트워크 요구량"
							: "적절한 네트워크 요구량"}
					</div>
				</div>
			</div>

			{/* 차트 */}
			<div className="bg-card/50 rounded-lg p-4 border shadow-sm">
				<div className="font-medium mb-4">리소스 요구사항 비교</div>
				<div className="h-64">
					<ResponsiveContainer width="100%" height="100%">
						<BarChart data={chartData}>
							<XAxis dataKey="name" />
							<YAxis />
							<Tooltip
								formatter={(value, name) => {
									if (name === "RAM") return `${value.toLocaleString()} MB`;
									if (name === "CPU") return `${value}%`;
									if (name === "Network")
										return `${(value as number) / 10} MB/s`;
									return value;
								}}
							/>
							<Legend />
							<Bar dataKey="RAM" fill="var(--chart-1)" name="RAM (MB)" />
							<Bar dataKey="CPU" fill="var(--chart-2)" name="CPU (%)" />
							<Bar
								dataKey="Network"
								fill="var(--chart-3)"
								name="Network (MB/s)"
							/>
						</BarChart>
					</ResponsiveContainer>
				</div>
			</div>
		</div>
	);
};

export default ResourceStats;
