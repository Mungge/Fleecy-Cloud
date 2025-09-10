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
		<div className="flex min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
			{/* 왼쪽 배경 섹션 */}
			<div className="w-1/2 bg-gradient-to-br from-sky-200 via-sky-200 to-blue-300 text-gray-700 p-8 flex flex-col relative overflow-hidden">
				{/* 배경 패턴 */}
				<div className="absolute inset-0 opacity-20">
					<div className="absolute top-0 -left-4 w-72 h-72 bg-white rounded-full mix-blend-multiply filter blur-xl animate-pulse"></div>
					<div className="absolute top-20 -right-4 w-72 h-72 bg-sky-100 rounded-full mix-blend-multiply filter blur-xl animate-pulse delay-1000"></div>
					<div className="absolute -bottom-8 left-20 w-72 h-72 bg-blue-200 rounded-full mix-blend-multiply filter blur-xl animate-pulse delay-500"></div>
				</div>

				<div className="relative z-10">
					<div className="flex items-center mt-4 mb-8">
						<div className="w-8 h-8 mr-3 bg-white rounded-lg flex items-center justify-center shadow-sm">
							<svg
								width="20"
								height="20"
								viewBox="0 0 24 24"
								fill="none"
								className="text-sky-500"
							>
								<path
									d="M3 7C3 5.89543 3.89543 5 5 5H19C20.1046 5 21 5.89543 21 7V17C21 18.1046 20.1046 19 19 19H5C3.89543 19 3 18.1046 3 17V7Z"
									stroke="currentColor"
									strokeWidth="2"
									fill="currentColor"
									fillOpacity="0.2"
								/>
								<path
									d="M8 9L16 9M8 13L13 13"
									stroke="currentColor"
									strokeWidth="2"
									strokeLinecap="round"
								/>
							</svg>
						</div>
						<span className="font-bold text-gray-700 text-xl">
							Fleecy Cloud
						</span>
					</div>

					<div className="flex flex-col justify-end h-full mb-16">
						<div className="space-y-6">
							<h2 className="text-3xl font-bold leading-tight text-gray-700">
								클라우드 연합학습의
								<br />
								새로운 경험
							</h2>
							<blockquote className="text-lg text-gray-600 leading-relaxed">
								멀티 클라우드 환경에서 효율적인 리소스 관리와 연합 학습을
								수행해보세요.
							</blockquote>
							<div className="flex items-center space-x-2">
								<div className="w-2 h-2 bg-sky-400 rounded-full"></div>
								<p className="text-gray-600 font-medium">Fleecy Cloud Team</p>
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* 오른쪽 폼 섹션 */}
			<div className="w-1/2 p-8 flex flex-col bg-white/80 backdrop-blur-sm">
				<div className="flex justify-end mb-6">
					<Link href={alternateLink.href}>
						<Button
							variant="ghost"
							className="text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 transition-colors"
						>
							{alternateLink.text}
						</Button>
					</Link>
				</div>

				<div className="flex-1 flex flex-col items-center justify-center max-w-md mx-auto w-full">
					<div className="w-full space-y-8">
						{/* 로고 섹션 */}
						<div className="text-center space-y-6">
							<div className="flex justify-center mb-4">
								<div className="text-center">
									<div className="flex items-center justify-center gap-3 mb-2">
										<div className="relative">
											<svg
												width="56"
												height="56"
												viewBox="0 0 100 100"
												fill="none"
												className="text-sky-300"
											>
												{/* 구름 모양 */}
												<path
													d="M25 60c-8.284 0-15-6.716-15-15 0-8.284 6.716-15 15-15 1.894 0 3.704.351 5.374.992C33.146 23.648 40.222 18 48.5 18c10.77 0 19.5 8.73 19.5 19.5 0 .715-.038 1.42-.113 2.113C70.53 40.25 73 42.95 73 46.25c0 4.004-3.246 7.25-7.25 7.25H25z"
													fill="currentColor"
													opacity="0.8"
												/>
												{/* 구름 하이라이트 */}
												<circle
													cx="35"
													cy="40"
													r="4"
													fill="white"
													opacity="0.6"
												/>
												<circle
													cx="45"
													cy="35"
													r="2.5"
													fill="white"
													opacity="0.4"
												/>
												{/* 구름 그림자 */}
												<ellipse
													cx="48.5"
													cy="65"
													rx="20"
													ry="3"
													fill="currentColor"
													opacity="0.2"
												/>
											</svg>
											{/* 구름 주변 반짝이 효과 */}
											<div className="absolute -top-2 -right-1 w-2.5 h-2.5 bg-sky-200 rounded-full animate-pulse"></div>
											<div className="absolute top-2 -left-3 w-2 h-2 bg-sky-300 rounded-full animate-pulse delay-300"></div>
											<div className="absolute -bottom-1 right-3 w-1.5 h-1.5 bg-sky-200 rounded-full animate-pulse delay-700"></div>
										</div>
										<h1 className="text-4xl font-bold bg-gradient-to-r from-sky-400 via-sky-500 to-blue-500 bg-clip-text text-transparent">
											Fleecy Cloud
										</h1>
									</div>
									<div className="w-32 h-1 bg-gradient-to-r from-sky-300 via-sky-400 to-blue-400 rounded-full mx-auto opacity-80"></div>
								</div>
							</div>

							<div className="space-y-2">
								<h2 className="text-2xl font-bold text-gray-900">{title}</h2>
								<p className="text-gray-600 text-base">{description}</p>
							</div>
						</div>

						{/* 폼 내용 */}
						<div className="space-y-1">{children}</div>
					</div>
				</div>
			</div>
		</div>
	);
}
