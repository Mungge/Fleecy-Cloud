// components/dashboard/federated-learning/components/CreateJobDialog.tsx
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { Slider } from "@/components/ui/slider";
import { Participant } from "@/types/participant";
import {
	AGGREGATION_ALGORITHMS,
	MODEL_TYPES,
	SUPPORTED_FILE_FORMATS,
} from "../constants";
import { CreateJobFormHookReturn } from "../types";
import { toast } from "sonner";
import { useState } from "react";
import { analyzeModelDefinition, formatModelSize, getDefaultModelSize, ModelAnalysis } from "../../aggregator/utils/modelDefinitionParser";

interface CreateJobDialogProps {
	participants: Participant[];
	formHook: CreateJobFormHookReturn;
}

export const CreateJobDialog = ({
	participants,
	formHook,
}: CreateJobDialogProps) => {
	const {
		form,
		modelFile,
		setModelFile,
		isDialogOpen,
		openDialog,
		closeDialog,
		handleSubmit,
	} = formHook;

	const [modelAnalysis, setModelAnalysis] = useState<ModelAnalysis | null>(null);
	const [isAnalyzing, setIsAnalyzing] = useState(false);

	const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files && e.target.files[0];
		if (!file) return;

		const isPy = file.name.toLowerCase().endsWith(".py");
		if (!isPy) {
			toast.error(".py 형식의 모델 구조 정의 파일만 업로드할 수 있습니다.");
			e.currentTarget.value = "";
			setModelFile(null);
			setModelAnalysis(null);
			return;
		}

		setModelFile(file);
		setIsAnalyzing(true);

		try {
			// 모델 정의 파일 분석
			const analysis = await analyzeModelDefinition(file);
			setModelAnalysis(analysis);
			
			// sessionStorage에 모델 크기 정보 저장 (나중에 집계자 최적화에서 사용)
			sessionStorage.setItem("modelAnalysis", JSON.stringify(analysis));
			
			// 기존 modelFileSize도 업데이트하여 호환성 유지
			sessionStorage.setItem("modelFileSize", analysis.modelSizeBytes.toString());
			
			toast.success(
				`모델 분석 완료: ${formatModelSize(analysis.totalParams)} (${analysis.framework})`
			);
		} catch (error) {
			console.error("모델 분석 실패:", error);
			
			// 분석 실패 시 모델 타입 기반 기본값 사용
			const modelType = form.getValues("modelType");
			const defaultParams = getDefaultModelSize(modelType);
			const fallbackAnalysis: ModelAnalysis = {
				totalParams: defaultParams,
				modelSizeBytes: defaultParams * 4,
				layerInfo: [],
				framework: 'unknown'
			};
			
			setModelAnalysis(fallbackAnalysis);
			sessionStorage.setItem("modelAnalysis", JSON.stringify(fallbackAnalysis));
			
			toast.warning(
				`모델 분석에 실패했습니다. 기본값을 사용합니다: ${formatModelSize(defaultParams)}`
			);
		} finally {
			setIsAnalyzing(false);
		}
	};

	return (
		<Dialog
			open={isDialogOpen}
			onOpenChange={(open) => {
				if (!open) {
					closeDialog();
				}
			}}
		>
			<DialogTrigger asChild>
				<Button className="ml-auto" onClick={openDialog}>
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

				{/* Progress Steps */}
				<div className="w-full py-4">
					<div className="flex items-center justify-between">
						{/* Step 1: 정보 입력 */}
						<div className="flex flex-col items-center">
							<div className="flex items-center justify-center w-8 h-8 rounded-full bg-blue-500 text-white text-sm font-medium">
								1
							</div>
							<span className="mt-2 text-sm font-medium text-blue-600">
								정보 입력
							</span>
						</div>

						{/* Connector Line */}
						<div className="flex-1 h-0.5 bg-gray-200 mx-4"></div>

						{/* Step 2: 집계자 생성 */}
						<div className="flex flex-col items-center">
							<div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-200 text-gray-400 text-sm font-medium">
								2
							</div>
							<span className="mt-2 text-sm text-gray-400">집계자 생성</span>
						</div>

						{/* Connector Line */}
						<div className="flex-1 h-0.5 bg-gray-200 mx-4"></div>

						{/* Step 3: 연합학습 생성 */}
						<div className="flex flex-col items-center">
							<div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-200 text-gray-400 text-sm font-medium">
								3
							</div>
							<span className="mt-2 text-sm text-gray-400">연합학습 생성</span>
						</div>
					</div>
				</div>

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(handleSubmit)}
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
												<Input placeholder="연합학습 작업 이름" {...field} />
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
									<Label>모델 정의 파일 업로드 (.py)</Label>
									<div className="grid w-full max-w-sm items-center gap-1.5">
										<Input
											type="file"
											accept={SUPPORTED_FILE_FORMATS}
											onChange={handleFileChange}
											disabled={isAnalyzing}
										/>
										<p className="text-sm text-muted-foreground">
											지원 형식: .py (PyTorch 또는 TensorFlow 모델 정의)
										</p>
									</div>
									
									{isAnalyzing && (
										<div className="text-sm text-blue-600 flex items-center gap-2">
											<div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-blue-600"></div>
											모델 분석 중...
										</div>
									)}
									
									{modelFile && (
										<div className="text-sm space-y-1">
											<div>선택된 파일: {modelFile.name} ({Math.round(modelFile.size / 1024)} KB)</div>
											{modelAnalysis && (
												<div className="bg-gray-50 p-3 rounded-md space-y-1">
													<div className="font-medium text-green-600">
														✓ 모델 분석 완료
													</div>
													<div>프레임워크: {modelAnalysis.framework}</div>
													<div>예상 파라미터 수: {formatModelSize(modelAnalysis.totalParams)}</div>
													<div>예상 모델 크기: {Math.round(modelAnalysis.modelSizeBytes / (1024 * 1024))} MB</div>
													{modelAnalysis.layerInfo.length > 0 && (
														<details className="text-xs">
															<summary className="cursor-pointer">레이어 정보 보기</summary>
															<div className="mt-1 space-y-1">
																{modelAnalysis.layerInfo.slice(0, 5).map((layer, idx) => (
																	<div key={idx}>
																		• {layer.name}: {layer.params.toLocaleString()} params
																	</div>
																))}
																{modelAnalysis.layerInfo.length > 5 && (
																	<div>... 및 {modelAnalysis.layerInfo.length - 5}개 더</div>
																)}
															</div>
														</details>
													)}
												</div>
											)}
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
												<FormLabel className="text-base">참여자 선택</FormLabel>
												<FormDescription>
													연합학습에 참여할 참여자를 선택하세요. 최소 1개 이상의
													참여자가 필요합니다.
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
																								(value: string) =>
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
							<Button type="submit" disabled={isAnalyzing}>
								다음: 집계자 생성
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
};