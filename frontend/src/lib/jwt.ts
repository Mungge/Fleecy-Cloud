// JWT 토큰 관련 유틸리티 함수들

export interface JWTPayload {
	user_id: number;
	email: string;
	name: string;
	type: string;
	exp: number;
	iat: number;
}

export interface User {
	id: number;
	email: string;
	name?: string;
}

/**
 * JWT 토큰을 디코딩합니다
 * 클라이언트 사이드에서 페이로드만 읽기 위한 용도
 */
export function decodeJWT(token: string): JWTPayload | null {
	try {
		const parts = token.split(".");
		if (parts.length !== 3) {
			return null;
		}

		const payload = parts[1];
		const decoded = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
		return JSON.parse(decoded) as JWTPayload;
	} catch (error) {
		console.error("JWT decode error:", error);
		return null;
	}
}

/**
 * JWT 토큰이 만료되었는지 확인합니다
 */
export function isTokenExpired(token: string): boolean {
	const payload = decodeJWT(token);
	if (!payload) return true;

	const currentTime = Math.floor(Date.now() / 1000);
	return payload.exp < currentTime;
}

/**
 * JWT 토큰에서 사용자 정보를 추출합니다
 */
export function getUserFromToken(token: string): User | null {
	const payload = decodeJWT(token);
	if (!payload) return null;

	return {
		id: payload.user_id,
		email: payload.email,
		name: payload.name,
	};
}
