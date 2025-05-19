// src/components/dashboard/aggregator/aggregator-content.tsx
"use client";

import React, { useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import ResourceEstimatorForm from "./resource-estimator-form";
import ResourceStats from "./resource-stats";
import InstanceRecommendations from "./instance-recommendations";

export interface ResourceEstimate {
	ram_gb: number;
	cpu_percent: number;
	net_mb_per_second: number;
}

export interface Instance {
	name: string;
	vcpu: number;
	ram_mb: number;
	family: string;
}

const AggregatorContent: React.FC = () => {
	const [estimationResult, setEstimationResult] =
		useState<ResourceEstimate | null>(null);
	const [recommendations, setRecommendations] = useState<Instance[]>([]);
	const [isLoading, setIsLoading] = useState(false);

	const handleEstimationComplete = (
		result: ResourceEstimate,
		instances: Instance[]
	) => {
		setEstimationResult(result);
		setRecommendations(instances);
		setIsLoading(false);
	};

	return (
		<div className="space-y-4">
			<div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
				{/* 리소스 추정 폼 */}
				<div className="lg:col-span-1">
					<Card>
						<CardHeader>
							<CardTitle>연합학습 리소스 추정</CardTitle>
							<CardDescription>
								연합학습 파라미터를 입력하여 필요한 리소스를 계산합니다
							</CardDescription>
						</CardHeader>
						<CardContent>
							<ResourceEstimatorForm
								onEstimationComplete={handleEstimationComplete}
								setIsLoading={setIsLoading}
							/>
						</CardContent>
					</Card>
				</div>

				<div className="lg:col-span-2 space-y-6">
					{/* 리소스 추정 결과 */}
					<Card>
						<CardHeader>
							<CardTitle>리소스 추정 결과</CardTitle>
							<CardDescription>
								연합학습에 필요한 컴퓨팅 자원 요구사항
							</CardDescription>
						</CardHeader>
						<CardContent>
							{isLoading ? (
								<div className="flex justify-center items-center py-12">
									<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
								</div>
							) : estimationResult ? (
								<ResourceStats estimationResult={estimationResult} />
							) : (
								<div className="text-center py-8 text-muted-foreground">
									<p>
										왼쪽 폼에서 연합학습 설정을 입력하면 리소스 요구사항이
										이곳에 표시됩니다.
									</p>
								</div>
							)}
						</CardContent>
					</Card>

					{/* 추천 인스턴스 */}
					<Card>
						<CardHeader>
							<CardTitle>추천 인스턴스</CardTitle>
							<CardDescription>
								요구사항에 적합한 클라우드 인스턴스 추천
							</CardDescription>
						</CardHeader>
						<CardContent>
							{isLoading ? (
								<div className="flex justify-center items-center py-12">
									<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
								</div>
							) : recommendations && recommendations.length > 0 ? (
								<InstanceRecommendations recommendations={recommendations} />
							) : (
								<div className="text-center py-8 text-muted-foreground">
									<p>
										왼쪽 폼에서 연합학습 설정을 입력하면 추천 인스턴스가 이곳에
										표시됩니다.
									</p>
								</div>
							)}
						</CardContent>
					</Card>
				</div>
			</div>
		</div>
	);
};

export default AggregatorContent;
