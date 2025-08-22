"use client";

import dynamic from "next/dynamic";

const FederatedLearningStartContent = dynamic(
	() =>
		import(
			"@/components/dashboard/federated-learning/federated-learning-start-content"
		),
	{
		loading: () => (
			<div className="flex min-h-screen items-center justify-center">
				<p>연합학습 시작 페이지 로딩 중...</p>
			</div>
		),
	}
);

export default function FederatedLearningStartPage() {
	return <FederatedLearningStartContent />;
}
