"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Github } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthLayout } from "@/components/auth/auth-layout";

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

		if (password !== confirmPassword) {
			setError("비밀번호가 일치하지 않습니다");
			return;
		}

		const passwordErrors = validatePassword(password);
		if (passwordErrors.length > 0) {
			setError(`비밀번호는 ${passwordErrors.join(", ")}해야 합니다`);
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

			router.push("/auth/login");
		} catch (err) {
			setError(err instanceof Error ? err.message : "회원가입에 실패했습니다");
		} finally {
			setIsLoading(false);
		}
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
			<form onSubmit={handleSubmit} className="space-y-4">
				{error && (
					<div className="p-3 text-sm text-red-500 bg-red-50 rounded-md">
						{error}
					</div>
				)}
				<div className="space-y-2">
					<Input
						type="text"
						placeholder="이름"
						value={name}
						onChange={(e) => setName(e.target.value)}
						className="w-full p-2 h-12"
						required
					/>
				</div>
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
						onChange={handlePasswordChange}
						className="w-full p-2 h-12"
						required
					/>
					{password && (
						<div className="space-y-1">
							<div className="h-2 w-full bg-gray-200 rounded-full overflow-hidden">
								<div
									className={`h-full transition-all duration-300 ${getStrengthColor(
										passwordStrength
									)}`}
									style={{ width: `${(passwordStrength / 5) * 100}%` }}
								/>
							</div>
							<p
								className={`text-sm ${
									passwordStrength >= 4 ? "text-green-600" : "text-gray-600"
								}`}
							>
								비밀번호 강도: {getStrengthText(passwordStrength)}
							</p>
						</div>
					)}
					{passwordError && (
						<p className="text-sm text-red-500">{passwordError}</p>
					)}
				</div>
				<div className="space-y-2">
					<Input
						type="password"
						placeholder="비밀번호 확인"
						value={confirmPassword}
						onChange={(e) => setConfirmPassword(e.target.value)}
						className="w-full p-2 h-12"
						required
					/>
				</div>

				<Button
					type="submit"
					className="w-full h-12 bg-black text-white hover:bg-gray-800"
					disabled={isLoading}
				>
					{isLoading ? "가입 중..." : "이메일로 가입하기"}
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
				>
					<Github size={20} />
					<span>GitHub</span>
				</Button>
			</form>
		</AuthLayout>
	);
}
