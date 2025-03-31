import React from "react";
import {
	PieChart,
	Pie,
	Cell,
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

interface ClusterData {
	name: string;
	value: number;
	color: string;
}

interface CloudDistributionChartProps {
	data: ClusterData[];
}

const CloudDistributionChart: React.FC<CloudDistributionChartProps> = ({
	data,
}) => {
	return (
		<Card className="col-span-1">
			<CardHeader>
				<CardTitle>클라우드 분포</CardTitle>
				<CardDescription>클라우드 제공업체별 자원 분포</CardDescription>
			</CardHeader>
			<CardContent>
				<ResponsiveContainer width="100%" height={300}>
					<PieChart>
						<Pie
							data={data}
							cx="50%"
							cy="50%"
							outerRadius={100}
							dataKey="value"
							label={({ name, percent }) =>
								`${name} ${(percent * 100).toFixed(0)}%`
							}
						>
							{data.map((entry, index) => (
								<Cell key={`cell-${index}`} fill={entry.color} />
							))}
						</Pie>
						<Tooltip />
						<Legend />
					</PieChart>
				</ResponsiveContainer>
			</CardContent>
		</Card>
	);
};

export default CloudDistributionChart;
