"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Github } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthLayout } from "@/components/auth/auth-layout";
import { useAuth } from "@/contexts/auth-context";
import { toast } from "sonner";

export default function LoginPage() {
	const router = useRouter();
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);
	const { login } = useAuth();

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError("");

		// 폼 검증
		if (!email.trim()) {
			const errorMessage = "이메일을 입력해주세요";
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

		if (!password.trim()) {
			const errorMessage = "비밀번호를 입력해주세요";
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

		setIsLoading(true);

		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			console.log("API URL:", apiUrl);

			const response = await fetch(`${apiUrl}/api/auth/login`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					Accept: "application/json",
				},
				body: JSON.stringify({ email, password }),
				credentials: "include",
			});
			const responseText = await response.text();

			let data;
			try {
				data = JSON.parse(responseText);
			} catch (e) {
				throw new Error(
					`서버 응답이 올바른 JSON 형식이 아닙니다: ${
						(e as Error).message
					}. 응답 내용: ${responseText.substring(0, 100)}`
				);
			}
			if (!response.ok) {
				throw new Error(data.error || "로그인에 실패했습니다");
			}
			login(data.access_token, data.user);

			console.log("로그인 성공, 리다이렉트 시도...");
			router.push("/dashboard");
		} catch (err) {
			console.error("로그인 오류:", err);

			const errorMessage =
				err instanceof Error ? err.message : "로그인에 실패했습니다";
			setError(errorMessage);
			toast.error(errorMessage);
		} finally {
			setIsLoading(false);
		}
	};

	const handleGitHubLogin = () => {
		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

			// 환경 변수 확인
			if (!apiUrl) {
				toast.error("API 서버 URL이 설정되지 않았습니다");
				return;
			}

			toast.loading("GitHub으로 로그인 중...");

			// 백엔드의 GitHub OAuth 엔드포인트로 리다이렉트
			// 백엔드에서 GitHub OAuth URL 생성 및 리다이렉트 처리
			window.location.href = `${apiUrl}/api/auth/github`;
		} catch (error) {
			console.error("GitHub 로그인 오류:", error);
			toast.error("GitHub 로그인 중 오류가 발생했습니다");
		}
	};

	return (
		<AuthLayout
			title="로그인"
			description="계정에 로그인하여 서비스를 이용하세요"
			alternateLink={{
				href: "/auth/register",
				text: "회원가입",
			}}
		>
			<form onSubmit={handleSubmit} className="space-y-6">
				{error && (
					<div className="p-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg">
						{error}
					</div>
				)}
				<div className="space-y-4">
					<div className="space-y-2">
						<Input
							type="email"
							placeholder="이메일 주소"
							value={email}
							onChange={(e) => setEmail(e.target.value)}
							className="w-full h-12 px-4 border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
							required
						/>
					</div>
					<div className="space-y-2">
						<Input
							type="password"
							placeholder="비밀번호"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
							className="w-full h-12 px-4 border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
							required
						/>
					</div>
				</div>

				<Button
					type="submit"
					className="w-full h-12 bg-black text-white hover:bg-gray-800 rounded-lg font-medium transition-all duration-200"
					disabled={isLoading}
				>
					{isLoading ? (
						<div className="flex items-center space-x-2">
							<div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
							<span>로그인 중...</span>
						</div>
					) : (
						"이메일로 로그인"
					)}
				</Button>

				<div className="relative my-6">
					<div className="absolute inset-0 flex items-center">
						<span className="w-full border-t border-gray-200" />
					</div>
					<div className="relative flex justify-center text-sm uppercase">
						<span className="bg-white px-4 text-gray-500 font-medium">
							또는 다음으로 계속
						</span>
					</div>
				</div>

				<Button
					type="button"
					variant="outline"
					className="w-full h-12 border-2 border-gray-200 hover:border-gray-300 rounded-lg flex items-center justify-center gap-3 text-gray-700 hover:bg-gray-50 transition-all duration-200"
					onClick={handleGitHubLogin}
				>
					<Github size={20} />
					<span className="font-medium">GitHub으로 계속하기</span>
				</Button>
			</form>
		</AuthLayout>
	);
}
