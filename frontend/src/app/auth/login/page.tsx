"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Github } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthLayout } from "@/components/auth/auth-layout";
import { useAuth } from "@/contexts/auth-context";

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

			console.log("Response status:", response.status);
			const responseText = await response.text();
			console.log("Response text:", responseText);

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

			// 중복된 쿠키 설정 제거하고 login 함수만 호출
			login(data.token, data.user);

			console.log("로그인 성공, 리다이렉트 시도...");
			router.push("/dashboard");
		} catch (err) {
			console.error("로그인 오류:", err);
			setError(err instanceof Error ? err.message : "로그인에 실패했습니다");
		} finally {
			setIsLoading(false);
		}
	};

	const handleGitHubLogin = () => {
		const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
		window.location.href = `${apiUrl}/api/auth/github`;
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
			<form onSubmit={handleSubmit} className="space-y-4">
				{error && (
					<div className="p-3 text-sm text-red-500 bg-red-50 rounded-md">
						{error}
					</div>
				)}
				<div className="space-y-2">
					<Input
						type="email"
						placeholder="name@example.com"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						className="w-full p-2 h-12"
						required
					/>
				</div>
				<div className="space-y-2">
					<Input
						type="password"
						placeholder="비밀번호"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						className="w-full p-2 h-12"
						required
					/>
				</div>

				<Button
					type="submit"
					className="w-full h-12 bg-black text-white hover:bg-gray-800"
					disabled={isLoading}
				>
					{isLoading ? "로그인 중..." : "이메일로 로그인"}
				</Button>

				<div className="relative">
					<div className="absolute inset-0 flex items-center">
						<span className="w-full border-t border-gray-200" />
					</div>
					<div className="relative flex justify-center text-xs uppercase">
						<span className="bg-white px-2 text-gray-500">
							또는 다음으로 계속
						</span>
					</div>
				</div>

				<Button
					type="button"
					variant="outline"
					className="w-full h-12 border-gray-300 flex items-center justify-center gap-2"
					onClick={handleGitHubLogin}
				>
					<Github size={20} />
					<span>GitHub</span>
				</Button>
			</form>
		</AuthLayout>
	);
}
