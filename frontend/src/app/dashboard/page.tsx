"use client";

import { useEffect } from "react";
import { useAuth } from "@/contexts/auth-context";
import { useRouter } from "next/navigation";
import Dashboard from "@/components/dashboard/dashboard";

export default function DashboardPage() {
	const { user, loading } = useAuth();
	const router = useRouter();

	useEffect(() => {
		// 인증이 완료되었고 사용자가 없으면 로그인 페이지로 리다이렉트
		if (!loading && !user) {
			router.push("/auth/login");
		}
	}, [user, loading, router]);

	if (loading) {
		return (
			<div className="flex min-h-screen items-center justify-center">
				<p>로딩 중...</p>
			</div>
		);
	}

	if (!user) {
		return null; // 리다이렉트 중이니 아무것도 표시하지 않음
	}

	return <Dashboard />;
}
