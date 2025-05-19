import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
	const token = request.cookies.get("token");
	const isAuthPage = request.nextUrl.pathname.startsWith("/auth");
	const isDashboardPage = request.nextUrl.pathname.startsWith("/dashboard");

	// 로그인되지 않은 상태에서 대시보드 접근 시도
	if (!token && isDashboardPage) {
		console.log(
			"[Middleware] 인증되지 않은 사용자, 로그인 페이지로 리다이렉트"
		);
		return NextResponse.redirect(new URL("/auth/login", request.url));
	}

	// 로그인된 상태에서 인증 페이지 접근 시도
	if (token && isAuthPage) {
		console.log("[Middleware] 인증된 사용자, 대시보드로 리다이렉트");

		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	return NextResponse.next();
}

export const config = {
	matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
