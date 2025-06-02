"use client";

import {
	createContext,
	useContext,
	useState,
	useEffect,
	ReactNode,
} from "react";
import Cookies from "js-cookie";

import { useRouter } from "next/navigation";

interface User {
	id: number;
	email: string;
	name?: string;
}

interface AuthContextType {
	user: User | null;
	loading: boolean;
	login: (token: string, user: User) => void;
	logout: () => void;
	register: (email: string, password: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);
	const router = useRouter();

	useEffect(() => {
		// 로컬 스토리지에서 사용자 정보 복원
		const storedUser = localStorage.getItem("user");
		if (storedUser) {
			setUser(JSON.parse(storedUser));
		}
		setLoading(false);
	}, []);

	const login = (token: string, userData: User) => {
		setUser(userData);
		localStorage.setItem("user", JSON.stringify(userData));
		// 쿠키도 설정
		Cookies.set("token", token, { expires: 2 });
	};

	const logout = () => {
		setUser(null);
		localStorage.removeItem("user");
		// 쿠키도 삭제
		Cookies.remove("token");
		router.push("/auth/login");
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
