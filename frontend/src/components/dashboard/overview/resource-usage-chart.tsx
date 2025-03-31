import React from "react";
import {
	BarChart,
	Bar,
	XAxis,
	YAxis,
	CartesianGrid,
	Tooltip,
	Legend,
	ResponsiveContainer,
} from "recharts";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

interface ResourceData {
	name: string;
	cpu: number;
	memory: number;
	network: number;
}

interface ResourceUsageChartProps {
	data: ResourceData[];
}

const ResourceUsageChart: React.FC<ResourceUsageChartProps> = ({ data }) => {
	return (
		<Card className="col-span-1">
			<CardHeader>
				<CardTitle>자원 사용량</CardTitle>
				<CardDescription>
					클러스터별 CPU, 메모리, 네트워크 사용량
				</CardDescription>
			</CardHeader>
			<CardContent>
				<ResponsiveContainer width="100%" height={300}>
					<BarChart
						data={data}
						margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
					>
						<CartesianGrid strokeDasharray="3 3" />
						<XAxis dataKey="name" />
						<YAxis />
						<Tooltip />
						<Legend />
						<Bar dataKey="cpu" fill="#3B82F6" name="CPU (%)" />
						<Bar dataKey="memory" fill="#10B981" name="메모리 (%)" />
						<Bar dataKey="network" fill="#F59E0B" name="네트워크 (%)" />
					</BarChart>
				</ResponsiveContainer>
			</CardContent>
		</Card>
	);
};

export default ResourceUsageChart;
