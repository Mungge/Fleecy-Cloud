"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { usePathname, useRouter } from "next/navigation";
import {
  LayoutDashboard,
  Cloud,
  Network,
  Menu,
  Merge,
  Globe,
  Users,
} from "lucide-react";

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
      label: "Overview",
      icon: LayoutDashboard,
      href: "/dashboard",
    },
    {
      label: "집계자",
      icon: Merge,
      href: "/dashboard/aggregator",
    },
    {
      label: "클라우드 인증 정보",
      icon: Cloud,
      href: "/dashboard/clouds",
    },
    {
      label: "연합학습 클러스터",
      icon: Users,
      href: "/dashboard/participants",
    },
    {
      label: "연합학습",
      icon: Network,
      href: "/dashboard/federated-learning",
    },
    {
      label: "글로벌 모델",
      icon: Globe,
      href: "/dashboard/global-models",
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
        className={cn(
          "absolute top-4 z-50",
          isSidebarOpen ? "-right-1" : "right-3"
        )}
        onClick={() => setIsSidebarOpen(!isSidebarOpen)}
      >
        <Menu className="h-4 w-4" />
      </Button>

      <ScrollArea className="h-full py-4">
        <div className="space-y-4 py-4">
          <div className="px-3 py-2">
            <div className="space-y-1 pt-8">
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
