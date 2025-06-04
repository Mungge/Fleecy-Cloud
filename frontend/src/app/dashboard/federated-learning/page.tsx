"use client";

import dynamic from "next/dynamic";

const FederatedLearningContent = dynamic(
	() =>
		import(
			"@/components/dashboard/federated-learning/federated-learning-content"
		),
	{
		loading: () => (
			<div className="flex min-h-screen items-center justify-center">
				<p>콘텐츠 로딩 중...</p>
			</div>
		),
	}
);

export default function FederatedLearningPage() {
	return <FederatedLearningContent />;
}
