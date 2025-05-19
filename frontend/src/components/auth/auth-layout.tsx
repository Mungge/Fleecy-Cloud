import { ReactNode } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface AuthLayoutProps {
	children: ReactNode;
	title: string;
	description: string;
	alternateLink: {
		href: string;
		text: string;
	};
}

export function AuthLayout({
	children,
	title,
	description,
	alternateLink,
}: AuthLayoutProps) {
	return (
		<div className="flex min-h-screen">
			{/* 왼쪽 배경 섹션 */}
			<div className="w-1/2 bg-blue-200 text-white p-6 flex flex-col">
				<div className="flex items-center mt-4">
					<div className="mr-2">
						<svg
							width="24"
							height="24"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							strokeWidth="2"
							strokeLinecap="round"
							strokeLinejoin="round"
						>
							<path d="M18 6 7 6 7 18 18 18" />
							<path d="M18 12 7 12" />
						</svg>
					</div>
					<span className="font-bold text-black text-lg">Fleecy Cloud</span>
				</div>

				<div className="flex flex-col justify-end h-full mb-16">
					<blockquote className="text-xl text-black font-medium mb-3">
						&ldquo;멀티 클라우드 연합학습 플랫폼으로 효율적인 리소스 관리와
						학습을 경험하세요.&rdquo;
					</blockquote>
					<p className="text-black font-medium">Fleecy Cloud Team</p>
				</div>
			</div>

			{/* 오른쪽 흰색 배경 섹션 */}
			<div className="w-1/2 p-6 flex flex-col">
				<div className="flex justify-end mb-6">
					<Link href={alternateLink.href}>
						<Button variant="ghost" className="text-sm">
							{alternateLink.text}
						</Button>
					</Link>
				</div>

				<div className="flex-1 flex flex-col items-center justify-center max-w-md mx-auto w-full">
					<div className="w-full space-y-6">
						<div className="text-center space-y-2">
							<h1 className="text-2xl font-bold">{title}</h1>
							<p className="text-gray-500">{description}</p>
						</div>
						{children}
					</div>
				</div>
			</div>
		</div>
	);
}
