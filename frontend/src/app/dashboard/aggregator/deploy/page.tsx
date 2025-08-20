"use client";

import dynamic from "next/dynamic";

const AggregatorDeployPage = dynamic(
	() => import("@/components/dashboard/aggregator/aggregator-deploy"),
	{ ssr: false }
);

export default function AggregatorPage() {
	return <AggregatorDeployPage />;
}
