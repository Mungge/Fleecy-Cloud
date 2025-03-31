import React from "react";
import { MoreHorizontal, PlusCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Progress } from "@/components/ui/progress";

interface ActiveModel {
	id: number;
	name: string;
	status: "학습 중" | "완료됨" | "대기 중" | "오류";
	progress: number;
	participatingNodes: number;
	lastUpdate: string;
}

interface ActiveModelsListProps {
	models: ActiveModel[];
}

const ActiveModelsList: React.FC<ActiveModelsListProps> = ({ models }) => {
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
				variant = "outline";
				break;
			case "대기 중":
				variant = "outline";
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
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<div>
					<CardTitle>활성 모델</CardTitle>
					<CardDescription>현재 플랫폼에서 실행 중인 모델</CardDescription>
				</div>
				<Button variant="outline" size="sm">
					<PlusCircle className="h-4 w-4 mr-2" />새 모델
				</Button>
			</CardHeader>
			<CardContent>
				<div className="space-y-4">
					{models.map((model) => (
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
									참여 노드: {model.participatingNodes} • 마지막 업데이트:{" "}
									{model.lastUpdate}
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
	);
};

export default ActiveModelsList;
