"use client";

import React, { useState } from "react";
import Sidebar from "@/components/dashboard/layout/sidebar";
import Header from "@/components/dashboard/layout/header";

export default function DashboardLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const [isSidebarOpen, setIsSidebarOpen] = useState(true);

	return (
		<div className="flex h-screen bg-background">
			{/* 사이드바 */}
			<Sidebar
				isSidebarOpen={isSidebarOpen}
				setIsSidebarOpen={setIsSidebarOpen}
			/>

			{/* 메인 콘텐츠 */}
			<div className="flex-1 flex flex-col overflow-hidden">
				{/* 헤더 */}
				<Header isSidebarOpen={isSidebarOpen} />

				{/* 페이지 콘텐츠 */}
				<main className="flex-1 overflow-auto p-6">{children}</main>
			</div>
		</div>
	);
}
