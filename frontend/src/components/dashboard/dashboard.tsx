"use client";

import React, { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

import OverviewContent from "./overview/overview-content";
import AggregatorContent from "./aggregator/aggregator-content";

const Dashboard = () => {
	const [activeTab, setActiveTab] = useState("overview");

	return (
		<Tabs
			defaultValue="overview"
			className="space-y-4"
			value={activeTab}
			onValueChange={setActiveTab}
		>
			<TabsList>
				<TabsTrigger value="overview">Overview</TabsTrigger>
				<TabsTrigger value="aggregator">Aggregator</TabsTrigger>
				<TabsTrigger value="models">Models</TabsTrigger>
				<TabsTrigger value="resources">Resources</TabsTrigger>
				<TabsTrigger value="activity">Activity</TabsTrigger>
			</TabsList>

			<TabsContent value="overview">
				<OverviewContent />
			</TabsContent>

			<TabsContent value="aggregator">
				<AggregatorContent />
			</TabsContent>

			<TabsContent value="models">
				<Card>
					<CardHeader>
						<CardTitle>모델 관리</CardTitle>
						<CardDescription>연합학습 모델의 관리 및 배포</CardDescription>
					</CardHeader>
					<CardContent>
						<p>모델 관리 탭 콘텐츠가 여기에 표시됩니다.</p>
					</CardContent>
				</Card>
			</TabsContent>

			<TabsContent value="resources">
				<Card>
					<CardHeader>
						<CardTitle>자원 관리</CardTitle>
						<CardDescription>클라우드 자원의 할당 및 모니터링</CardDescription>
					</CardHeader>
					<CardContent>
						<p>자원 관리 탭 콘텐츠가 여기에 표시됩니다.</p>
					</CardContent>
				</Card>
			</TabsContent>

			<TabsContent value="activity">
				<Card>
					<CardHeader>
						<CardTitle>활동 로그</CardTitle>
						<CardDescription>시스템 및 사용자 활동 기록</CardDescription>
					</CardHeader>
					<CardContent>
						<p>활동 로그 탭 콘텐츠가 여기에 표시됩니다.</p>
					</CardContent>
				</Card>
			</TabsContent>
		</Tabs>
	);
};

export default Dashboard;
