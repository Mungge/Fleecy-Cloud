"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Instance } from "./aggregator-content";
import { Badge } from "@/components/ui/badge";

interface InstanceRecommendationsProps {
	recommendations: Instance[];
}

const InstanceRecommendations: React.FC<InstanceRecommendationsProps> = ({
	recommendations,
}) => {
	const [selectedProvider, setSelectedProvider] = useState<string>("all");

	// 클라우드 제공업체 필터링 옵션 추출
	const providers = [
		"all",
		...new Set(
			recommendations.map((instance) => instance.family.split(" ")[0])
		),
	];

	// 선택된 제공업체에 따라 필터링
	const filteredInstances =
		selectedProvider === "all"
			? recommendations
			: recommendations.filter((instance) =>
					instance.family.startsWith(selectedProvider)
			  );

	// 인스턴스를 RAM 크기로 정렬
	const sortedInstances = [...filteredInstances].sort(
		(a, b) => a.ram_mb - b.ram_mb
	);

	// 클라우드 제공업체별 색상 지정
	const getProviderStyles = (family: string) => {
		if (family.startsWith("AWS")) {
			return {
				borderColor: "border-orange-500",
				bgColor: "bg-orange-50",
			};
		}
		if (family.startsWith("GCP")) {
			return {
				borderColor: "border-blue-500",
				bgColor: "bg-blue-50",
			};
		}
		if (family.startsWith("Azure")) {
			return {
				borderColor: "border-indigo-500",
				bgColor: "bg-indigo-50",
			};
		}
		return {
			borderColor: "border-gray-500",
			bgColor: "bg-gray-50",
		};
	};

	return (
		<div>
			<div className="flex flex-wrap gap-2 mb-4">
				{providers.map((provider) => (
					<Button
						key={provider}
						variant={selectedProvider === provider ? "secondary" : "outline"}
						size="sm"
						onClick={() => setSelectedProvider(provider)}
						className="rounded-full text-xs"
					>
						{provider === "all" ? "모든 제공업체" : provider}
					</Button>
				))}
			</div>

			{sortedInstances.length > 0 ? (
				<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
					{sortedInstances.map((instance, index) => {
						const styles = getProviderStyles(instance.family);

						return (
							<div
								key={index}
								className={`border-l-4 ${styles.borderColor} ${styles.bgColor} bg-opacity-30 p-4 rounded-md shadow-sm`}
							>
								<div className="flex justify-between items-start mb-2">
									<div>
										<div className="font-medium text-lg">{instance.name}</div>
										<Badge variant="outline" className="mt-1">
											{instance.family}
										</Badge>
									</div>
									<Button variant="outline" size="sm" className="text-xs">
										배포하기
									</Button>
								</div>
								<div className="grid grid-cols-2 gap-2 mt-3">
									<div className="text-sm">
										<span className="text-muted-foreground">vCPU: </span>
										<span className="font-medium">{instance.vcpu}</span>
									</div>
									<div className="text-sm">
										<span className="text-muted-foreground">RAM: </span>
										<span className="font-medium">
											{(instance.ram_mb / 1024).toFixed(1)} GB
										</span>
									</div>
								</div>
							</div>
						);
					})}
				</div>
			) : (
				<div className="text-center py-8 text-muted-foreground">
					<p>선택한 필터에 맞는 인스턴스가 없습니다.</p>
				</div>
			)}
		</div>
	);
};

export default InstanceRecommendations;
