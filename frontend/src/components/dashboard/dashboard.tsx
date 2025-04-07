"use client";

import dynamic from "next/dynamic";
import React, { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

// 분리된 컴포넌트 import
import Sidebar from "./layout/sidebar";
import Header from "./layout/header";

const OverviewContent = dynamic(() => import("./overview/overview-content"), {
	ssr: false,
});
const AggregatorContent = dynamic(
	() => import("./aggregator/aggregator-content"),
	{ ssr: false }
);

const Dashboard = () => {
	const [isSidebarOpen, setIsSidebarOpen] = useState(true);
	const [activeTab, setActiveTab] = useState("overview");

	return (
		<div className="flex h-screen bg-background">
			{/* 사이드바 컴포넌트 */}
			<Sidebar
				isSidebarOpen={isSidebarOpen}
				setIsSidebarOpen={setIsSidebarOpen}
				activeTab={activeTab}
				setActiveTab={setActiveTab}
			/>

			{/* 메인 콘텐츠 */}
			<div className="flex-1 flex flex-col overflow-hidden">
				{/* 헤더 컴포넌트 */}
				<Header isSidebarOpen={isSidebarOpen} />

				{/* 대시보드 콘텐츠 */}
				<main className="flex-1 overflow-auto p-6">
					<Tabs defaultValue="overview" className="space-y-4">
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

						{/* Aggregator 탭 콘텐츠 추가 */}
						<TabsContent value="aggregator">
							<AggregatorContent />
						</TabsContent>

						<TabsContent value="models">
							<Card>
								<CardHeader>
									<CardTitle>모델 관리</CardTitle>
									<CardDescription>
										연합학습 모델의 관리 및 배포
									</CardDescription>
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
									<CardDescription>
										클라우드 자원의 할당 및 모니터링
									</CardDescription>
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
				</main>
			</div>
		</div>
	);
};

export default Dashboard;
