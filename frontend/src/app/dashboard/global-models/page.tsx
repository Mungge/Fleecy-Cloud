"use client";

import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export default function GlobalModelsPage() {
	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-3xl font-bold tracking-tight">글로벌 모델</h2>
					<p className="text-muted-foreground">
						연합학습을 통해 생성된 글로벌 모델을 관리하세요.
					</p>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>글로벌 모델 관리</CardTitle>
					<CardDescription>
						학습된 글로벌 모델을 관리하고 배포하세요.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex justify-center items-center h-40 text-muted-foreground">
						글로벌 모델 관리 기능이 곧 제공될 예정입니다.
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
