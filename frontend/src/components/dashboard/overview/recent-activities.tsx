import React from "react";
import { Activity } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

interface ActivityItem {
	id: number;
	action: string;
	user: string;
	time: string;
	details: string;
}

interface RecentActivitiesProps {
	activities: ActivityItem[];
}

const RecentActivities: React.FC<RecentActivitiesProps> = ({ activities }) => {
	return (
		<Card className="col-span-1">
			<CardHeader>
				<CardTitle>최근 활동</CardTitle>
				<CardDescription>플랫폼의 최근 활동 내역</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="space-y-4">
					{activities.map((activity) => (
						<div
							key={activity.id}
							className="flex items-start pb-4 border-b last:border-0 last:pb-0"
						>
							<div className="mr-4 mt-0.5 bg-primary/10 p-2 rounded-full">
								<Activity className="h-4 w-4 text-primary" />
							</div>
							<div className="space-y-1">
								<p className="font-medium">{activity.action}</p>
								<p className="text-sm text-muted-foreground">
									{activity.details}
								</p>
								<p className="text-xs text-muted-foreground">
									{activity.user} • {activity.time}
								</p>
							</div>
						</div>
					))}
				</div>
			</CardContent>
			<CardFooter>
				<Button variant="outline" size="sm" className="w-full">
					모든 활동 보기
				</Button>
			</CardFooter>
		</Card>
	);
};

export default RecentActivities;
