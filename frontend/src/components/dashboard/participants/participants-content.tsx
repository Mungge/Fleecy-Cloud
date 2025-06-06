"use client";

import { useState, useEffect, useCallback } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Progress } from "@/components/ui/progress";
import { Plus, Edit, Trash2, Activity, CheckCircle } from "lucide-react";
import { toast } from "sonner";
import {
	createParticipant,
	getParticipants,
	updateParticipant,
	deleteParticipant,
	monitorVM,
	healthCheckVM,
} from "@/api/participants";
import type { Participant, VMMonitoringInfo } from "@/types/federatedLearning";

// 폼 스키마 정의 (YAML 파일 업로드 방식으로 변경)
const participantSchema = z.object({
	name: z.string().min(1, "이름은 필수입니다"),
	metadata: z.string().optional(),
});

type ParticipantFormData = z.infer<typeof participantSchema>;

export default function ParticipantsContent() {
	const [participants, setParticipants] = useState<Participant[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [createDialogOpen, setCreateDialogOpen] = useState(false);
	const [editDialogOpen, setEditDialogOpen] = useState(false);
	const [monitorDialogOpen, setMonitorDialogOpen] = useState(false);
	const [selectedParticipant, setSelectedParticipant] =
		useState<Participant | null>(null);
	const [isMonitoringLoading, setIsMonitoringLoading] = useState(false);
	const [monitoringData, setMonitoringData] = useState<VMMonitoringInfo | null>(
		null
	);
	const [isRealtimeEnabled, setIsRealtimeEnabled] = useState(false);
	const [realtimeMonitoring, setRealtimeMonitoring] = useState<
		Map<string, VMMonitoringInfo>
	>(new Map());
	const [monitoringInterval, setMonitoringInterval] =
		useState<NodeJS.Timeout | null>(null);
	const [configFile, setConfigFile] = useState<File | null>(null);

	const form = useForm<ParticipantFormData>({
		resolver: zodResolver(participantSchema),
		defaultValues: {
			name: "",
			metadata: "",
		},
	});

	// 클러스터 목록 로드
	const loadParticipants = async () => {
		try {
			setIsLoading(true);
			const data = await getParticipants();
			setParticipants(data);
		} catch (error) {
			console.error("클러스터 목록 로드 실패:", error);
			toast.error("클러스터 목록을 불러오는데 실패했습니다.");
		} finally {
			setIsLoading(false);
		}
	};

	// YAML 파일 업로드 처리
	const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];

			// YAML 파일 확장자 검증
			if (
				!file.name.toLowerCase().endsWith(".yaml") &&
				!file.name.toLowerCase().endsWith(".yml")
			) {
				toast.error("YAML 파일만 업로드 가능합니다.");
				return;
			}

			setConfigFile(file);
		}
	};

	useEffect(() => {
		loadParticipants();
	}, []);

	// 클러스터 생성
	const handleCreateParticipant = async (data: ParticipantFormData) => {
		try {
			// FormData 생성
			const formData = new FormData();
			formData.append("name", data.name);
			if (data.metadata) {
				formData.append("metadata", data.metadata);
			}

			// YAML 설정 파일 추가
			if (configFile) {
				formData.append("configFile", configFile);
			}
			await createParticipant(formData);

			toast.success("클러스터가 성공적으로 추가되었습니다.");

			form.reset({
				name: "",
				metadata: "",
			});
			setConfigFile(null);
			setCreateDialogOpen(false);
			loadParticipants();
		} catch (error) {
			console.error("참여자 생성 실패:", error);
			toast.error("클러스터 추가에 실패했습니다.");
		}
	};

	// 참여자 수정
	const handleUpdateParticipant = async (data: ParticipantFormData) => {
		if (!selectedParticipant) return;

		try {
			if (configFile) {
				const formData = new FormData();
				formData.append("name", data.name);
				if (data.metadata) {
					formData.append("metadata", data.metadata);
				}
				formData.append("configFile", configFile);

				await updateParticipant(selectedParticipant.id, formData);
			} else {
				// 파일이 없는 경우 FormData만 사용 (name, metadata만)
				const formData = new FormData();
				formData.append("name", data.name);
				if (data.metadata) {
					formData.append("metadata", data.metadata);
				}
				await updateParticipant(selectedParticipant.id, formData);
			}
			toast.success("클러스터 정보가 성공적으로 수정되었습니다.");
			setEditDialogOpen(false);
			setSelectedParticipant(null);
			setConfigFile(null);
			form.reset();
			loadParticipants();
		} catch (error) {
			console.error("참여자 수정 실패:", error);
			toast.error("클러스터 수정에 실패했습니다.");
		}
	};

	// 참여자 삭제
	const handleDeleteParticipant = async (id: string) => {
		try {
			await deleteParticipant(id);
			toast.success("참여자가 성공적으로 삭제되었습니다.");
			loadParticipants();
		} catch (error) {
			console.error("참여자 삭제 실패:", error);
			toast.error("참여자 삭제에 실패했습니다.");
		}
	};

	// VM 모니터링
	const handleMonitorVM = async (participant: Participant) => {
		setSelectedParticipant(participant);
		setIsMonitoringLoading(true);
		setMonitorDialogOpen(true);

		try {
			const monitoring = await monitorVM(participant.id);
			setMonitoringData(monitoring);
		} catch (error) {
			console.error("VM 모니터링 실패:", error);
			toast.error("VM 모니터링에 실패했습니다.");
			setMonitoringData(null);
		} finally {
			setIsMonitoringLoading(false);
		}
	};

	// VM 헬스체크
	const handleHealthCheck = async (participant: Participant) => {
		try {
			const healthResult = await healthCheckVM(participant.id);

			// 상세한 헬스체크 결과 표시
			if (healthResult.healthy) {
				toast.success(
					<div className="space-y-1">
						<div className="font-semibold">
							✅ {participant.name} 헬스체크 성공
						</div>
						<div className="text-sm">상태: {healthResult.status}</div>
						<div className="text-sm">
							응답시간: {healthResult.response_time_ms}ms
						</div>
						<div className="text-sm">{healthResult.message}</div>
					</div>,
					{
						duration: 5000,
					}
				);
			} else {
				toast.error(
					<div className="space-y-1">
						<div className="font-semibold">
							❌ {participant.name} 헬스체크 실패
						</div>
						<div className="text-sm">상태: {healthResult.status}</div>
						<div className="text-sm">
							응답시간: {healthResult.response_time_ms}ms
						</div>
						<div className="text-sm">{healthResult.message}</div>
					</div>,
					{
						duration: 8000,
					}
				);
			}
		} catch (error) {
			console.error("헬스체크 실패:", error);
			toast.error(
				<div className="space-y-1">
					<div className="font-semibold">
						🚨 {participant.name} 헬스체크 오류
					</div>
					<div className="text-sm">
						{error instanceof Error ? error.message : String(error)}
					</div>
				</div>
			);
		}
	};

	// 편집 다이얼로그 열기
	const openEditDialog = (participant: Participant) => {
		setSelectedParticipant(participant);
		form.reset({
			name: participant.name,
			metadata: participant.metadata || "",
		});
		setEditDialogOpen(true);
	};

	// 실시간 모니터링 토글
	const toggleRealtimeMonitoring = useCallback(async () => {
		if (isRealtimeEnabled) {
			// 모니터링 중지
			if (monitoringInterval) {
				clearInterval(monitoringInterval);
				setMonitoringInterval(null);
			}
			setIsRealtimeEnabled(false);
			setRealtimeMonitoring(new Map());
			toast("실시간 모니터링이 중지되었습니다.");
		} else {
			// 모니터링 시작
			setIsRealtimeEnabled(true);

			const updateMonitoring = async () => {
				const newMonitoringData = new Map<string, VMMonitoringInfo>();

				await Promise.all(
					participants.map(async (participant) => {
						try {
							const monitoring = await monitorVM(participant.id);
							newMonitoringData.set(participant.id, monitoring);
						} catch (error) {
							toast.error(
								`VM 모니터링 실패 (${participant.name}): ${
									error instanceof Error ? error.message : String(error)
								}`
							);
						}
					})
				);

				setRealtimeMonitoring(newMonitoringData);
			};

			// 즉시 한 번 실행
			await updateMonitoring();

			// 30초마다 업데이트
			const interval = setInterval(updateMonitoring, 30000);
			setMonitoringInterval(interval);

			toast("vm 실시간 모니터링이 시작되었습니다.");
		}
	}, [isRealtimeEnabled, monitoringInterval, participants]);

	// 컴포넌트 언마운트 시 인터벌 정리
	useEffect(() => {
		return () => {
			if (monitoringInterval) {
				clearInterval(monitoringInterval);
			}
		};
	}, [monitoringInterval]);

	// 유틸리티 함수들
	const getStatusBadge = (status: string) => {
		const colorClass =
			status === "active"
				? "bg-green-500"
				: status === "inactive"
				? "bg-gray-500"
				: "bg-yellow-500";
		return <Badge className={colorClass}>{status}</Badge>;
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-3xl font-bold tracking-tight">
						연합학습 클러스터 관리
					</h2>
					<p className="text-muted-foreground">
						연합학습에 클러스터를 관리하세요.
					</p>
				</div>

				<div className="flex items-center gap-4">
					<Button
						variant={isRealtimeEnabled ? "destructive" : "outline"}
						onClick={toggleRealtimeMonitoring}
					>
						<Activity className="mr-2 h-4 w-4" />
						{isRealtimeEnabled ? "모니터링 중지" : "실시간 모니터링"}
					</Button>

					<Dialog
						open={createDialogOpen}
						onOpenChange={(open) => {
							if (open) {
								form.reset({
									name: "",
									metadata: "",
								});
								setConfigFile(null);
							}
							setCreateDialogOpen(open);
							if (!open) {
								form.reset({
									name: "",
									metadata: "",
								});
								setConfigFile(null);
							}
						}}
					>
						<DialogTrigger asChild>
							<Button>
								<Plus className="mr-2 h-4 w-4" />
								클러스터 추가
							</Button>
						</DialogTrigger>
						<DialogContent className="max-w-2xl">
							<DialogHeader>
								<DialogTitle>클러스터 추가</DialogTitle>
								<DialogDescription>
									새로운 클러스터 정보를 입력하세요.
								</DialogDescription>
							</DialogHeader>

							<Form {...form}>
								<form
									onSubmit={form.handleSubmit(handleCreateParticipant)}
									className="space-y-4"
								>
									<FormField
										control={form.control}
										name="name"
										render={({ field }) => (
											<FormItem>
												<FormLabel>이름</FormLabel>
												<FormControl>
													<Input placeholder="참여자 이름" {...field} />
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>

									<FormField
										control={form.control}
										name="metadata"
										render={({ field }) => (
											<FormItem>
												<FormLabel>메타데이터</FormLabel>
												<FormControl>
													<Input placeholder="추가 정보" {...field} />
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>

									{/* OpenStack 설정 YAML 파일 업로드 */}
									<div className="space-y-4 border-t pt-4">
										<div>
											<Label className="text-base font-semibold">
												OpenStack 설정
											</Label>
											<p className="text-sm text-muted-foreground mt-1">
												OpenStack 클러스터 설정이 포함된 YAML 파일을
												업로드하세요.
											</p>
										</div>

										<div className="space-y-2">
											<Label htmlFor="config-file">
												설정 파일 (*.yaml, *.yml)
											</Label>
											<Input
												id="config-file"
												type="file"
												accept=".yaml,.yml"
												onChange={handleFileChange}
												className="cursor-pointer"
											/>
											{configFile && (
												<div className="text-sm text-green-600">
													선택된 파일: {configFile.name} (
													{Math.round(configFile.size / 1024)} KB)
												</div>
											)}
										</div>
									</div>

									<DialogFooter>
										<Button type="submit">클러스터 추가</Button>
									</DialogFooter>
								</form>
							</Form>
						</DialogContent>
					</Dialog>
				</div>
			</div>

			{/* 편집 다이얼로그 */}
			<Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
				<DialogContent className="max-w-2xl">
					<DialogHeader>
						<DialogTitle>클러스터 수정</DialogTitle>
						<DialogDescription>클러스터 정보를 수정하세요.</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form
							onSubmit={form.handleSubmit(handleUpdateParticipant)}
							className="space-y-4"
						>
							<FormField
								control={form.control}
								name="name"
								render={({ field }) => (
									<FormItem>
										<FormLabel>이름</FormLabel>
										<FormControl>
											<Input placeholder="클러스터 이름" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="metadata"
								render={({ field }) => (
									<FormItem>
										<FormLabel>메타데이터</FormLabel>
										<FormControl>
											<Input placeholder="추가 정보" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							{/* OpenStack 설정 업데이트 */}
							<div className="space-y-4 border-t pt-4">
								<h3 className="text-lg font-semibold">
									OpenStack 설정 업데이트
								</h3>
								<p className="text-sm text-muted-foreground">
									기존 설정을 유지하거나 새로운 YAML 파일로 업데이트할 수
									있습니다.
								</p>

								<div className="space-y-2">
									<Label htmlFor="edit-config-file">
										새 설정 파일 (선택사항)
									</Label>
									<Input
										id="edit-config-file"
										type="file"
										accept=".yaml,.yml"
										onChange={handleFileChange}
										className="cursor-pointer"
									/>
									{configFile && (
										<div className="text-sm text-green-600">
											선택된 파일: {configFile.name} (
											{Math.round(configFile.size / 1024)} KB)
										</div>
									)}
									<p className="text-xs text-muted-foreground">
										파일을 선택하지 않으면 기존 설정이 유지됩니다.
									</p>
								</div>
							</div>

							<DialogFooter>
								<Button type="submit">클러스터 수정</Button>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</Dialog>

			{isLoading ? (
				<div className="flex justify-center items-center py-12">
					<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
				</div>
			) : (
				<>
					{/* 실시간 모니터링 대시보드 */}
					{isRealtimeEnabled && realtimeMonitoring.size > 0 && (
						<Card>
							<CardHeader>
								<CardTitle>실시간 VM 모니터링</CardTitle>
								<CardDescription>
									OpenStack 클라우드 클러스터들의 실시간 리소스 사용량
								</CardDescription>
							</CardHeader>
							<CardContent>
								<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
									{Array.from(realtimeMonitoring.entries()).map(
										([participantId, monitoringInfo]) => {
											const participant = participants.find(
												(p) => p.id === participantId
											);
											if (!participant) return null;

											return (
												<Card key={participantId} className="border-2">
													<CardHeader className="pb-2">
														<CardTitle className="text-lg">
															{participant.name}
														</CardTitle>
														<div className="flex items-center gap-2">
															<Badge
																className={
																	monitoringInfo.status === "ACTIVE"
																		? "bg-green-500"
																		: monitoringInfo.status === "SHUTOFF"
																		? "bg-gray-500"
																		: monitoringInfo.status === "ERROR"
																		? "bg-red-500"
																		: "bg-yellow-500"
																}
															>
																{monitoringInfo.status}
															</Badge>
															<span className="text-xs text-gray-500">
																{monitoringInfo.instance_id.substring(0, 8)}...
															</span>
														</div>
													</CardHeader>
													<CardContent className="space-y-3">
														{/* CPU 사용률 */}
														<div>
															<div className="flex justify-between text-sm mb-1">
																<span>CPU</span>
																<span>
																	{monitoringInfo.cpu_usage.toFixed(1)}%
																</span>
															</div>
															<Progress
																value={Math.min(monitoringInfo.cpu_usage, 100)}
																className="h-2"
															/>
														</div>

														{/* 메모리 사용률 */}
														<div>
															<div className="flex justify-between text-sm mb-1">
																<span>메모리</span>
																<span>
																	{monitoringInfo.memory_usage.toFixed(1)}%
																</span>
															</div>
															<Progress
																value={Math.min(
																	monitoringInfo.memory_usage,
																	100
																)}
																className="h-2"
															/>
														</div>

														{/* 디스크 사용률 */}
														<div>
															<div className="flex justify-between text-sm mb-1">
																<span>디스크</span>
																<span>
																	{monitoringInfo.disk_usage.toFixed(1)}%
																</span>
															</div>
															<Progress
																value={Math.min(monitoringInfo.disk_usage, 100)}
																className="h-2"
															/>
														</div>
													</CardContent>
												</Card>
											);
										}
									)}
								</div>
								<div className="text-xs text-gray-500 mt-4 text-center">
									마지막 업데이트: {new Date().toLocaleTimeString()} (30초마다
									자동 갱신)
								</div>
							</CardContent>
						</Card>
					)}

					<Card>
						<CardHeader>
							<CardTitle>클러스터 목록</CardTitle>
							<CardDescription>
								등록된 클러스터들을 관리하고 상태를 모니터링하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>이름</TableHead>
										<TableHead>상태</TableHead>
										<TableHead>생성일</TableHead>
										<TableHead>액션</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{participants.map((participant) => (
										<TableRow key={participant.id}>
											<TableCell className="font-medium">
												{participant.name}
											</TableCell>
											<TableCell>
												{getStatusBadge(participant.status)}
											</TableCell>
											<TableCell>
												{new Date(participant.created_at).toLocaleDateString()}
											</TableCell>
											<TableCell>
												<div className="flex space-x-1">
													{/* 기본 편집 버튼 */}
													<Button
														variant="outline"
														size="sm"
														onClick={() => openEditDialog(participant)}
														title="편집"
													>
														<Edit className="h-4 w-4" />
													</Button>

													{/* 모니터링 버튼 */}

													<Button
														variant="outline"
														size="sm"
														onClick={() => handleHealthCheck(participant)}
														title="헬스체크"
													>
														<CheckCircle className="h-4 w-4" />
													</Button>
													{/* 삭제 버튼 */}
													<AlertDialog>
														<AlertDialogTrigger asChild>
															<Button variant="outline" size="sm" title="삭제">
																<Trash2 className="h-4 w-4" />
															</Button>
														</AlertDialogTrigger>
														<AlertDialogContent>
															<AlertDialogHeader>
																<AlertDialogTitle>
																	클러스터 삭제
																</AlertDialogTitle>
																<AlertDialogDescription>
																	이 클러스터를 삭제하시겠습니까? 이 작업은
																	되돌릴 수 없습니다.
																</AlertDialogDescription>
															</AlertDialogHeader>
															<AlertDialogFooter>
																<AlertDialogCancel>취소</AlertDialogCancel>
																<AlertDialogAction
																	onClick={() =>
																		handleDeleteParticipant(participant.id)
																	}
																>
																	삭제
																</AlertDialogAction>
															</AlertDialogFooter>
														</AlertDialogContent>
													</AlertDialog>
												</div>
											</TableCell>
										</TableRow>
									))}
									{participants.length === 0 && (
										<TableRow>
											<TableCell colSpan={8} className="text-center py-8">
												클러스터가 없습니다. 새 클러스터를 추가해보세요.
											</TableCell>
										</TableRow>
									)}
								</TableBody>
							</Table>
						</CardContent>
					</Card>
				</>
			)}

			{/* VM 모니터링 다이얼로그 */}
			<Dialog open={monitorDialogOpen} onOpenChange={setMonitorDialogOpen}>
				<DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
					<DialogHeader>
						<DialogTitle>VM 모니터링 정보</DialogTitle>
						<DialogDescription>
							{selectedParticipant?.name}의 OpenStack VM 상태 및 리소스 사용량
						</DialogDescription>
					</DialogHeader>

					{isMonitoringLoading ? (
						<div className="flex items-center justify-center py-8">
							<div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
							<span className="ml-2">모니터링 정보를 가져오는 중...</span>
						</div>
					) : monitoringData ? (
						<div className="space-y-6">
							{/* VM 기본 정보 */}
							<Card>
								<CardHeader>
									<CardTitle className="text-lg">VM 기본 정보</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="grid grid-cols-2 gap-4">
										<div>
											<label className="text-sm font-medium text-gray-500">
												인스턴스 ID
											</label>
											<p className="mt-1">{monitoringData.instance_id}</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												상태
											</label>
											<p className="mt-1">
												<Badge
													className={
														monitoringData.status === "ACTIVE"
															? "bg-green-500"
															: monitoringData.status === "SHUTOFF"
															? "bg-gray-500"
															: monitoringData.status === "ERROR"
															? "bg-red-500"
															: "bg-yellow-500"
													}
												>
													{monitoringData.status}
												</Badge>
											</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												가용 영역
											</label>
											<p className="mt-1">{monitoringData.availability_zone}</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												호스트
											</label>
											<p className="mt-1">{monitoringData.host}</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												생성 시간
											</label>
											<p className="mt-1">
												{new Date(monitoringData.created_at).toLocaleString()}
											</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												업데이트 시간
											</label>
											<p className="mt-1">
												{new Date(monitoringData.updated_at).toLocaleString()}
											</p>
										</div>
									</div>
								</CardContent>
							</Card>

							{/* 리소스 사용량 */}
							<Card>
								<CardHeader>
									<CardTitle className="text-lg">리소스 사용량</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="grid grid-cols-3 gap-4">
										<div className="text-center">
											<div className="text-2xl font-bold text-blue-600">
												{monitoringData.cpu_usage.toFixed(1)}%
											</div>
											<div className="text-sm text-gray-500">CPU 사용률</div>
											<Progress
												value={monitoringData.cpu_usage}
												className="h-2 mt-2"
											/>
										</div>
										<div className="text-center">
											<div className="text-2xl font-bold text-green-600">
												{monitoringData.memory_usage.toFixed(1)}%
											</div>
											<div className="text-sm text-gray-500">메모리 사용률</div>
											<Progress
												value={monitoringData.memory_usage}
												className="h-2 mt-2"
											/>
										</div>
										<div className="text-center">
											<div className="text-2xl font-bold text-purple-600">
												{monitoringData.disk_usage.toFixed(1)}%
											</div>
											<div className="text-sm text-gray-500">디스크 사용률</div>
											<Progress
												value={monitoringData.disk_usage}
												className="h-2 mt-2"
											/>
										</div>
									</div>

									<div className="grid grid-cols-2 gap-4 mt-6">
										<div>
											<label className="text-sm font-medium text-gray-500">
												네트워크 입력
											</label>
											<p className="mt-1 text-lg font-semibold">
												{(monitoringData.network_in / 1024 / 1024).toFixed(2)}{" "}
												MB
											</p>
										</div>
										<div>
											<label className="text-sm font-medium text-gray-500">
												네트워크 출력
											</label>
											<p className="mt-1 text-lg font-semibold">
												{(monitoringData.network_out / 1024 / 1024).toFixed(2)}{" "}
												MB
											</p>
										</div>
									</div>
								</CardContent>
							</Card>

							{/* 연합 학습 정보 */}
							{monitoringData.federated_learning_status && (
								<Card>
									<CardHeader>
										<CardTitle className="text-lg">연합 학습 상태</CardTitle>
									</CardHeader>
									<CardContent>
										<div className="grid grid-cols-2 gap-4">
											<div>
												<label className="text-sm font-medium text-gray-500">
													학습 상태
												</label>
												<p className="mt-1">
													<Badge
														className={
															monitoringData.federated_learning_status ===
															"training"
																? "bg-blue-500"
																: monitoringData.federated_learning_status ===
																  "idle"
																? "bg-gray-500"
																: monitoringData.federated_learning_status ===
																  "completed"
																? "bg-green-500"
																: "bg-red-500"
														}
													>
														{monitoringData.federated_learning_status}
													</Badge>
												</p>
											</div>
											{monitoringData.current_task_id && (
												<div>
													<label className="text-sm font-medium text-gray-500">
														현재 작업 ID
													</label>
													<p className="mt-1 font-mono text-sm">
														{monitoringData.current_task_id}
													</p>
												</div>
											)}
											{monitoringData.task_progress !== undefined && (
												<div>
													<label className="text-sm font-medium text-gray-500">
														작업 진행률
													</label>
													<p className="mt-1">
														{monitoringData.task_progress.toFixed(1)}%
													</p>
													<Progress
														value={monitoringData.task_progress}
														className="h-2 mt-1"
													/>
												</div>
											)}
											{monitoringData.last_training_time && (
												<div>
													<label className="text-sm font-medium text-gray-500">
														마지막 학습 시간
													</label>
													<p className="mt-1">
														{new Date(
															monitoringData.last_training_time
														).toLocaleString()}
													</p>
												</div>
											)}
										</div>
									</CardContent>
								</Card>
							)}
						</div>
					) : (
						<div className="text-center py-8">
							<p className="text-gray-500">
								모니터링 정보를 가져올 수 없습니다.
							</p>
						</div>
					)}

					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setMonitorDialogOpen(false)}
						>
							닫기
						</Button>
						<Button
							onClick={() =>
								selectedParticipant && handleMonitorVM(selectedParticipant)
							}
							disabled={isMonitoringLoading}
						>
							{isMonitoringLoading ? "새로고침 중..." : "새로고침"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
