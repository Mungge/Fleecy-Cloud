// components/dashboard/federated-learning/components/GlobalModelManagementDialog.tsx
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
import { Badge } from "@/components/ui/badge";
import { Settings, RefreshCw, Download, CheckCircle, Clock, XCircle, AlertTriangle, LetterText } from "lucide-react";
import { FederatedLearningJob } from "@/types/federated-learning";
import { toast } from "sonner";

interface GlobalModelManagementDialogProps {
	job: FederatedLearningJob;
	children: React.ReactNode;
}

type ApiRound = {
	id: string;
	aggregator_id: string;
	round: number;
	model_metrics?: {
		accuracy?: number;
		loss: number;
		precision?: number;
		recall?: number;
		f1_score?: number;
	};
	duration?: number;
	participants_count?: number;
	started_at?: string;
	completed_at?: string;
	created_at?: string;
};

type ApiResponse = {
	data?: {
		aggregatorId?: string;
		federatedLearningId?: string;
		lastUpdated?: string;
		totalRounds?: number;
		trainingHistory?: ApiRound[];
	};
};

type Round = {
	id: string;
	round: number;
	accuracy?: number;
	loss: number;
	precision?: number;
	recall?: number;
	f1Score?: number;
	duration: number;
	participantsCount: number;
	startedAt?: string;
	completedAt?: string;
	createdAt?: string;
};

type GlobalModelData = {
	aggregatorId: string;
	federatedLearningId: string;
	lastUpdated: string;
	totalRounds: number;
	rounds: Round[];
};

export const GlobalModelManagementDialog = ({
	job,
	children,
}: GlobalModelManagementDialogProps) => {
	const [modelData, setModelData] = useState<GlobalModelData | null>(null);
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [isOpen, setIsOpen] = useState(false);
    const [noDataAvailable, setNoDataAvailable] = useState(false);

	const fetchGlobalModelData = useCallback(async () => {
		setIsLoading(true);
		setError(null);
        setNoDataAvailable(false);

		try {
			const token = Cookies.get("token");

			if (!token) {
				throw new Error("로그인이 필요합니다. 다시 로그인해주세요.");
			}

			const response = await fetch(
				`${process.env.NEXT_PUBLIC_API_URL}/api/federated-learning/${job.id}/training-history`,
				{
					headers: {
						Authorization: `Bearer ${token}`,
						"Content-Type": "application/json",
					},
					credentials: "include",
				}
			);

			if (response.status === 401) {
				throw new Error("인증이 만료되었습니다. 다시 로그인해주세요.");
			}

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({}));
				throw new Error(
					errorData.error || `HTTP error! status: ${response.status}`
				);
			}

			const apiResponse: ApiResponse = await response.json();
			
			// API 응답을 GlobalModelData로 변환
			if (apiResponse.data?.trainingHistory && apiResponse.data.trainingHistory.length > 0) {
				const transformedData: GlobalModelData = {
					aggregatorId: apiResponse.data.aggregatorId || "",
					federatedLearningId: apiResponse.data.federatedLearningId || job.id,
					lastUpdated: apiResponse.data.lastUpdated || new Date().toISOString(),
					totalRounds: apiResponse.data.totalRounds || 0,
					rounds: apiResponse.data.trainingHistory.map((apiRound: ApiRound): Round => ({
						id: apiRound.id,
						round: apiRound.round,
						accuracy: apiRound.model_metrics?.accuracy,
						loss: apiRound.model_metrics?.loss || 0,
						precision: apiRound.model_metrics?.precision,
						recall: apiRound.model_metrics?.recall,
						f1Score: apiRound.model_metrics?.f1_score,
						duration: apiRound.duration || 0,
						participantsCount: apiRound.participants_count || 0,
						startedAt: apiRound.started_at,
						completedAt: apiRound.completed_at,
						createdAt: apiRound.created_at,
					})),
				};
				
				setModelData(transformedData);
			} else {
                setNoDataAvailable(true);
				throw new Error("응답 데이터가 올바르지 않습니다.");
			}
		} catch (err) {
			console.error("글로벌 모델 데이터 조회 에러:", err);
			setError(
				err instanceof Error
					? err.message
					: "글로벌 모델 데이터를 가져오는 중 오류가 발생했습니다."
			);
		} finally {
			setIsLoading(false);
		}
	}, [job.id]);

	useEffect(() => {
		if (isOpen) {
			fetchGlobalModelData();
		}
	}, [isOpen, fetchGlobalModelData]);

    const generateDownloadUrl = (roundNumber: number) => {
		const filename = `round-${roundNumber.toString().padStart(3, '0')}.pt`;
		return `${process.env.NEXT_PUBLIC_API_URL}/api/federated-learning/${job.id}/models/round/${roundNumber}/download/${filename}`;
	};

	const getStatusIcon = (status: string) => {
		return <CheckCircle className="h-4 w-4 text-green-600" />;
	};

	const getStatusText = (status: string) => {
		return "완료";
	};

	const getStatusColor = (status: string) => {
		return "bg-green-100 text-green-800";
	};

	const formatDate = (dateString?: string) => {
		if (!dateString) return "N/A";
		return new Date(dateString).toLocaleString("ko-KR");
	};

	const getBestRound = (rounds: Round[]) => {
		if (!rounds.length) return null;
		
		// loss가 가장 낮은 라운드 찾기
		let bestRound = rounds[0];
		for (const round of rounds) {
			if (round.loss && bestRound.loss && round.loss < bestRound.loss) {
				bestRound = round;
			}
		}
		
		return bestRound;
	};

	const handleDownloadModel = async (downloadUrl: string, round: number) => {
        let loadingToast = toast.loading("모델을 다운로드하는 중입니다...");
        
		try {
			const token = Cookies.get("token");

            if (!token) {
				toast.dismiss(loadingToast);
				toast.error("로그인이 필요합니다.");
				return;
			}

			const response = await fetch(downloadUrl, {
				headers: {
					Authorization: `Bearer ${token}`,
				},
			});

			if (!response.ok) {
				throw new Error("모델 다운로드에 실패했습니다.");
			}

            const contentLength = response.headers.get('content-length');
			if (contentLength) {
				const fileSize = (parseInt(contentLength) / 1024 / 1024).toFixed(2);
				toast.dismiss(loadingToast);
				loadingToast = toast.loading(`파일을 처리하는 중입니다... (${fileSize}MB)`);
			}

			const blob = await response.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.style.display = "none";
			a.href = url;
			a.download = `global_model_${modelData?.aggregatorId}_round-${round.toString().padStart(3, '0')}.pt`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);

            toast.dismiss(loadingToast);
			toast.success(`라운드 ${round} 모델이 성공적으로 다운로드되었습니다.`);

		} catch (error) {
			console.error("모델 다운로드 에러:", error);
            toast.dismiss(loadingToast);

            const errorMessage = error instanceof Error 
				? error.message 
				: "모델 다운로드 중 오류가 발생했습니다.";
			
            toast.error(errorMessage);
            setError(errorMessage);
		}
	};

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger asChild>{children}</DialogTrigger>
			<DialogContent className="!w-[90vw] !max-w-none max-h-[90vh] overflow-hidden">
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2">
						<Settings className="h-5 w-5" />
						글로벌 모델 관리 - {job.name}
					</DialogTitle>
					<DialogDescription>
						연합학습 과정에서 생성된 라운드별 글로벌 모델 정보를 확인하고 관리할 수 있습니다.
					</DialogDescription>
				</DialogHeader>

				<div className="flex gap-2 mb-4">
					<Button
						onClick={fetchGlobalModelData}
						disabled={isLoading}
						size="sm"
						variant="outline"
					>
						<RefreshCw
							className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
						/>
						새로고침
					</Button>
				</div>

				<ScrollArea className="h-[70vh]">
					{error && (
						<div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
							<div className="flex items-start gap-2">
								<div className="text-red-600 font-medium">오류 발생:</div>
								<div className="text-red-800 text-sm flex-1">{error}</div>
							</div>
						</div>
					)}

                {/* 데이터가 없는 경우 특별한 메시지 표시 */}
                {noDataAvailable && !isLoading && (
                    <div className="flex flex-col justify-center items-center h-[50vh] text-center">
                        <AlertTriangle className="h-16 w-16 text-orange-500 mb-4" />
                        <h3 className="text-lg font-medium text-gray-900 mb-2">
                            트레이닝 히스토리가 없습니다
                        </h3>
                        <p className="text-gray-600 mb-4">
                            집계자 정보를 확인해보고 다시 확인해주세요
                        </p>
                        <Button
                            onClick={fetchGlobalModelData}
                            variant="outline"
                            size="sm"
                        >
                            <RefreshCw className="h-4 w-4 mr-2" />
                            다시 확인
                        </Button>
                    </div>
                )}

					{modelData && (
						<div className="space-y-6">
							{/* 전체 요약 정보 */}
							<Card>
								<CardHeader className="pb-3">
									<CardTitle className="text-base">전체 요약</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
										<div className="text-center p-3 bg-muted rounded-lg">
											<div className="text-lg font-bold">{modelData.totalRounds}</div>
											<div className="text-muted-foreground">총 라운드</div>
										</div>
										<div className="text-center p-3 bg-muted rounded-lg">
											<div className="text-lg font-bold text-green-600">{modelData.rounds.length}</div>
											<div className="text-muted-foreground">완료된 라운드</div>
										</div>
										<div className="text-center p-3 bg-muted rounded-lg">
											<div className="text-lg font-bold text-blue-600">
												{getBestRound(modelData.rounds)?.loss ? getBestRound(modelData.rounds)!.loss.toFixed(4) : "N/A"}
											</div>
											<div className="text-muted-foreground">최저 손실값</div>
										</div>
										<div className="text-center p-3 bg-muted rounded-lg">
											<div className="text-lg font-bold text-purple-600">
												{getBestRound(modelData.rounds) ? `라운드 ${getBestRound(modelData.rounds)!.round}` : "N/A"}
											</div>
											<div className="text-muted-foreground">최고 성능 라운드</div>
                                            {getBestRound(modelData.rounds) && (
                                            <div className="mt-4">
                                                <Button
                                                    onClick={() => {
                                                        const bestRound = getBestRound(modelData.rounds)!;
                                                        handleDownloadModel(generateDownloadUrl(bestRound.round), bestRound.round);
                                                    }}
                                                    variant="outline"
                                                    size="sm"
                                                >
                                                    <Download className="h-4 w-4 mr-2" />
                                                    최고 성능 모델 다운로드
                                                </Button>
                                            </div>
                                        )}
										</div>
                                        
									</div>
								</CardContent>
							</Card>

							{/* 라운드별 모델 목록 */}
							<Card>
								<CardHeader className="pb-3">
									<CardTitle className="text-base">라운드별 글로벌 모델</CardTitle>
								</CardHeader>
								<CardContent>
									{modelData.rounds.length > 0 ? (
										<div className="space-y-4">
											{modelData.rounds.map((round) => (
												<div
													key={round.id}
													className="border rounded-lg p-4 hover:bg-muted/50 transition-colors"
												>
													<div className="flex items-center justify-between mb-3">
														<div className="flex items-center gap-3">
															<div className="font-semibold text-lg">라운드 {round.round}</div>
															<Badge className={getStatusColor("completed")}>
																<div className="flex items-center gap-1">
																	{getStatusIcon("completed")}
																	{getStatusText("completed")}
																</div>
															</Badge>
														</div>
														{round.completedAt && (
															<div className="text-sm text-muted-foreground">
																{formatDate(round.completedAt)}
															</div>
														)}
													</div>

													<div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-3">
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">참여자:</span>
															<div className="font-medium">{round.participantsCount}개</div>
														</div>
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">정확도:</span>
															<div className="font-medium text-green-600">
																{round.accuracy ? (round.accuracy * 100).toFixed(2) + "%" : "N/A"}
															</div>
														</div>
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">손실값:</span>
															<div className="font-medium">{round.loss ? round.loss.toFixed(4) : "N/A"}</div>
														</div>
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">소요 시간:</span>
															<div className="font-medium">{round.duration}초</div>
														</div>
													</div>

													{/* 추가 메트릭 */}
													{(round.f1Score || round.precision || round.recall) && (
														<div className="grid grid-cols-3 gap-4 mb-3 pt-2 border-t">
															{round.f1Score && (
																<div className="text-sm">
																	<span className="font-medium text-muted-foreground">F1 Score:</span>
																	<div className="font-medium">{round.f1Score.toFixed(3)}</div>
																</div>
															)}
															{round.precision && (
																<div className="text-sm">
																	<span className="font-medium text-muted-foreground">Precision:</span>
																	<div className="font-medium">{round.precision.toFixed(3)}</div>
																</div>
															)}
															{round.recall && (
																<div className="text-sm">
																	<span className="font-medium text-muted-foreground">Recall:</span>
																	<div className="font-medium">{round.recall.toFixed(3)}</div>
																</div>
															)}
														</div>
													)}

													{/* 시간 정보 */}
													<div className="grid grid-cols-2 gap-4 pt-2 border-t">
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">시작 시간:</span>
															<div className="font-medium">{formatDate(round.startedAt)}</div>
														</div>
														<div className="text-sm">
															<span className="font-medium text-muted-foreground">완료 시간:</span>
															<div className="font-medium">{formatDate(round.completedAt)}</div>
														</div>
													</div>
												</div>
											))}
										</div>
									) : (
										<div className="text-center py-8 text-muted-foreground">
											<p>아직 완료된 라운드가 없습니다.</p>
										</div>
									)}
								</CardContent>
							</Card>
						</div>
					)}

					{isLoading && !modelData && (
						<div className="flex justify-center items-center h-40">
							<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
						</div>
					)}

					{!isLoading && !modelData && !error && (
						<div className="flex justify-center items-center h-40 text-muted-foreground">
							<div className="text-center">
								<Settings className="h-12 w-12 mx-auto mb-2 opacity-50" />
								<p>글로벌 모델 데이터를 불러오려면 새로고침 버튼을 클릭하세요.</p>
							</div>
						</div>
					)}
				</ScrollArea>
			</DialogContent>
		</Dialog>
	);
};