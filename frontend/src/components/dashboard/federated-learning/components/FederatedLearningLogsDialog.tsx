// components/dashboard/federated-learning/components/FederatedLearningLogsDialog.tsx
import { useState, useEffect, useCallback } from "react";
import Cookies from "js-cookie";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText, RefreshCw, Activity } from "lucide-react";
import { FederatedLearningJob } from "@/types/federated-learning";

interface FederatedLearningLogsDialogProps {
	job: FederatedLearningJob;
	children: React.ReactNode;
}

interface AggregatorLogs {
	aggregatorId: string;
	aggregatorName: string;
	publicIP: string;
	logFilePath: string;
	flowerLogs: string;
	timestamp: string;
}

interface LogsResponse {
	federatedLearningId: string;
	status: string;
	aggregatorLogs: AggregatorLogs;
}

export const FederatedLearningLogsDialog = ({
	job,
	children,
}: FederatedLearningLogsDialogProps) => {
	const [logs, setLogs] = useState<LogsResponse | null>(null);
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [isOpen, setIsOpen] = useState(false);
	const [autoRefresh, setAutoRefresh] = useState(false);

	const fetchLogs = useCallback(async () => {
		setIsLoading(true);
		setError(null);

		try {
			// 쿠키에서 토큰 가져오기
			const token = Cookies.get("token");

			if (!token) {
				throw new Error("로그인이 필요합니다. 다시 로그인해주세요.");
			}

			console.log(
				"API 호출:",
				`${process.env.NEXT_PUBLIC_API_URL}/api/federated-learning/${job.id}/logs`
			);
			console.log("토큰 존재:", !!token);

			const response = await fetch(
				`${process.env.NEXT_PUBLIC_API_URL}/api/federated-learning/${job.id}/logs`,
				{
					headers: {
						Authorization: `Bearer ${token}`,
						"Content-Type": "application/json",
					},
					credentials: "include", // 쿠키 포함
				}
			);

			console.log("응답 상태:", response.status);

			if (response.status === 401) {
				// 토큰이 만료되었거나 유효하지 않음
				throw new Error("인증이 만료되었습니다. 다시 로그인해주세요.");
			}

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({}));
				throw new Error(
					errorData.error || `HTTP error! status: ${response.status}`
				);
			}

			const data = await response.json();
			console.log("로그 데이터:", data);
			setLogs(data.data);
		} catch (err) {
			console.error("로그 조회 에러:", err);
			setError(
				err instanceof Error
					? err.message
					: "로그를 가져오는 중 오류가 발생했습니다."
			);
		} finally {
			setIsLoading(false);
		}
	}, [job.id]);

	useEffect(() => {
		if (isOpen) {
			fetchLogs();
		}
	}, [isOpen, fetchLogs]);

	useEffect(() => {
		let interval: NodeJS.Timeout;

		if (autoRefresh && isOpen) {
			interval = setInterval(() => {
				fetchLogs();
			}, 5000); // 5초마다 갱신
		}

		return () => {
			if (interval) {
				clearInterval(interval);
			}
		};
	}, [autoRefresh, isOpen, fetchLogs]);

	const formatLogText = (text: string) => {
		return text.split("\n").map((line, index) => (
			<div key={index} className="font-mono text-xs">
				{line || "\u00A0"}
			</div>
		));
	};

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger asChild>{children}</DialogTrigger>
			<DialogContent className="max-w-6xl max-h-[90vh] overflow-hidden">
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2">
						<FileText className="h-5 w-5" />
						연합학습 로그 - {job.name}
					</DialogTitle>
					<DialogDescription>
						집계자의 실행 로그와 시스템 상태를 확인할 수 있습니다.
					</DialogDescription>
				</DialogHeader>

				<div className="flex gap-2 mb-4">
					<Button
						onClick={fetchLogs}
						disabled={isLoading}
						size="sm"
						variant="outline"
					>
						<RefreshCw
							className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
						/>
						새로고침
					</Button>
					<Button
						onClick={() => setAutoRefresh(!autoRefresh)}
						size="sm"
						variant={autoRefresh ? "default" : "outline"}
					>
						<Activity className="h-4 w-4 mr-2" />
						{autoRefresh ? "자동 갱신 중" : "자동 갱신"}
					</Button>
				</div>

				<ScrollArea className="h-[70vh]">
					{error && (
						<div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
							<div className="flex items-start gap-2">
								<div className="text-red-600 font-medium">오류 발생:</div>
								<div className="text-red-800 text-sm flex-1">{error}</div>
							</div>
							{error.includes("인증") && (
								<div className="mt-2 text-red-700 text-xs">
									• 브라우저를 새로고침하고 다시 로그인해보세요
									<br />• 토큰이 만료되었을 수 있습니다
								</div>
							)}
						</div>
					)}

					{logs && (
						<div className="space-y-4">
							{/* 집계자 정보 */}
							<Card>
								<CardHeader className="pb-3">
									<CardTitle className="text-base">집계자 정보</CardTitle>
								</CardHeader>
								<CardContent className="space-y-2">
									<div className="grid grid-cols-2 gap-4 text-sm">
										<div>
											<span className="font-medium">이름:</span>{" "}
											{logs.aggregatorLogs.aggregatorName}
										</div>
										<div>
											<span className="font-medium">IP:</span>{" "}
											{logs.aggregatorLogs.publicIP}
										</div>
										<div className="col-span-2">
											<span className="font-medium">로그 파일:</span>{" "}
											{logs.aggregatorLogs.logFilePath}
										</div>
										<div className="col-span-2">
											<span className="font-medium">마지막 업데이트:</span>{" "}
											{logs.aggregatorLogs.timestamp}
										</div>
									</div>
								</CardContent>
							</Card>

							{/* Flower 서버 로그 */}
							<Card>
								<CardHeader className="pb-3">
									<CardTitle className="text-base flex items-center justify-between">
										<span>Flower 서버 로그</span>
									</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="bg-black text-green-400 p-4 rounded-lg max-h-96 overflow-auto">
										{logs.aggregatorLogs.flowerLogs ? (
											formatLogText(logs.aggregatorLogs.flowerLogs)
										) : (
											<div className="text-gray-500 italic">
												로그가 없습니다.
											</div>
										)}
									</div>
								</CardContent>
							</Card>
						</div>
					)}

					{isLoading && !logs && (
						<div className="flex justify-center items-center h-40">
							<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
						</div>
					)}
				</ScrollArea>
			</DialogContent>
		</Dialog>
	);
};
