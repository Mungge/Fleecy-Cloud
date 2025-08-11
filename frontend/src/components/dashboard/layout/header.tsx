import React from "react";
import { Bell, Calendar, ChevronDown, Search } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/contexts/auth-context";

interface HeaderProps {
	isSidebarOpen: boolean;
}

const Header: React.FC<HeaderProps> = ({ isSidebarOpen }) => {
	const { user, logout } = useAuth();

	const handleLogout = async () => {
		await logout();
	};

	const getDate = () => {
		return new Date().toLocaleDateString("ko-KR", {
			year: "numeric",
			month: "long",
			day: "numeric",
		});
	};
	return (
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
					<Button variant="outline" size="sm" className="hidden md:flex gap-2">
						<Calendar className="h-4 w-4" />
						<span>{getDate()}</span>
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
									<AvatarFallback>
										{user?.name?.[0] || user?.email?.[0] || "사"}
									</AvatarFallback>
								</Avatar>
								{isSidebarOpen && (
									<span>{user?.name || user?.email || "사용자"}</span>
								)}
								<ChevronDown className="h-4 w-4 text-muted-foreground" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuLabel>내 계정</DropdownMenuLabel>
							<DropdownMenuSeparator />
							<DropdownMenuItem>프로필</DropdownMenuItem>
							<DropdownMenuItem>설정</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem onClick={handleLogout}>
								로그아웃
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
			</div>
		</header>
	);
};

export default Header;
