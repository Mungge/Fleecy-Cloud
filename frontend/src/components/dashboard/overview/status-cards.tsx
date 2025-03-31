import React from "react";
import { Server, Users, Database, Activity, ArrowUpRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const StatusCards: React.FC = () => {
	return (
		<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-medium">활성 클러스터</CardTitle>
					<Server className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<div className="text-2xl font-bold">24</div>
					<p className="text-xs text-muted-foreground flex items-center mt-1">
						<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
						<span className="text-green-500">12%</span> 증가
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-medium">참여 노드</CardTitle>
					<Users className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<div className="text-2xl font-bold">128</div>
					<p className="text-xs text-muted-foreground flex items-center mt-1">
						<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
						<span className="text-green-500">5%</span> 증가
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-medium">활성 모델</CardTitle>
					<Database className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<div className="text-2xl font-bold">4</div>
					<p className="text-xs text-muted-foreground flex items-center mt-1">
						<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
						<span className="text-green-500">2</span> 증가
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader className="flex flex-row items-center justify-between pb-2">
					<CardTitle className="text-sm font-medium">평균 학습 속도</CardTitle>
					<Activity className="h-4 w-4 text-muted-foreground" />
				</CardHeader>
				<CardContent>
					<div className="text-2xl font-bold">1.8x</div>
					<p className="text-xs text-muted-foreground flex items-center mt-1">
						<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
						<span className="text-green-500">0.3x</span> 증가
					</p>
				</CardContent>
			</Card>
		</div>
	);
};

export default StatusCards;
