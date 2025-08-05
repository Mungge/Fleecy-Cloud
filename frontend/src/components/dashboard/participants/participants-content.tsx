"use client";

import { useState, useEffect } from "react";
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
import { Plus, Edit, Trash2, CheckCircle, Server } from "lucide-react";
import { toast } from "sonner";
import {
	createParticipant,
	getParticipants,
	updateParticipant,
	deleteParticipant,
	healthCheckVM,
	getOpenStackVMs,
} from "@/api/participants";
import { Participant } from "@/types/participant";
import { OpenStackVMInstance } from "@/types/virtual-machine";

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
	const [vmListDialogOpen, setVmListDialogOpen] = useState(false);
	const [selectedParticipant, setSelectedParticipant] =
		useState<Participant | null>(null);
	const [configFile, setConfigFile] = useState<File | null>(null);

	// VM 목록 관련 상태
	const [vmList, setVmList] = useState<OpenStackVMInstance[]>([]);
	const [isVmListLoading, setIsVmListLoading] = useState(false);

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
			// 삭제된 참여자가 선택되어 있었다면 선택 해제
			if (selectedParticipant?.id === id) {
				setSelectedParticipant(null);
			}
			loadParticipants();
		} catch (error) {
			console.error("참여자 삭제 실패:", error);
			toast.error("참여자 삭제에 실패했습니다.");
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

			// 헬스체크 완료 후 참여자 목록 새로고침으로 UI 상태 동기화
			await loadParticipants();
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

			// 오류 발생 시에도 참여자 목록 새로고침
			await loadParticipants();
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

	// VM 목록 조회
	const handleViewVMs = async (participant: Participant) => {
		setSelectedParticipant(participant);
		setIsVmListLoading(true);
		setVmListDialogOpen(true);

		try {
			const vms = await getOpenStackVMs(participant.id);
			setVmList(vms);
		} catch (error) {
			console.error("VM 목록 조회 실패:", error);
			toast.error("VM 목록을 불러오는데 실패했습니다.");
			setVmList([]);
		} finally {
			setIsVmListLoading(false);
		}
	};

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
											OpenStack 클러스터 설정이 포함된 YAML 파일을 업로드하세요.
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
				<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
					{/* 클러스터 목록 */}
					<Card className="md:col-span-2">
						<CardHeader>
							<CardTitle>클러스터 목록</CardTitle>
							<CardDescription>
								등록된 클러스터들을 관리하고 상태를 모니터링하세요. 행을
								클릭하면 상세 정보를 확인할 수 있습니다.
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
										<TableRow
											key={participant.id}
											className={`cursor-pointer hover:bg-muted/50 ${
												selectedParticipant?.id === participant.id
													? "bg-muted"
													: ""
											}`}
											onClick={() => setSelectedParticipant(participant)}
										>
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
												<div
													className="flex space-x-1"
													onClick={(e) => e.stopPropagation()}
												>
													<Button
														variant="outline"
														size="sm"
														onClick={() => openEditDialog(participant)}
														title="편집"
													>
														<Edit className="h-4 w-4" />
													</Button>
													<Button
														variant="outline"
														size="sm"
														onClick={() => handleViewVMs(participant)}
														title="VM 목록 보기"
													>
														<Server className="h-4 w-4" />
													</Button>
													<Button
														variant="outline"
														size="sm"
														onClick={() => handleHealthCheck(participant)}
														title="헬스체크"
													>
														<CheckCircle className="h-4 w-4" />
													</Button>
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

					{/* 상세 정보 카드 */}
					<Card>
						<CardHeader>
							<CardTitle>클러스터 상세 정보</CardTitle>
							<CardDescription>
								선택한 클러스터의 상세 정보를 확인하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							{selectedParticipant ? (
								<div className="space-y-4">
									<div>
										<span className="text-sm font-medium">이름:</span>
										<p className="text-sm">{selectedParticipant.name}</p>
									</div>
									<div>
										<span className="text-sm font-medium">상태:</span>
										<div className="mt-1">
											{getStatusBadge(selectedParticipant.status)}
										</div>
									</div>
									<div>
										<span className="text-sm font-medium">생성일:</span>
										<p className="text-sm">
											{new Date(
												selectedParticipant.created_at
											).toLocaleString()}
										</p>
									</div>
									{selectedParticipant.metadata && (
										<div>
											<span className="text-sm font-medium">메타데이터:</span>
											<p className="text-sm">{selectedParticipant.metadata}</p>
										</div>
									)}
									<div>
										<span className="text-sm font-medium">
											Cluster Endpoint:
										</span>
										<p className="text-sm font-mono break-all">
											{selectedParticipant.openstack_endpoint}
										</p>
									</div>

									{/* 액션 버튼들 */}
									<div className="space-y-2 pt-4 border-t">
										<h4 className="text-sm font-medium">액션</h4>
										<div className="flex flex-col gap-2">
											<Button
												variant="outline"
												size="sm"
												onClick={() => openEditDialog(selectedParticipant)}
												className="justify-start"
											>
												<Edit className="h-4 w-4 mr-2" />
												편집
											</Button>
											<Button
												variant="outline"
												size="sm"
												onClick={() => handleViewVMs(selectedParticipant)}
												className="justify-start"
											>
												<Server className="h-4 w-4 mr-2" />
												가상머신 목록
											</Button>
											<Button
												variant="outline"
												size="sm"
												onClick={() => handleHealthCheck(selectedParticipant)}
												className="justify-start"
											>
												<CheckCircle className="h-4 w-4 mr-2" />
												헬스체크
											</Button>
										</div>
									</div>
								</div>
							) : (
								<div className="text-center py-8 text-muted-foreground">
									<p>클러스터를 선택해주세요</p>
									<p className="text-sm mt-2">
										왼쪽 목록에서 클러스터를 클릭하면 상세 정보를 확인할 수
										있습니다.
									</p>
								</div>
							)}
						</CardContent>
					</Card>
				</div>
			)}

			{/* VM 목록 다이얼로그 */}
			<Dialog open={vmListDialogOpen} onOpenChange={setVmListDialogOpen}>
				<DialogContent className="max-w-[95vw] w-full max-h-[95vh] overflow-hidden flex flex-col">
					<DialogHeader>
						<DialogTitle>가상머신 목록</DialogTitle>
						<DialogDescription>
							{selectedParticipant?.name} 클러스터의 가상머신 목록
						</DialogDescription>
					</DialogHeader>

					{isVmListLoading ? (
						<div className="flex items-center justify-center py-8">
							<div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
							<span className="ml-2">VM 목록을 가져오는 중...</span>
						</div>
					) : (
						<div className="flex-1 overflow-auto space-y-4">
							{vmList.length > 0 ? (
								<>
									<div className="rounded-md border">
										<Table>
											<TableHeader className="sticky top-0 bg-white z-10">
												<TableRow>
													<TableHead className="w-[200px]">이름</TableHead>
													<TableHead className="w-[150px]">상태</TableHead>
													<TableHead className="w-[200px]">
														스펙 (CPU/RAM/Disk)
													</TableHead>
													<TableHead className="w-[250px]">IP 주소</TableHead>
												</TableRow>
											</TableHeader>
											<TableBody>
												{vmList.map((vm) => (
													<TableRow key={vm.id} className="hover:bg-muted/50">
														<TableCell className="align-top">
															<div>
																<div className="font-medium break-words">
																	{vm.name}
																</div>
															</div>
														</TableCell>
														<TableCell className="align-top">
															<div className="space-y-1">
																<Badge
																	className={
																		vm.status === "ACTIVE"
																			? "bg-green-500"
																			: vm.status === "SHUTOFF"
																			? "bg-gray-500"
																			: vm.status === "ERROR"
																			? "bg-red-500"
																			: "bg-yellow-500"
																	}
																>
																	{vm.status}
																</Badge>
																<div className="text-xs text-gray-500">
																	{vm["OS-EXT-STS:power_state"] === 1
																		? "Running"
																		: "Stopped"}
																</div>
															</div>
														</TableCell>
														<TableCell className="align-top">
															<div className="space-y-1">
																<div className="font-medium text-sm">
																	{vm.flavor.name || vm.flavor.id}
																</div>
																<div className="text-xs text-gray-600 space-y-0.5">
																	<div className="flex items-center gap-1">
																		<span className="font-mono">CPU:</span>
																		<span>{vm.flavor.vcpus || 0} vCPU</span>
																	</div>
																	<div className="flex items-center gap-1">
																		<span className="font-mono">RAM:</span>
																		<span>
																			{vm.flavor.ram
																				? `${(vm.flavor.ram / 1024).toFixed(
																						1
																				  )} GB`
																				: "0 GB"}
																		</span>
																	</div>
																	<div className="flex items-center gap-1">
																		<span className="font-mono">Disk:</span>
																		<span>{vm.flavor.disk || 0} GB</span>
																	</div>
																</div>
															</div>
														</TableCell>
														<TableCell className="align-top">
															<div className="space-y-1 max-w-[250px]">
																{Object.keys(vm.addresses || {}).length > 0 ? (
																	Object.entries(vm.addresses).map(
																		([networkName, addresses]) =>
																			addresses.map((addr, index) => (
																				<div
																					key={`${networkName}-${index}`}
																					className="space-y-1"
																				>
																					<div className="flex items-center gap-2 flex-wrap">
																						<span className="font-mono text-sm break-all">
																							{addr.addr}
																						</span>
																						<Badge
																							variant="outline"
																							className="text-xs flex-shrink-0"
																						>
																							{addr.type}
																						</Badge>
																					</div>
																					<div className="text-xs text-gray-500">
																						{networkName}
																					</div>
																				</div>
																			))
																	)
																) : (
																	<span className="text-sm text-gray-500">
																		없음
																	</span>
																)}
															</div>
														</TableCell>
													</TableRow>
												))}
											</TableBody>
										</Table>
									</div>

									<div className="flex items-center justify-between text-sm text-gray-500 px-2 pb-2">
										<span>총 {vmList.length}개의 가상머신이 있습니다.</span>
										<span>
											마지막 업데이트: {new Date().toLocaleTimeString()}
										</span>
									</div>
								</>
							) : (
								<div className="text-center py-12">
									<div className="mx-auto w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mb-4">
										<Server className="h-12 w-12 text-gray-400" />
									</div>
									<h3 className="text-lg font-medium text-gray-900 mb-2">
										가상머신이 없습니다
									</h3>
									<p className="text-gray-500">
										이 클러스터에는 아직 가상머신이 없습니다.
									</p>
								</div>
							)}
						</div>
					)}

					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setVmListDialogOpen(false)}
						>
							닫기
						</Button>
						<Button
							onClick={() =>
								selectedParticipant && handleViewVMs(selectedParticipant)
							}
							disabled={isVmListLoading}
						>
							{isVmListLoading ? "새로고침 중..." : "새로고침"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
