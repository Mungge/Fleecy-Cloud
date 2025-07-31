"use client";

import {
	createContext,
	useContext,
	useState,
	useEffect,
	ReactNode,
} from "react";
import Cookies from "js-cookie";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { getUserFromToken, isTokenExpired } from "@/lib/jwt";

interface User {
	id: number;
	email: string;
	name?: string;
}

interface AuthContextType {
	user: User | null;
	loading: boolean;
	login: (token: string, user: User) => void;
	logout: () => Promise<void>;
	register: (email: string, password: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);
	const router = useRouter();

	useEffect(() => {
		// 쿠키에서 토큰을 확인하고 사용자 정보 복원
		const token = Cookies.get("token");
		if (token && !isTokenExpired(token)) {
			const userData = getUserFromToken(token);
			if (userData) {
				setUser(userData);
			}
		} else if (token) {
			// 토큰이 만료된 경우 삭제
			Cookies.remove("token");
		}
		setLoading(false);
	}, []);

	const login = (token: string, userData: User) => {
		setUser(userData);
		// 토큰만 쿠키에 저장 (2시간 후 만료)
		Cookies.set("token", token, { expires: 2 / 24 }); // 2시간을 일 단위로 변환
		toast.success(`환영합니다, ${userData.name || userData.email}님!`);
	};

	const logout = async () => {
		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			const token = Cookies.get("token");

			// 백엔드 logout API 호출
			if (token) {
				await fetch(`${apiUrl}/api/auth/logout`, {
					method: "POST",
					headers: {
						Authorization: `Bearer ${token}`,
						"Content-Type": "application/json",
					},
					credentials: "include", // 쿠키 포함
				});
			}
		} catch (error) {
			console.error("로그아웃 요청 실패:", error);
			// 네트워크 오류가 발생해도 로컬 상태는 정리
		} finally {
			// 로컬 상태 정리
			setUser(null);
			Cookies.remove("token");
			toast.success("로그아웃되었습니다");
			router.push("/auth/login");
		}
	};

	const register = async (email: string, password: string) => {
		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			const response = await fetch(`${apiUrl}/api/auth/register`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ email, password }),
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "회원가입 실패");
			}
		} catch (error) {
			throw error;
		}
	};

	return (
		<AuthContext.Provider value={{ user, loading, login, logout, register }}>
			{children}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (context === undefined) {
		throw new Error("useAuth must be used within an AuthProvider");
	}
	return context;
}
