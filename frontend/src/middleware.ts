import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
	// 쿠키와 로컬 스토리지에서 토큰 확인
	const cookieToken = request.cookies.get("token");
	const localStorageToken = request.headers.get("x-auth-token");
	const token = cookieToken?.value || localStorageToken;

	const isAuthPage = request.nextUrl.pathname.startsWith("/auth");

	if (!token && !isAuthPage) {
		// 인증되지 않은 사용자를 로그인 페이지로 리다이렉트
		return NextResponse.redirect(new URL("/auth/login", request.url));
	}

	if (token && isAuthPage) {
		// 이미 로그인한 사용자를 대시보드로 리다이렉트
		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	return NextResponse.next();
}

export const config = {
	matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
