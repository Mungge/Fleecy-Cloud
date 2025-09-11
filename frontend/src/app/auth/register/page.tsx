"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Github } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthLayout } from "@/components/auth/auth-layout";
import { toast } from "sonner";

export default function RegisterPage() {
	const router = useRouter();
	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [confirmPassword, setConfirmPassword] = useState("");
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);
	const [passwordError, setPasswordError] = useState("");
	const [passwordStrength, setPasswordStrength] = useState(0);

	const validatePassword = (password: string) => {
		const hasMinLen = password.length >= 8;
		const hasUpper = /[A-Z]/.test(password);
		const hasLower = /[a-z]/.test(password);
		const hasNumber = /[0-9]/.test(password);
		const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(password);

		const errors = [];
		if (!hasMinLen) errors.push("8자 이상");
		if (!hasUpper) errors.push("대문자 포함");
		if (!hasLower) errors.push("소문자 포함");
		if (!hasNumber) errors.push("숫자 포함");
		if (!hasSpecial) errors.push("특수문자 포함");

		return errors;
	};

	const calculatePasswordStrength = (password: string) => {
		let strength = 0;

		// 길이 체크
		if (password.length >= 8) strength += 1;

		// 문자 종류 체크
		if (/[A-Z]/.test(password)) strength += 1;
		if (/[a-z]/.test(password)) strength += 1;
		if (/[0-9]/.test(password)) strength += 1;
		if (/[^A-Za-z0-9]/.test(password)) strength += 1;

		return strength;
	};

	const getStrengthColor = (strength: number) => {
		switch (strength) {
			case 0:
				return "bg-gray-200";
			case 1:
				return "bg-red-500";
			case 2:
				return "bg-orange-500";
			case 3:
				return "bg-yellow-500";
			case 4:
				return "bg-green-500";
			case 5:
				return "bg-emerald-500";
			default:
				return "bg-gray-200";
		}
	};

	const getStrengthText = (strength: number) => {
		switch (strength) {
			case 0:
				return "매우 약함";
			case 1:
				return "약함";
			case 2:
				return "보통";
			case 3:
				return "강함";
			case 4:
				return "매우 강함";
			case 5:
				return "최강";
			default:
				return "";
		}
	};

	const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const newPassword = e.target.value;
		setPassword(newPassword);
		const errors = validatePassword(newPassword);
		setPasswordError(
			errors.length > 0 ? `비밀번호는 ${errors.join(", ")}해야 합니다` : ""
		);
		setPasswordStrength(calculatePasswordStrength(newPassword));
	};

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError("");

		// 폼 검증
		if (!name.trim()) {
			const errorMessage = "이름을 입력해주세요";
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

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

		if (!confirmPassword.trim()) {
			const errorMessage = "비밀번호 확인을 입력해주세요";
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

		if (password !== confirmPassword) {
			const errorMessage = "비밀번호가 일치하지 않습니다";
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

		const passwordErrors = validatePassword(password);
		if (passwordErrors.length > 0) {
			const errorMessage = `비밀번호는 ${passwordErrors.join(", ")}해야 합니다`;
			setError(errorMessage);
			toast.error(errorMessage);
			return;
		}

		setIsLoading(true);

		try {
			const requestData = { name, email, password };
			const sanitizedRequestData = { name, email }; // Exclude password from logs
			console.log("Request data:", sanitizedRequestData);
			const response = await fetch(
				`${process.env.NEXT_PUBLIC_API_URL}/api/auth/register`,
				{
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(requestData),
					credentials: "include",
				}
			);
			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "회원가입에 실패했습니다");
			}

			toast.success("회원가입이 완료되었습니다! 로그인 페이지로 이동합니다.");
			router.push("/auth/login");
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "회원가입에 실패했습니다";
			setError(errorMessage);
			toast.error(errorMessage);
		} finally {
			setIsLoading(false);
		}
	};

	const handleGitHubLogin = () => {
		const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
		toast.loading("GitHub으로 로그인 중...");
		window.location.href = `${apiUrl}/api/auth/github`;
	};

	return (
		<AuthLayout
			title="회원가입"
			description="새로운 계정을 만들어 서비스를 이용하세요"
			alternateLink={{
				href: "/auth/login",
				text: "로그인",
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
							type="text"
							placeholder="이름"
							value={name}
							onChange={(e) => setName(e.target.value)}
							className="w-full h-12 px-4 border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
							required
						/>
					</div>
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
							onChange={handlePasswordChange}
							className="w-full h-12 px-4 border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
							required
						/>
						{password && (
							<div className="space-y-2">
								<div className="h-2 w-full bg-gray-200 rounded-full overflow-hidden">
									<div
										className={`h-full transition-all duration-300 ${getStrengthColor(
											passwordStrength
										)}`}
										style={{ width: `${(passwordStrength / 5) * 100}%` }}
									/>
								</div>
								<p
									className={`text-sm font-medium ${
										passwordStrength >= 4 ? "text-green-600" : "text-gray-600"
									}`}
								>
									비밀번호 강도: {getStrengthText(passwordStrength)}
								</p>
							</div>
						)}
						{passwordError && (
							<p className="text-sm text-red-500 font-medium">
								{passwordError}
							</p>
						)}
					</div>
					<div className="space-y-2">
						<Input
							type="password"
							placeholder="비밀번호 확인"
							value={confirmPassword}
							onChange={(e) => setConfirmPassword(e.target.value)}
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
							<span>가입 중...</span>
						</div>
					) : (
						"이메일로 가입하기"
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
