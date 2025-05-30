import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function redirectMiddleware(request: NextRequest) {
	const token = request.cookies.get("token");
	const isAuthPage = request.nextUrl.pathname.startsWith("/auth");
	const isRootPage = request.nextUrl.pathname === "/";

	// 토큰이 없고 인증 페이지가 아닌 경우 로그인 페이지로 리다이렉트
	if (!token && !isAuthPage) {
		console.log(
			"[Middleware] 인증되지 않은 사용자, 로그인 페이지로 리다이렉트"
		);
		return NextResponse.redirect(new URL("/auth/login", request.url));
	}

	// 토큰이 있고 인증 페이지에 접근하는 경우 대시보드로 리다이렉트
	if (token && isAuthPage) {
		console.log("[Middleware] 인증된 사용자, 대시보드로 리다이렉트");
		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	// 토큰이 있고 루트 페이지에 접근하는 경우 대시보드로 리다이렉트
	if (token && isRootPage) {
		console.log("[Middleware] 인증된 사용자, 대시보드로 리다이렉트");
		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	return NextResponse.next();
}

export const config = {
	matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
