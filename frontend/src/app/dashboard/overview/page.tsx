"use client";

import dynamic from "next/dynamic";

const OverviewContent = dynamic(
	() => import("@/components/dashboard/overview/overview-content"),
	{ ssr: false }
);

export default function OverviewPage() {
	return <OverviewContent />;
}
