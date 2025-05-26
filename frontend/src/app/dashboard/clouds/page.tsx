"use client";

import dynamic from "next/dynamic";

const CloudsContent = dynamic(
	() => import("@/components/dashboard/clouds/clouds-content"),
	{
		loading: () => (
			<div className="flex min-h-screen items-center justify-center">
				<p>콘텐츠 로딩 중...</p>
			</div>
		),
	}
);

export default function CloudsPage() {
	return <CloudsContent />;
}
