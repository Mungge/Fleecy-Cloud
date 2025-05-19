"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { usePathname, useRouter } from "next/navigation";
import { LayoutDashboard, Cloud, Network, Menu } from "lucide-react";

interface SidebarProps {
	isSidebarOpen: boolean;
	setIsSidebarOpen: (open: boolean) => void;
}

export default function Sidebar({
	isSidebarOpen,
	setIsSidebarOpen,
}: SidebarProps) {
	const pathname = usePathname();
	const router = useRouter();

	const routes = [
		{
			label: "개요",
			icon: LayoutDashboard,
			href: "/dashboard/overview",
		},
		{
			label: "어그리게이터",
			icon: Network,
			href: "/dashboard/aggregator",
		},
		{
			label: "클라우드",
			icon: Cloud,
			href: "/dashboard/clouds",
		},
	];

	return (
		<div
			className={cn(
				"relative h-full border-r bg-background transition-all duration-300",
				isSidebarOpen ? "w-64" : "w-16"
			)}
		>
			<Button
				variant="ghost"
				size="icon"
				className="absolute -right-4 top-4 z-50"
				onClick={() => setIsSidebarOpen(!isSidebarOpen)}
			>
				<Menu className="h-4 w-4" />
			</Button>

			<ScrollArea className="h-full py-4">
				<div className="space-y-4 py-4">
					<div className="px-3 py-2">
						<div className="space-y-1">
							{routes.map((route) => (
								<Button
									key={route.href}
									variant={pathname === route.href ? "secondary" : "ghost"}
									className={cn(
										"w-full justify-start",
										isSidebarOpen ? "px-4" : "px-2"
									)}
									onClick={() => router.push(route.href)}
								>
									<route.icon
										className={cn("h-4 w-4", isSidebarOpen ? "mr-2" : "")}
									/>
									{isSidebarOpen && <span>{route.label}</span>}
								</Button>
							))}
						</div>
					</div>
				</div>
			</ScrollArea>
		</div>
	);
}
