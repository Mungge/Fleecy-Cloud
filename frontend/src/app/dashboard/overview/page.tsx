"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function OverviewPage() {
	const router = useRouter();

	useEffect(() => {
		router.replace("/dashboard");
	}, [router]);

	return null; // 리다이렉트 중이니 아무것도 표시하지 않음
}
