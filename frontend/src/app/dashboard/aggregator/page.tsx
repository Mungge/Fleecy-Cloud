"use client";

import dynamic from "next/dynamic";

const AggregatorContent = dynamic(
	() => import("@/components/dashboard/aggregator/aggregator-content"),
	{ ssr: false }
);

export default function AggregatorPage() {
	return <AggregatorContent />;
}
