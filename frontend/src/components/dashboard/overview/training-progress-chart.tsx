import React from "react";
import {
	LineChart,
	Line,
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

interface TrainingProgressData {
	round: number;
	accuracy: number;
	loss: number;
}

interface TrainingProgressChartProps {
	data: TrainingProgressData[];
}

const TrainingProgressChart: React.FC<TrainingProgressChartProps> = ({
	data,
}) => {
	return (
		<Card className="col-span-1">
			<CardHeader>
				<CardTitle>학습 진행 상황</CardTitle>
				<CardDescription>라운드별 정확도와 손실</CardDescription>
			</CardHeader>
			<CardContent>
				<ResponsiveContainer width="100%" height={300}>
					<LineChart
						data={data}
						margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
					>
						<CartesianGrid strokeDasharray="3 3" />
						<XAxis dataKey="round" />
						<YAxis />
						<Tooltip />
						<Legend />
						<Line
							type="monotone"
							dataKey="accuracy"
							stroke="#3B82F6"
							name="정확도"
							strokeWidth={2}
						/>
						<Line
							type="monotone"
							dataKey="loss"
							stroke="#EF4444"
							name="손실"
							strokeWidth={2}
						/>
					</LineChart>
				</ResponsiveContainer>
			</CardContent>
		</Card>
	);
};

export default TrainingProgressChart;
