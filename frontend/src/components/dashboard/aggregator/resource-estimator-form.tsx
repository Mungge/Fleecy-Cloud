"use client";

import { recommendInstances } from "@/lib/api-service";
import React, { useState } from "react";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectTrigger,
	SelectValue,
	SelectContent,
	SelectItem,
} from "@/components/ui/select";
import { ResourceEstimate, Instance } from "./aggregator-content";

interface ResourceEstimatorFormProps {
	onEstimationComplete: (
		result: ResourceEstimate,
		instances: Instance[]
	) => void;
	setIsLoading: (isLoading: boolean) => void;
}

interface FormData {
	max_clients: string | number;
	avg_model_size_mb: string | number;
	flops: string | number;
	upload_freq_min: string | number;
	aggregation_type: string;
}

const ResourceEstimatorForm: React.FC<ResourceEstimatorFormProps> = ({
	onEstimationComplete,
	setIsLoading,
}) => {
	const [formData, setFormData] = useState<FormData>({
		max_clients: "",
		avg_model_size_mb: "",
		flops: "",
		upload_freq_min: "",
		aggregation_type: "",
	});

	const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		setFormData((prev) => ({
			...prev,
			[name]: value,
		}));
	};

	const handleSelectChange = (name: string, value: string) => {
		setFormData((prev) => ({
			...prev,
			[name]: value,
		}));
	};

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);

		const numericFormData = {
			max_clients: Number(formData.max_clients),
			avg_model_size_mb: Number(formData.avg_model_size_mb),
			flops: Number(formData.flops),
			upload_freq_min: Number(formData.upload_freq_min),
			aggregation_type: formData.aggregation_type || "FedAvg", // 값이 없을 경우 기본값 설정
		};

		try {
			// API 호출
			const data = await recommendInstances(numericFormData);
			onEstimationComplete(data.estimate, data.recommendations);
			toast.success("리소스 추정이 완료되었습니다.");
		} catch (error) {
			console.error("Error:", error);
			toast.error("추정 과정에서 오류가 발생했습니다");

			// 개발 환경이나 오류 발생 시 더미 데이터 사용
			if (process.env.NODE_ENV === "development") {
				setTimeout(() => {
					const dummyData = {
						estimate: {
							ram_gb: 3200 / 1024,
							cpu_percent: 65,
							net_mb_per_second: 2.8,
						},
						recommendations: [
							{ name: "t3.large", vcpu: 2, ram_mb: 8192, family: "AWS t3" },
							{ name: "m5.large", vcpu: 2, ram_mb: 8192, family: "AWS m5" },
							{
								name: "e2-standard-2",
								vcpu: 2,
								ram_mb: 8192,
								family: "GCP e2",
							},
							{ name: "e2-medium", vcpu: 2, ram_mb: 4096, family: "GCP e2" },
							{ name: "c5.large", vcpu: 2, ram_mb: 4096, family: "AWS c5" },
						],
					};

					onEstimationComplete(dummyData.estimate, dummyData.recommendations);
					toast.success("리소스 추정이 완료되었습니다 (더미 데이터)");
				}, 1500);
			} else {
				setIsLoading(false);
			}
		}
	};

	return (
		<form onSubmit={handleSubmit} className="space-y-4">
			<div className="space-y-2">
				<Label htmlFor="max_clients">최대 클라이언트 수</Label>
				<Input
					id="max_clients"
					name="max_clients"
					type="number"
					value={formData.max_clients}
					onChange={handleInputChange}
					min="1"
					required
				/>
			</div>

			<div className="space-y-2">
				<Label htmlFor="avg_model_size_mb">평균 모델 크기 (MB)</Label>
				<Input
					id="avg_model_size_mb"
					name="avg_model_size_mb"
					type="number"
					value={formData.avg_model_size_mb}
					onChange={handleInputChange}
					min="1"
					step="0.1"
					required
				/>
			</div>

			<div className="space-y-2">
				<Label htmlFor="flops">연산량 (FLOPs)</Label>
				<Input
					id="flops"
					name="flops"
					type="number"
					value={formData.flops}
					onChange={handleInputChange}
					min="1"
					required
				/>
				<p className="text-xs text-muted-foreground">
					예: 5,000,000,000 (5 GFLOPs)
				</p>
			</div>

			<div className="space-y-2">
				<Label htmlFor="upload_freq_min">업로드 주기 (분)</Label>
				<Input
					id="upload_freq_min"
					name="upload_freq_min"
					type="number"
					value={formData.upload_freq_min}
					onChange={handleInputChange}
					min="1"
					required
				/>
			</div>

			<div className="space-y-2">
				<Label htmlFor="aggregation_type">집계 방식</Label>
				<Select
					value={formData.aggregation_type}
					onValueChange={(value) =>
						handleSelectChange("aggregation_type", value)
					}
				>
					<SelectTrigger id="aggregation_type" className="w-full">
						<SelectValue placeholder="집계 방식 선택" />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="FedAvg">FedAvg</SelectItem>
						<SelectItem value="FedKD">FedKD</SelectItem>
						<SelectItem value="FedProto">FedProto</SelectItem>
					</SelectContent>
				</Select>
			</div>

			<Button type="submit" className="w-full mt-6">
				리소스 추정하기
			</Button>
		</form>
	);
};

export default ResourceEstimatorForm;
