import path from "path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	webpack: (config) => {
		config.resolve.alias = {
			...(config.resolve.alias || {}),
			"@": path.resolve(__dirname, "src"),
		};
		return config;
	},
	// 에러 페이지 비활성화
	onDemandEntries: {
		// 개발 서버가 페이지를 메모리에 유지하는 시간(ms)
		maxInactiveAge: 60 * 60 * 1000,
		// 개발 서버가 한 번에 유지할 페이지 수
		pagesBufferLength: 5,
	},
	// 디버깅을 위한 소스맵 사용
	productionBrowserSourceMaps: true,
	// favicon 및 정적 파일 처리
	images: {
		domains: ["localhost"],
	},
	// CORS 헤더 설정
	async headers() {
		return [
			{
				source: "/(.*)",
				headers: [
					{
						key: "Access-Control-Allow-Origin",
						value: "*",
					},
				],
			},
		];
	},
};

export default nextConfig;
