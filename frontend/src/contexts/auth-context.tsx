"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

interface User {
	id: number;
	name: string;
	email: string;
}

interface AuthContextType {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
	login: (token: string, user: User) => void;
	logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [token, setToken] = useState<string | null>(null);
	const [isAuthenticated, setIsAuthenticated] = useState(false);
	const router = useRouter();

	useEffect(() => {
		// 페이지 로드 시 로컬 스토리지에서 인증 정보 복원
		const storedToken = localStorage.getItem("token");
		const storedUser = localStorage.getItem("user");

		if (storedToken && storedUser) {
			setToken(storedToken);
			setUser(JSON.parse(storedUser));
			setIsAuthenticated(true);
		}
	}, []);

	const login = (newToken: string, newUser: User) => {
		localStorage.setItem("token", newToken);
		localStorage.setItem("user", JSON.stringify(newUser));
		setToken(newToken);
		setUser(newUser);
		setIsAuthenticated(true);
	};

	const logout = () => {
		localStorage.removeItem("token");
		localStorage.removeItem("user");
		setToken(null);
		setUser(null);
		setIsAuthenticated(false);
		router.push("/auth/login");
	};

	return (
		<AuthContext.Provider
			value={{ user, token, isAuthenticated, login, logout }}
		>
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
