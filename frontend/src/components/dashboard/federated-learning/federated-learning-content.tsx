"use client";

import { useState, useEffect, useCallback } from "react";
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
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
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
import { Plus } from "lucide-react";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import * as z from "zod";
import { Slider } from "@/components/ui/slider";
import {
	getFederatedLearnings,
	createFederatedLearning,
	deleteFederatedLearning,
} from "@/api/federatedLearning";
import { getAvailableParticipants } from "@/api/participants";
import { FederatedLearningJob, Participant } from "@/types/federatedLearning";
import { toast } from "sonner";

// 집계 알고리즘 목록
const AGGREGATION_ALGORITHMS = [
	{ id: "fedavg", name: "FedAvg (Federated Averaging)" },
	{ id: "fedprox", name: "FedProx" },
	{ id: "scaffold", name: "SCAFFOLD" },
	{ id: "fedopt", name: "FedOpt" },
];

// 지원하는 모델 유형
const MODEL_TYPES = [
	{ id: "image_classification", name: "이미지 분류" },
	{ id: "nlp", name: "자연어 처리" },
	{ id: "tabular", name: "테이블 형식 데이터" },
];

const formSchema = z.object({
	name: z.string().min(1, "이름을 입력해주세요"),
	description: z.string().optional(),
	modelType: z.string().min(1, "모델 유형을 선택해주세요"),
	algorithm: z.string().min(1, "집계 알고리즘을 선택해주세요"),
	rounds: z
		.number()
		.int()
		.min(1, "최소 1회 이상의 라운드가 필요합니다")
		.max(100, "최대 100회까지 설정 가능합니다"),
	participants: z
		.array(z.string())
		.min(1, "최소 1개 이상의 참여자가 필요합니다"),
});

type FormValues = z.infer<typeof formSchema>;

const FederatedLearningContent = () => {
	const [federatedLearningJobs, setFederatedLearningJobs] = useState<
		FederatedLearningJob[]
	>([]);
	const [participants, setParticipants] = useState<Participant[]>([]);
	const [selectedJob, setSelectedJob] = useState<FederatedLearningJob | null>(
		null
	);
	const [createDialogOpen, setCreateDialogOpen] = useState(false);
	const [isLoading, setIsLoading] = useState(true);
	const [modelFile, setModelFile] = useState<File | null>(null);

	// 연합학습 생성 폼
	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			name: "",
			description: "",
			modelType: "",
			algorithm: "fedavg",
			rounds: 10,
			participants: [],
		},
	});

	const fetchFederatedLearningJobs = useCallback(async () => {
		try {
			setIsLoading(true);

			const jobs = await getFederatedLearnings();
			setFederatedLearningJobs(jobs);
		} catch (error) {
			toast.error("연합학습 작업 조회에 실패했습니다: " + error);
			setFederatedLearningJobs([]);
		} finally {
			setIsLoading(false);
		}
	}, []);

	// 참여자 목록 가져오기
	const fetchParticipants = useCallback(async () => {
		try {
			const participantList = await getAvailableParticipants();
			setParticipants(participantList);
		} catch (error) {
			toast.error("참여자 목록 조회에 실패했습니다: " + error);
			setParticipants([]);
		}
	}, []);

	// 페이지 로드 시 연합학습 작업 목록과 참여자 목록 가져오기
	useEffect(() => {
		let mounted = true;

		const loadData = async () => {
			if (mounted) {
				await Promise.all([fetchFederatedLearningJobs(), fetchParticipants()]);
			}
		};

		loadData();

		return () => {
			mounted = false;
		};
	}, [fetchFederatedLearningJobs, fetchParticipants]);

	// 파일 업로드 처리
	const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			setModelFile(e.target.files[0]);
		}
	};

	// 연합학습 생성 함수
	const handleCreateJob = async (values: FormValues) => {
		try {
			// FormData 생성
			const formData = new FormData();
			formData.append("name", values.name);
			if (values.description)
				formData.append("description", values.description);
			formData.append("modelType", values.modelType);
			formData.append("algorithm", values.algorithm);
			formData.append("rounds", values.rounds.toString());
			values.participants.forEach((participant) => {
				formData.append("participants[]", participant);
			});

			if (modelFile) {
				formData.append("modelFile", modelFile);
			}

			// API 호출
			await createFederatedLearning(formData);

			// 성공 메시지 표시
			toast.success("연합학습 작업이 생성되었습니다.");

			// 폼 초기화 및 다이얼로그 닫기
			form.reset();
			setModelFile(null);
			setCreateDialogOpen(false);

			// 목록 새로고침
			fetchFederatedLearningJobs();
		} catch (error) {
			toast.error("연합학습 작업 생성에 실패했습니다: " + error);
		}
	};

	// 연합학습 삭제 함수
	const handleDeleteJob = async (id: string) => {
		try {
			await deleteFederatedLearning(id);

			// 성공 메시지 표시
			toast.success("연합학습 작업이 삭제되었습니다.");

			// 목록 새로고침
			fetchFederatedLearningJobs();

			// 선택된 작업이 삭제된 작업인 경우 선택 해제
			if (selectedJob?.id === id) {
				setSelectedJob(null);
			}
		} catch (error) {
			console.error("연합학습 작업 삭제에 실패했습니다: ", error);
			toast.error("연합학습 작업 삭제에 실패했습니다.");
		}
	};

	// 상태에 따른 배지 색상 설정
	const getStatusBadge = (status: string) => {
		switch (status) {
			case "완료":
				return <Badge className="bg-green-500">완료</Badge>;
			case "진행중":
				return <Badge className="bg-blue-500">진행중</Badge>;
			case "대기중":
				return <Badge className="bg-yellow-500">대기중</Badge>;
			default:
				return <Badge>{status}</Badge>;
		}
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-3xl font-bold tracking-tight">연합학습</h2>
					<p className="text-muted-foreground">
						연합학습 작업을 생성하고 모니터링하세요.
					</p>
				</div>

				<Dialog
					open={createDialogOpen}
					onOpenChange={(open) => {
						setCreateDialogOpen(open);
						// 다이얼로그가 닫힐 때만 폼과 상태 초기화
						if (!open) {
							form.reset();
							setModelFile(null);
						}
					}}
				>
					<DialogTrigger asChild>
						<Button className="ml-auto">
							<Plus className="mr-2 h-4 w-4" />
							연합학습 생성
						</Button>
					</DialogTrigger>
					<DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
						<DialogHeader>
							<DialogTitle>연합학습 생성</DialogTitle>
							<DialogDescription>
								새로운 연합학습 작업에 필요한 정보를 입력하세요.
							</DialogDescription>
						</DialogHeader>

						<Form {...form}>
							<form
								onSubmit={form.handleSubmit(handleCreateJob)}
								className="space-y-6"
							>
								<Tabs defaultValue="basic" className="w-full">
									<TabsList className="grid grid-cols-3 mb-4">
										<TabsTrigger value="basic">기본 정보</TabsTrigger>
										<TabsTrigger value="model">모델 설정</TabsTrigger>
										<TabsTrigger value="participants">참여자 설정</TabsTrigger>
									</TabsList>

									<TabsContent value="basic" className="space-y-4">
										<FormField
											control={form.control}
											name="name"
											render={({ field }) => (
												<FormItem>
													<FormLabel>이름</FormLabel>
													<FormControl>
														<Input
															placeholder="연합학습 작업 이름"
															{...field}
														/>
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>

										<FormField
											control={form.control}
											name="description"
											render={({ field }) => (
												<FormItem>
													<FormLabel>설명</FormLabel>
													<FormControl>
														<Input
															placeholder="작업에 대한 간략한 설명"
															{...field}
														/>
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>

										<FormField
											control={form.control}
											name="algorithm"
											render={({ field }) => (
												<FormItem>
													<FormLabel>집계 알고리즘</FormLabel>
													<Select
														onValueChange={field.onChange}
														defaultValue={field.value}
													>
														<FormControl>
															<SelectTrigger>
																<SelectValue placeholder="집계 알고리즘 선택" />
															</SelectTrigger>
														</FormControl>
														<SelectContent>
															{AGGREGATION_ALGORITHMS.map((algo) => (
																<SelectItem key={algo.id} value={algo.id}>
																	{algo.name}
																</SelectItem>
															))}
														</SelectContent>
													</Select>
													<FormDescription>
														클라이언트 모델을 집계하는 알고리즘입니다.
													</FormDescription>
													<FormMessage />
												</FormItem>
											)}
										/>

										<FormField
											control={form.control}
											name="rounds"
											render={({ field }) => (
												<FormItem>
													<FormLabel>라운드 수: {field.value}</FormLabel>
													<FormControl>
														<Slider
															defaultValue={[field.value]}
															min={1}
															max={100}
															step={1}
															onValueChange={(vals) => field.onChange(vals[0])}
														/>
													</FormControl>
													<FormDescription>
														연합학습이 수행될 라운드 수를 설정하세요.
													</FormDescription>
													<FormMessage />
												</FormItem>
											)}
										/>
									</TabsContent>

									<TabsContent value="model" className="space-y-4">
										<FormField
											control={form.control}
											name="modelType"
											render={({ field }) => (
												<FormItem>
													<FormLabel>모델 유형</FormLabel>
													<Select
														onValueChange={field.onChange}
														defaultValue={field.value}
													>
														<FormControl>
															<SelectTrigger>
																<SelectValue placeholder="모델 유형 선택" />
															</SelectTrigger>
														</FormControl>
														<SelectContent>
															{MODEL_TYPES.map((type) => (
																<SelectItem key={type.id} value={type.id}>
																	{type.name}
																</SelectItem>
															))}
														</SelectContent>
													</Select>
													<FormDescription>
														연합학습에 사용될 모델의 유형을 선택하세요.
													</FormDescription>
													<FormMessage />
												</FormItem>
											)}
										/>

										<div className="space-y-2">
											<Label>모델 파일 업로드</Label>
											<div className="grid w-full max-w-sm items-center gap-1.5">
												<Input
													type="file"
													accept=".h5,.pb,.pt,.pth,.onnx,.pkl"
													onChange={handleFileChange}
												/>
												<p className="text-sm text-muted-foreground">
													지원 형식: .h5, .pb, .pt, .pth, .onnx, .pkl
												</p>
											</div>
											{modelFile && (
												<div className="text-sm">
													선택된 파일: {modelFile.name} (
													{Math.round(modelFile.size / 1024)} KB)
												</div>
											)}
										</div>
									</TabsContent>

									<TabsContent value="participants" className="space-y-4">
										<FormField
											control={form.control}
											name="participants"
											render={() => (
												<FormItem>
													<div className="mb-4">
														<FormLabel className="text-base">
															참여자 선택
														</FormLabel>
														<FormDescription>
															연합학습에 참여할 참여자를 선택하세요. 최소 1개
															이상의 참여자가 필요합니다.
														</FormDescription>
														<FormMessage />
													</div>
													<div className="space-y-4">
														{participants.map((participant) => (
															<FormField
																key={participant.id}
																control={form.control}
																name="participants"
																render={({ field }) => {
																	return (
																		<FormItem
																			key={participant.id}
																			className="flex flex-row items-start space-x-3 space-y-0"
																		>
																			<FormControl>
																				<Checkbox
																					disabled={
																						participant.status === "inactive"
																					}
																					checked={field.value?.includes(
																						participant.id
																					)}
																					onCheckedChange={(checked) => {
																						return checked
																							? field.onChange([
																									...field.value,
																									participant.id,
																							  ])
																							: field.onChange(
																									field.value?.filter(
																										(value) =>
																											value !== participant.id
																									)
																							  );
																					}}
																				/>
																			</FormControl>
																			<div className="space-y-1 leading-none">
																				<FormLabel
																					className={
																						participant.status === "inactive"
																							? "text-muted-foreground"
																							: ""
																					}
																				>
																					{participant.name}
																				</FormLabel>
																				<div className="text-xs text-muted-foreground">
																					{participant.openstack_endpoint ||
																						"OpenStack 엔드포인트 정보 없음"}
																					<span
																						className={
																							participant.status === "active"
																								? "text-green-500 ml-1"
																								: "text-red-500 ml-1"
																						}
																					>
																						{participant.status === "active"
																							? "활성"
																							: "비활성"}
																					</span>
																				</div>
																			</div>
																		</FormItem>
																	);
																}}
															/>
														))}
														{participants.length === 0 && (
															<div className="text-center py-4 text-muted-foreground">
																사용 가능한 참여자가 없습니다.
																<br />
																참여자 정보를 먼저 설정해주세요.
															</div>
														)}
													</div>
												</FormItem>
											)}
										/>
									</TabsContent>
								</Tabs>

								<DialogFooter>
									<Button type="submit">연합학습 생성</Button>
								</DialogFooter>
							</form>
						</Form>
					</DialogContent>
				</Dialog>
			</div>

			{isLoading ? (
				<div className="flex justify-center items-center py-12">
					<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
				</div>
			) : (
				<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
					{/* 연합학습 목록 */}
					<Card className="md:col-span-2">
						<CardHeader>
							<CardTitle>연합학습 목록</CardTitle>
							<CardDescription>
								연합학습 작업의 상태와 세부 정보를 확인하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<ScrollArea className="h-[calc(100vh-320px)]">
								<Table>
									<TableHeader>
										<TableRow>
											<TableHead>이름</TableHead>
											<TableHead>상태</TableHead>
											<TableHead>참여자</TableHead>
											<TableHead>생성일</TableHead>
											<TableHead>액션</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{federatedLearningJobs.map((job) => (
											<TableRow
												key={job.id}
												className="cursor-pointer"
												onClick={() => setSelectedJob(job)}
											>
												<TableCell className="font-medium">
													{job.name}
												</TableCell>
												<TableCell>{getStatusBadge(job.status)}</TableCell>
												<TableCell>{job.participants}</TableCell>
												<TableCell>{job.created_at}</TableCell>
												<TableCell>
													<AlertDialog>
														<AlertDialogTrigger asChild>
															<Button
																variant="outline"
																size="sm"
																onClick={(e) => e.stopPropagation()}
															>
																삭제
															</Button>
														</AlertDialogTrigger>
														<AlertDialogContent>
															<AlertDialogHeader>
																<AlertDialogTitle>
																	연합학습 삭제
																</AlertDialogTitle>
																<AlertDialogDescription>
																	이 연합학습 작업을 삭제하시겠습니까? 이 작업은
																	되돌릴 수 없습니다.
																</AlertDialogDescription>
															</AlertDialogHeader>
															<AlertDialogFooter>
																<AlertDialogCancel>취소</AlertDialogCancel>
																<AlertDialogAction
																	onClick={() => handleDeleteJob(job.id)}
																>
																	삭제
																</AlertDialogAction>
															</AlertDialogFooter>
														</AlertDialogContent>
													</AlertDialog>
												</TableCell>
											</TableRow>
										))}
									</TableBody>
								</Table>
							</ScrollArea>
						</CardContent>
					</Card>

					{/* 연합학습 상세 정보 */}
					<Card>
						<CardHeader>
							<CardTitle>연합학습 상세 정보</CardTitle>
							<CardDescription>
								선택한 연합학습의 세부 정보를 확인하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							{selectedJob ? (
								<div className="space-y-4">
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">ID:</div>
										<div className="text-sm col-span-2">{selectedJob.id}</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">이름:</div>
										<div className="text-sm col-span-2">{selectedJob.name}</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">상태:</div>
										<div className="text-sm col-span-2">
											{getStatusBadge(selectedJob.status)}
										</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">참여자:</div>
										<div className="text-sm col-span-2">
											{selectedJob.participants}명
										</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">라운드 수:</div>
										<div className="text-sm col-span-2">
											{selectedJob.rounds || "-"}
										</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">알고리즘:</div>
										<div className="text-sm col-span-2">
											{AGGREGATION_ALGORITHMS.find(
												(algo) => algo.id === selectedJob.algorithm
											)?.name ||
												selectedJob.algorithm ||
												"-"}
										</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">모델 유형:</div>
										<div className="text-sm col-span-2">
											{MODEL_TYPES.find(
												(type) => type.id === selectedJob.model_type
											)?.name ||
												selectedJob.model_type ||
												"-"}
										</div>
									</div>
									<div className="grid grid-cols-3 gap-2">
										<div className="text-sm font-medium">생성일:</div>
										<div className="text-sm col-span-2">
											{selectedJob.created_at}
										</div>
									</div>
									{selectedJob.completed_at && (
										<div className="grid grid-cols-3 gap-2">
											<div className="text-sm font-medium">완료일:</div>
											<div className="text-sm col-span-2">
												{selectedJob.completed_at}
											</div>
										</div>
									)}
									{selectedJob.accuracy && (
										<div className="grid grid-cols-3 gap-2">
											<div className="text-sm font-medium">정확도:</div>
											<div className="text-sm col-span-2">
												{selectedJob.accuracy}
											</div>
										</div>
									)}
								</div>
							) : (
								<div className="flex justify-center items-center h-40 text-muted-foreground">
									좌측에서 연합학습 작업을 선택하세요.
								</div>
							)}
						</CardContent>
					</Card>
				</div>
			)}
		</div>
	);
};

export default FederatedLearningContent;
