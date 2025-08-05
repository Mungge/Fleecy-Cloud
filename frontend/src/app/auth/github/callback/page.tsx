"use client";

import { useEffect, useState, useRef, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/contexts/auth-context";
import { toast } from "sonner";

function GitHubCallbackContent() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const { login } = useAuth();
	const [isProcessing, setIsProcessing] = useState(true);
	const [hasError, setHasError] = useState(false);
	const hasProcessed = useRef(false);

	useEffect(() => {
		const handleCallback = async () => {
			if (hasProcessed.current) {
				return;
			}
			hasProcessed.current = true;

			try {
				const token = searchParams.get("token");
				const email = searchParams.get("email");
				const name = searchParams.get("name");
				const idStr = searchParams.get("id");

				if (token && email && idStr) {
					const user = {
						id: parseInt(idStr, 10),
						email,
						name: name || email,
					};
					login(token, user);
					toast.success(`환영합니다, ${user.name}님!`);

					// 성공 시 대시보드로 리다이렉트
					router.push("/dashboard");
					return;
				}

				// URL 파라미터에 토큰이 없는 경우, 기존 방식으로 처리
				const code = searchParams.get("code");
				const error = searchParams.get("error");
				const state = searchParams.get("state");

				if (error) {
					throw new Error(`GitHub OAuth 오류: ${error}`);
				}

				if (!code) {
					throw new Error("인증 코드가 없습니다");
				}

				const apiUrl =
					process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

				// 백엔드의 GitHub 콜백 엔드포인트에 요청
				const callbackURL = `${apiUrl}/api/auth/github/callback`;
				const params = new URLSearchParams({
					code,
					...(state && { state }),
				});

				const response = await fetch(`${callbackURL}?${params}`, {
					method: "GET",
					credentials: "include", // 쿠키 포함 (refresh token을 위해)
					headers: {
						Accept: "application/json",
					},
				});

				if (!response.ok) {
					const errorText = await response.text();
					throw new Error(
						`GitHub 로그인 처리 실패 (${response.status}): ${errorText}`
					);
				}

				const data = await response.json();

				// 백엔드에서 반환한 토큰과 사용자 정보로 로그인 처리
				if (data.access_token && data.user) {
					login(data.access_token, data.user);
					toast.success(`환영합니다, ${data.user.name || data.user.email}님!`);

					// 성공 시 대시보드로 리다이렉트
					router.push("/dashboard");
				} else {
					throw new Error("로그인 정보가 완전하지 않습니다");
				}
			} catch (error) {
				const errorMessage =
					error instanceof Error
						? error.message
						: "GitHub 로그인 중 오류가 발생했습니다";
				toast.error(errorMessage);
				setHasError(true);

				// 3초 후 로그인 페이지로 리다이렉트
				setTimeout(() => {
					router.push("/auth/login");
				}, 2000);
			} finally {
				setIsProcessing(false);
			}
		};

		// 약간의 지연을 주어 Next.js가 완전히 로드된 후 처리
		const timer = setTimeout(() => {
			// searchParams가 로드되었는지 확인
			if (searchParams !== null) {
				// URL에 code나 token 파라미터가 있는지 확인
				const hasAuthParams =
					searchParams.get("code") || searchParams.get("token");

				if (hasAuthParams) {
					handleCallback();
				} else {
					// 인증 파라미터가 없으면 로그인 페이지로 리다이렉트
					setIsProcessing(false);
					setHasError(true);
					setTimeout(() => {
						router.push("/auth/login");
					}, 1000);
				}
			}
		}, 100); // 100ms 지연

		return () => clearTimeout(timer);
	}, [searchParams, router, login]);

	if (isProcessing) {
		return (
			<div className="flex items-center justify-center min-h-screen bg-gray-50">
				<div className="text-center">
					<div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600 mx-auto"></div>
					<p className="mt-4 text-lg text-gray-700">GitHub 로그인 처리 중...</p>
					<p className="mt-2 text-sm text-gray-500">잠시만 기다려주세요</p>
				</div>
			</div>
		);
	}

	// 에러가 발생한 경우에만 에러 페이지 표시
	if (hasError) {
		return (
			<div className="flex items-center justify-center min-h-screen bg-gray-50">
				<div className="text-center">
					<div className="text-red-500 text-6xl mb-4">⚠️</div>
					<h1 className="text-xl font-semibold text-gray-900 mb-2">
						로그인 처리 중 오류가 발생했습니다
					</h1>
					<p className="text-gray-600 mb-4">곧 로그인 페이지로 이동합니다...</p>
				</div>
			</div>
		);
	}

	// 성공적으로 처리된 경우 (리다이렉트 대기 중)
	return (
		<div className="flex items-center justify-center min-h-screen bg-gray-50">
			<div className="text-center">
				<div className="text-green-500 text-6xl mb-4">✅</div>
				<h1 className="text-xl font-semibold text-gray-900 mb-2">
					로그인 완료
				</h1>
				<p className="text-gray-600 mb-4">대시보드로 이동 중...</p>
			</div>
		</div>
	);
}

// 로딩 컴포넌트
function LoadingFallback() {
	return (
		<div className="flex items-center justify-center min-h-screen bg-gray-50">
			<div className="text-center">
				<div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600 mx-auto"></div>
				<p className="mt-4 text-lg text-gray-700">GitHub 로그인 초기화 중...</p>
				<p className="mt-2 text-sm text-gray-500">잠시만 기다려주세요</p>
			</div>
		</div>
	);
}

// 메인 페이지 컴포넌트 - Suspense로 감싸서 useSearchParams 사용 문제 해결
export default function GitHubCallbackPage() {
	return (
		<Suspense fallback={<LoadingFallback />}>
			<GitHubCallbackContent />
		</Suspense>
	);
}
