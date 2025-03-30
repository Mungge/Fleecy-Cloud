"use client";

import React, { useState } from "react";
import {
	LineChart,
	Line,
	BarChart,
	Bar,
	XAxis,
	YAxis,
	CartesianGrid,
	Tooltip,
	Legend,
	ResponsiveContainer,
	PieChart,
	Pie,
	Cell,
} from "recharts";
import {
	Cloud,
	Server,
	Database,
	Activity,
	Users,
	Bell,
	Settings,
	Menu,
	X,
	ChevronDown,
	ArrowUpRight,
	PlusCircle,
	MoreHorizontal,
	Search,
	Calendar,
} from "lucide-react";

// shadcn ui 컴포넌트 import
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";

// 샘플 데이터
const trainingProgress = [
	{ round: 1, accuracy: 0.65, loss: 0.75 },
	{ round: 2, accuracy: 0.72, loss: 0.62 },
	{ round: 3, accuracy: 0.78, loss: 0.51 },
	{ round: 4, accuracy: 0.82, loss: 0.43 },
	{ round: 5, accuracy: 0.85, loss: 0.38 },
	{ round: 6, accuracy: 0.87, loss: 0.34 },
	{ round: 7, accuracy: 0.89, loss: 0.31 },
	{ round: 8, accuracy: 0.91, loss: 0.28 },
];

const clusterData = [
	{ name: "AWS", value: 45, color: "#FF9900" },
	{ name: "Azure", value: 30, color: "#0078D4" },
	{ name: "GCP", value: 25, color: "#4285F4" },
];

const resourceUsage = [
	{ name: "클러스터 1", cpu: 78, memory: 65, network: 45 },
	{ name: "클러스터 2", cpu: 45, memory: 72, network: 63 },
	{ name: "클러스터 3", cpu: 82, memory: 51, network: 38 },
	{ name: "클러스터 4", cpu: 56, memory: 48, network: 72 },
];

const activeModels: {
	id: number;
	name: string;
	status: "학습 중" | "완료됨" | "대기 중" | "오류";
	progress: number;
	participatingNodes: number;
	lastUpdate: string;
}[] = [
	{
		id: 1,
		name: "이미지 분류 모델",
		status: "학습 중",
		progress: 72,
		participatingNodes: 12,
		lastUpdate: "10분 전",
	},
	{
		id: 2,
		name: "자연어 처리 모델",
		status: "완료됨",
		progress: 100,
		participatingNodes: 8,
		lastUpdate: "2시간 전",
	},
	{
		id: 3,
		name: "시계열 예측 모델",
		status: "대기 중",
		progress: 0,
		participatingNodes: 15,
		lastUpdate: "30분 전",
	},
	{
		id: 4,
		name: "추천 시스템",
		status: "학습 중",
		progress: 45,
		participatingNodes: 10,
		lastUpdate: "5분 전",
	},
];

const recentActivities = [
	{
		id: 1,
		action: "새 모델 배포",
		user: "김관리자",
		time: "10분 전",
		details: "이미지 분류 모델 v2",
	},
	{
		id: 2,
		action: "클러스터 확장",
		user: "이개발자",
		time: "1시간 전",
		details: "AWS 클러스터 5개 노드 추가",
	},
	{
		id: 3,
		action: "학습 완료",
		user: "시스템",
		time: "2시간 전",
		details: "자연어 처리 모델 (정확도: 92%)",
	},
	{
		id: 4,
		action: "데이터 업데이트",
		user: "박데이터",
		time: "3시간 전",
		details: "이미지 데이터셋 추가 (2,500개)",
	},
];

// const statusColorMap = {
// 	"학습 중": "bg-blue-500",
// 	완료됨: "bg-green-500",
// 	"대기 중": "bg-yellow-500",
// 	오류: "bg-red-500",
// };

const Dashboard = () => {
	const [isSidebarOpen, setIsSidebarOpen] = useState(true);
	const [activeTab, setActiveTab] = useState("overview");

	const renderStatusBadge = (
		status: "학습 중" | "완료됨" | "대기 중" | "오류"
	) => {
		let variant: "default" | "secondary" | "destructive" | "outline" =
			"default";

		switch (status) {
			case "학습 중":
				variant = "secondary";
				break;
			case "완료됨":
				variant = "outline"; // Map "success" to "outline"
				break;
			case "대기 중":
				variant = "outline"; // Map "warning" to "outline"
				break;
			case "오류":
				variant = "destructive";
				break;
			default:
				variant = "default";
		}

		return <Badge variant={variant}>{status}</Badge>;
	};

	return (
		<div className="flex h-screen bg-background">
			{/* 사이드바 */}
			<div
				className={`${
					isSidebarOpen ? "w-64" : "w-16"
				} bg-card border-r transition-all duration-300 ease-in-out`}
			>
				<div className="p-4 flex justify-between items-center">
					{isSidebarOpen && (
						<h1 className="font-bold text-lg">연합학습 플랫폼</h1>
					)}
					<Button
						variant="ghost"
						size="icon"
						onClick={() => setIsSidebarOpen(!isSidebarOpen)}
					>
						{isSidebarOpen ? <X size={20} /> : <Menu size={20} />}
					</Button>
				</div>
				<ScrollArea className="h-[calc(100vh-64px)]">
					<nav className="mt-2 px-2">
						<Button
							variant={activeTab === "overview" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("overview")}
						>
							<Activity className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>대시보드</span>}
						</Button>
						<Button
							variant={activeTab === "data" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("data")}
						>
							<Database className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>데이터 관리</span>}
						</Button>
						<Button
							variant={activeTab === "models" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("models")}
						>
							<Server className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>모델 관리</span>}
						</Button>
						<Button
							variant={activeTab === "cloud" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("cloud")}
						>
							<Cloud className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>클라우드 관리</span>}
						</Button>
						<Button
							variant={activeTab === "users" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("users")}
						>
							<Users className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>참여자 관리</span>}
						</Button>

						{isSidebarOpen && <Separator className="my-4" />}

						<Button
							variant={activeTab === "settings" ? "secondary" : "ghost"}
							className="w-full justify-start mb-1"
							onClick={() => setActiveTab("settings")}
						>
							<Settings className="h-5 w-5 mr-2" />
							{isSidebarOpen && <span>설정</span>}
						</Button>
					</nav>
				</ScrollArea>
			</div>

			{/* 메인 콘텐츠 */}
			<div className="flex-1 flex flex-col overflow-hidden">
				{/* 상단 네비게이션 */}
				<header className="h-16 border-b bg-card">
					<div className="px-6 h-full flex justify-between items-center">
						<div className="flex gap-4 items-center">
							<h2 className="text-xl font-semibold">대시보드</h2>
							<div className="relative w-64 hidden md:block">
								<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
								<Input placeholder="검색..." className="pl-8" />
							</div>
						</div>
						<div className="flex items-center gap-4">
							<Button
								variant="outline"
								size="sm"
								className="hidden md:flex gap-2"
							>
								<Calendar className="h-4 w-4" />
								<span>2025년 3월</span>
							</Button>
							<DropdownMenu>
								<DropdownMenuTrigger asChild>
									<Button variant="ghost" size="icon">
										<Bell className="h-5 w-5" />
									</Button>
								</DropdownMenuTrigger>
								<DropdownMenuContent align="end" className="w-80">
									<DropdownMenuLabel>알림</DropdownMenuLabel>
									<DropdownMenuSeparator />
									<DropdownMenuItem>
										<div className="flex flex-col gap-1">
											<p className="font-medium">새 모델 배포</p>
											<p className="text-sm text-muted-foreground">
												이미지 분류 모델 v2가 배포되었습니다.
											</p>
											<p className="text-xs text-muted-foreground">10분 전</p>
										</div>
									</DropdownMenuItem>
									<DropdownMenuSeparator />
									<DropdownMenuItem>
										<div className="flex flex-col gap-1">
											<p className="font-medium">학습 완료</p>
											<p className="text-sm text-muted-foreground">
												자연어 처리 모델 학습이 완료되었습니다.
											</p>
											<p className="text-xs text-muted-foreground">2시간 전</p>
										</div>
									</DropdownMenuItem>
								</DropdownMenuContent>
							</DropdownMenu>
							<DropdownMenu>
								<DropdownMenuTrigger asChild>
									<Button variant="ghost" className="gap-2">
										<Avatar className="h-8 w-8">
											<AvatarFallback>관</AvatarFallback>
										</Avatar>
										{isSidebarOpen && <span>관리자</span>}
										<ChevronDown className="h-4 w-4 text-muted-foreground" />
									</Button>
								</DropdownMenuTrigger>
								<DropdownMenuContent align="end">
									<DropdownMenuLabel>내 계정</DropdownMenuLabel>
									<DropdownMenuSeparator />
									<DropdownMenuItem>프로필</DropdownMenuItem>
									<DropdownMenuItem>설정</DropdownMenuItem>
									<DropdownMenuSeparator />
									<DropdownMenuItem>로그아웃</DropdownMenuItem>
								</DropdownMenuContent>
							</DropdownMenu>
						</div>
					</div>
				</header>

				{/* 대시보드 콘텐츠 */}
				<main className="flex-1 overflow-auto p-6">
					<Tabs defaultValue="overview" className="space-y-4">
						<TabsList>
							<TabsTrigger value="overview">개요</TabsTrigger>
							<TabsTrigger value="models">모델</TabsTrigger>
							<TabsTrigger value="resources">자원</TabsTrigger>
							<TabsTrigger value="activity">활동</TabsTrigger>
						</TabsList>

						<TabsContent value="overview" className="space-y-4">
							{/* 주요 지표 요약 */}
							<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
								<Card>
									<CardHeader className="flex flex-row items-center justify-between pb-2">
										<CardTitle className="text-sm font-medium">
											활성 클러스터
										</CardTitle>
										<Server className="h-4 w-4 text-muted-foreground" />
									</CardHeader>
									<CardContent>
										<div className="text-2xl font-bold">24</div>
										<p className="text-xs text-muted-foreground flex items-center mt-1">
											<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
											<span className="text-green-500">12%</span> 증가
										</p>
									</CardContent>
								</Card>

								<Card>
									<CardHeader className="flex flex-row items-center justify-between pb-2">
										<CardTitle className="text-sm font-medium">
											참여 노드
										</CardTitle>
										<Users className="h-4 w-4 text-muted-foreground" />
									</CardHeader>
									<CardContent>
										<div className="text-2xl font-bold">128</div>
										<p className="text-xs text-muted-foreground flex items-center mt-1">
											<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
											<span className="text-green-500">5%</span> 증가
										</p>
									</CardContent>
								</Card>

								<Card>
									<CardHeader className="flex flex-row items-center justify-between pb-2">
										<CardTitle className="text-sm font-medium">
											활성 모델
										</CardTitle>
										<Database className="h-4 w-4 text-muted-foreground" />
									</CardHeader>
									<CardContent>
										<div className="text-2xl font-bold">4</div>
										<p className="text-xs text-muted-foreground flex items-center mt-1">
											<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
											<span className="text-green-500">2</span> 증가
										</p>
									</CardContent>
								</Card>

								<Card>
									<CardHeader className="flex flex-row items-center justify-between pb-2">
										<CardTitle className="text-sm font-medium">
											평균 학습 속도
										</CardTitle>
										<Activity className="h-4 w-4 text-muted-foreground" />
									</CardHeader>
									<CardContent>
										<div className="text-2xl font-bold">1.8x</div>
										<p className="text-xs text-muted-foreground flex items-center mt-1">
											<ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
											<span className="text-green-500">0.3x</span> 증가
										</p>
									</CardContent>
								</Card>
							</div>

							{/* 차트와 그래프 */}
							<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
								<Card className="col-span-1">
									<CardHeader>
										<CardTitle>학습 진행 상황</CardTitle>
										<CardDescription>라운드별 정확도와 손실</CardDescription>
									</CardHeader>
									<CardContent>
										<ResponsiveContainer width="100%" height={300}>
											<LineChart
												data={trainingProgress}
												margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
											>
												<CartesianGrid strokeDasharray="3 3" />
												<XAxis dataKey="round" />
												<YAxis />
												<Tooltip />
												<Legend />
												<Line
													type="monotone"
													dataKey="accuracy"
													stroke="#3B82F6"
													name="정확도"
													strokeWidth={2}
												/>
												<Line
													type="monotone"
													dataKey="loss"
													stroke="#EF4444"
													name="손실"
													strokeWidth={2}
												/>
											</LineChart>
										</ResponsiveContainer>
									</CardContent>
								</Card>

								<Card className="col-span-1">
									<CardHeader>
										<CardTitle>클라우드 분포</CardTitle>
										<CardDescription>
											클라우드 제공업체별 자원 분포
										</CardDescription>
									</CardHeader>
									<CardContent>
										<ResponsiveContainer width="100%" height={300}>
											<PieChart>
												<Pie
													data={clusterData}
													cx="50%"
													cy="50%"
													outerRadius={100}
													dataKey="value"
													label={({ name, percent }) =>
														`${name} ${(percent * 100).toFixed(0)}%`
													}
												>
													{clusterData.map((entry, index) => (
														<Cell key={`cell-${index}`} fill={entry.color} />
													))}
												</Pie>
												<Tooltip />
												<Legend />
											</PieChart>
										</ResponsiveContainer>
									</CardContent>
								</Card>
							</div>

							{/* 활성 모델 목록 */}
							<Card>
								<CardHeader className="flex flex-row items-center justify-between">
									<div>
										<CardTitle>활성 모델</CardTitle>
										<CardDescription>
											현재 플랫폼에서 실행 중인 모델
										</CardDescription>
									</div>
									<Button variant="outline" size="sm">
										<PlusCircle className="h-4 w-4 mr-2" />새 모델
									</Button>
								</CardHeader>
								<CardContent>
									<div className="space-y-4">
										{activeModels.map((model) => (
											<div
												key={model.id}
												className="flex items-center justify-between pb-4 border-b last:border-0 last:pb-0"
											>
												<div className="space-y-1">
													<div className="flex items-center">
														<h4 className="font-medium mr-2">{model.name}</h4>
														{renderStatusBadge(model.status)}
													</div>
													<div className="text-sm text-muted-foreground">
														참여 노드: {model.participatingNodes} • 마지막
														업데이트: {model.lastUpdate}
													</div>
												</div>
												<div className="flex items-center gap-4">
													<div className="w-32">
														<Progress value={model.progress} className="h-2" />
														<div className="text-xs text-right mt-1">
															{model.progress}%
														</div>
													</div>
													<DropdownMenu>
														<DropdownMenuTrigger asChild>
															<Button variant="ghost" size="icon">
																<MoreHorizontal className="h-4 w-4" />
															</Button>
														</DropdownMenuTrigger>
														<DropdownMenuContent align="end">
															<DropdownMenuItem>세부 정보</DropdownMenuItem>
															<DropdownMenuItem>일시 중지</DropdownMenuItem>
															<DropdownMenuItem>재시작</DropdownMenuItem>
															<DropdownMenuSeparator />
															<DropdownMenuItem>삭제</DropdownMenuItem>
														</DropdownMenuContent>
													</DropdownMenu>
												</div>
											</div>
										))}
									</div>
								</CardContent>
							</Card>

							{/* 자원 사용량 및 최근 활동 */}
							<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
								<Card className="col-span-1">
									<CardHeader>
										<CardTitle>자원 사용량</CardTitle>
										<CardDescription>
											클러스터별 CPU, 메모리, 네트워크 사용량
										</CardDescription>
									</CardHeader>
									<CardContent>
										<ResponsiveContainer width="100%" height={300}>
											<BarChart
												data={resourceUsage}
												margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
											>
												<CartesianGrid strokeDasharray="3 3" />
												<XAxis dataKey="name" />
												<YAxis />
												<Tooltip />
												<Legend />
												<Bar dataKey="cpu" fill="#3B82F6" name="CPU (%)" />
												<Bar
													dataKey="memory"
													fill="#10B981"
													name="메모리 (%)"
												/>
												<Bar
													dataKey="network"
													fill="#F59E0B"
													name="네트워크 (%)"
												/>
											</BarChart>
										</ResponsiveContainer>
									</CardContent>
								</Card>

								<Card className="col-span-1">
									<CardHeader>
										<CardTitle>최근 활동</CardTitle>
										<CardDescription>플랫폼의 최근 활동 내역</CardDescription>
									</CardHeader>
									<CardContent>
										<div className="space-y-4">
											{recentActivities.map((activity) => (
												<div
													key={activity.id}
													className="flex items-start pb-4 border-b last:border-0 last:pb-0"
												>
													<div className="mr-4 mt-0.5 bg-primary/10 p-2 rounded-full">
														<Activity className="h-4 w-4 text-primary" />
													</div>
													<div className="space-y-1">
														<p className="font-medium">{activity.action}</p>
														<p className="text-sm text-muted-foreground">
															{activity.details}
														</p>
														<p className="text-xs text-muted-foreground">
															{activity.user} • {activity.time}
														</p>
													</div>
												</div>
											))}
										</div>
									</CardContent>
									<CardFooter>
										<Button variant="outline" size="sm" className="w-full">
											모든 활동 보기
										</Button>
									</CardFooter>
								</Card>
							</div>
						</TabsContent>

						<TabsContent value="models">
							<Card>
								<CardHeader>
									<CardTitle>모델 관리</CardTitle>
									<CardDescription>
										연합학습 모델의 관리 및 배포
									</CardDescription>
								</CardHeader>
								<CardContent>
									<p>모델 관리 탭 콘텐츠가 여기에 표시됩니다.</p>
								</CardContent>
							</Card>
						</TabsContent>

						<TabsContent value="resources">
							<Card>
								<CardHeader>
									<CardTitle>자원 관리</CardTitle>
									<CardDescription>
										클라우드 자원의 할당 및 모니터링
									</CardDescription>
								</CardHeader>
								<CardContent>
									<p>자원 관리 탭 콘텐츠가 여기에 표시됩니다.</p>
								</CardContent>
							</Card>
						</TabsContent>

						<TabsContent value="activity">
							<Card>
								<CardHeader>
									<CardTitle>활동 로그</CardTitle>
									<CardDescription>시스템 및 사용자 활동 기록</CardDescription>
								</CardHeader>
								<CardContent>
									<p>활동 로그 탭 콘텐츠가 여기에 표시됩니다.</p>
								</CardContent>
							</Card>
						</TabsContent>
					</Tabs>
				</main>
			</div>
		</div>
	);
};

export default Dashboard;
