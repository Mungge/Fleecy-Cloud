// src/components/dashboard/layout/sidebar.tsx 수정본
// 기존 sidebar.tsx 파일에 다음 변경사항을 적용해야 합니다

import React from "react";
import {
	Activity,
	Cloud,
	Database,
	Menu,
	Server,
	Settings,
	Users,
	X,
	Calculator, // 추가: 리소스 추정을 위한 아이콘
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";

interface SidebarProps {
	isSidebarOpen: boolean;
	setIsSidebarOpen: (isOpen: boolean) => void;
	activeTab: string;
	setActiveTab: (tab: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({
	isSidebarOpen,
	setIsSidebarOpen,
	activeTab,
	setActiveTab,
}) => {
	return (
		<div
			className={`${
				isSidebarOpen ? "w-64" : "w-16"
			} bg-card border-r transition-all duration-300 ease-in-out`}
		>
			<div className="p-4 flex justify-between items-center">
				{isSidebarOpen && <h1 className="font-bold text-lg">Fleecy-Cloud</h1>}
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

					{/* Aggregator Menu */}
					<Button
						variant={activeTab === "aggregator" ? "secondary" : "ghost"}
						className="w-full justify-start mb-1"
						onClick={() => setActiveTab("aggregator")}
					>
						<Calculator className="h-5 w-5 mr-2" />
						{isSidebarOpen && <span>Aggregator</span>}
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
	);
};

export default Sidebar;
