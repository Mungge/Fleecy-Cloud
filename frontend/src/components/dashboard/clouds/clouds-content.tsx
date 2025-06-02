"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Trash2 } from "lucide-react";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
	DialogDescription,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Plus } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { useToast } from "@/components/ui/use-toast";
import Cookies from "js-cookie";

interface CloudConnection {
	id: string;
	provider: "AWS" | "GCP";
	name: string;
	region: string;
	zone: string;
	status: "active" | "inactive";
}

// GCP 자격 증명 파일 인터페이스 정의
interface GCPCredentials {
	project_id: string;
	private_key: string;
	client_email: string;
	[key: string]: string | string[];
}

const awsSchema = z.object({
	name: z.string().min(1, "이름을 입력해주세요"),
	region: z.string().min(1, "리전을 입력해주세요"),
	zone: z.string().min(1, "영역을 입력해주세요"),
});

const gcpSchema = z.object({
	name: z.string().min(1, "이름을 입력해주세요"),
	region: z.string().min(1, "리전을 입력해주세요"),
	zone: z.string().min(1, "영역을 입력해주세요"),
});

type CloudFormData = z.infer<typeof awsSchema> | z.infer<typeof gcpSchema>;

const CloudsContent = () => {
	const [clouds, setClouds] = useState<CloudConnection[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [isDialogOpen, setIsDialogOpen] = useState(false);
	const [selectedProvider, setSelectedProvider] = useState<"AWS" | "GCP">(
		"AWS"
	);
	const [credentialFile, setCredentialFile] = useState<File | null>(null);
	const [extractedProjectId, setExtractedProjectId] = useState<string>("");
	const [fileContents, setFileContents] = useState<GCPCredentials | null>(null);
	const [selectedCloud, setSelectedCloud] = useState<CloudConnection | null>(null);
	
	const { toast } = useToast();

	const awsForm = useForm<z.infer<typeof awsSchema>>({
		resolver: zodResolver(awsSchema),
		defaultValues: {
			name: "",
			region: "",
			zone: "",
		},
	});

	const gcpForm = useForm<z.infer<typeof gcpSchema>>({
		resolver: zodResolver(gcpSchema),
		defaultValues: {
			name: "",
			region: "",
			zone: "",
		},
	});

	useEffect(() => {
		fetchClouds();
	}, []);

	const fetchClouds = async () => {
		try {
			setIsLoading(true);
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

			// 토큰 가져오기
			const token = Cookies.get("token");

			const response = await fetch(`${apiUrl}/api/clouds`, {
				headers: {
					Authorization: token ? `Bearer ${token}` : "",
				},
				credentials: "include",
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const result = await response.json();

			if (!result || typeof result !== "object") {
				console.error("API 응답이 객체가 아닙니다:", result);
				setClouds([]);
				return;
			}

			const cloudsData = result.data
				? result.data
				: Array.isArray(result)
				? result
				: [];

			if (!Array.isArray(cloudsData)) {
				console.error("클라우드 데이터가 배열이 아닙니다:", cloudsData);
				setClouds([]);
				return;
			}

			setClouds(cloudsData);
		} catch (error) {
			console.error("Failed to fetch clouds:", error);
			setClouds([]);
		} finally {
			setIsLoading(false);
		}
	};

	const handleFileChange = async (
		event: React.ChangeEvent<HTMLInputElement>
	) => {
		const file = event.target.files?.[0];
		if (!file) return;

		setCredentialFile(file);

		// GCP 자격 증명 파일인 경우 파일 내용 읽기
		if (selectedProvider === "GCP" && file.type === "application/json") {
			try {
				const text = await file.text();
				const jsonData = JSON.parse(text) as GCPCredentials;
				setFileContents(jsonData);

				if (jsonData.project_id) {
					setExtractedProjectId(jsonData.project_id);
				}
			} catch (err) {
				console.error("파일 파싱 실패:", err);
			}
		}
	};

	const onSubmit = async (data: CloudFormData) => {
		try {
			if (!credentialFile) {
				throw new Error("자격 증명 파일이 필요합니다");
			}

			const formData = new FormData();
			formData.append("credentialFile", credentialFile, credentialFile.name);
			formData.append("provider", selectedProvider);
			formData.append("name", data.name);

			if (selectedProvider === "AWS") {
				formData.append("region", (data as z.infer<typeof awsSchema>).region);
				formData.append("zone", (data as z.infer<typeof awsSchema>).zone);
			} else {
				const gcp = data as z.infer<typeof gcpSchema>;

				// 파일에서 추출한 projectId 사용
				const projectId = fileContents?.project_id || "";
				if (!projectId) {
					throw new Error("프로젝트 ID를 추출할 수 없습니다");
				}

				formData.append("projectId", projectId);
				formData.append("region", gcp.region); // 리전 정보
				formData.append("zone", gcp.zone); // 영역 정보
			}

			const formDataObj: Record<string, string> = {};
			for (const pair of formData.entries()) {
				formDataObj[pair[0]] =
					pair[1] instanceof File
						? `(파일: ${pair[1].name})`
						: (pair[1] as string);
			}

			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

			// 토큰 가져오기
			const token = Cookies.get("token");

			// 업로드 시도
			try {
				const response = await fetch(`${apiUrl}/api/clouds/upload`, {
					method: "POST",
					body: formData,
					headers: {
						Authorization: token ? `Bearer ${token}` : "",
					},
					credentials: "include",
				});

				// 응답 텍스트 가져오기
				const responseText = await response.text();

				if (!response.ok) {
					let errorMessage = "자격 증명 파일 업로드 실패";
					try {
						if (responseText && responseText.trim().startsWith("{")) {
							const errorData = JSON.parse(responseText);
							errorMessage = errorData.error || errorMessage;
						}
					} catch (e) {
						console.error("응답 파싱 실패:", e);
					}
					throw new Error(errorMessage);
				}

				setIsDialogOpen(false);
				fetchClouds();
				
				// 성공 메시지 표시
				toast({
					title: "성공",
					description: "클라우드 연결이 추가되었습니다.",
				});
			} catch (fetchError) {
				console.error("Fetch 에러:", fetchError);
				toast({
					title: "오류",
					description: "클라우드 연결 추가에 실패했습니다.",
					variant: "destructive",
				});
				throw fetchError;
			}
		} catch (error) {
			console.error("클라우드 추가 실패:", error);
			toast({
				title: "오류",
				description: "클라우드 연결 추가에 실패했습니다.",
				variant: "destructive",
			});
		}
	};

	const handleDeleteCloud = async (id: string) => {
		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

			// 토큰 가져오기
			const token = Cookies.get("token");

			const response = await fetch(`${apiUrl}/api/clouds/${id}`, {
				method: "DELETE",
				headers: {
					Authorization: token ? `Bearer ${token}` : "",
				},
				credentials: "include",
			});

			if (response.ok) {
				fetchClouds();
				
				// 성공 메시지 표시
				toast({
					title: "성공",
					description: "클라우드 연결이 삭제되었습니다.",
				});
				
				// 선택된 클라우드가 삭제된 경우 선택 해제
				if (selectedCloud?.id === id) {
					setSelectedCloud(null);
				}
			}
		} catch (error) {
			console.error("Failed to delete cloud:", error);
			toast({
				title: "오류",
				description: "클라우드 연결 삭제에 실패했습니다.",
				variant: "destructive",
			});
		}
	};

	// 상태에 따른 배지 색상 설정
	const getStatusBadge = (status: string) => {
		switch (status) {
			case "active":
				return <Badge className="bg-green-500">활성</Badge>;
			case "inactive":
				return <Badge className="bg-gray-500">비활성</Badge>;
			default:
				return <Badge>{status}</Badge>;
		}
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-3xl font-bold tracking-tight">클라우드 인증 정보</h2>
					<p className="text-muted-foreground">
						클라우드 제공자의 인증 정보를 관리하세요.
					</p>
				</div>

				<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
					<DialogTrigger asChild>
						<Button className="ml-auto">
							<Plus className="mr-2 h-4 w-4" />
							인증 정보 추가
						</Button>
					</DialogTrigger>
					<DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
						<DialogHeader>
							<DialogTitle>클라우드 추가</DialogTitle>
							<DialogDescription>
								클라우드 제공자를 선택하고 인증 정보를 업로드해주세요.
							</DialogDescription>
						</DialogHeader>
						<Tabs
							defaultValue="AWS"
							value={selectedProvider}
							onValueChange={(value) =>
								setSelectedProvider(value as "AWS" | "GCP")
							}
						>
							<TabsList className="grid w-full grid-cols-2">
								<TabsTrigger value="AWS">AWS</TabsTrigger>
								<TabsTrigger value="GCP">GCP</TabsTrigger>
							</TabsList>
							<TabsContent value="AWS">
								<Form {...awsForm}>
									<form
										onSubmit={awsForm.handleSubmit(onSubmit)}
										className="space-y-6"
									>
										<FormField
											control={awsForm.control}
											name="name"
											render={({ field }) => (
												<FormItem>
													<FormLabel>이름</FormLabel>
													<FormControl>
														<Input placeholder="AWS 연결 이름" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<FormField
											control={awsForm.control}
											name="region"
											render={({ field }) => (
												<FormItem>
													<FormLabel>리전(Region)</FormLabel>
													<FormControl>
														<Input placeholder="ap-northeast-2" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<FormField
											control={awsForm.control}
											name="zone"
											render={({ field }) => (
												<FormItem>
													<FormLabel>영역(Zone)</FormLabel>
													<FormControl>
														<Input placeholder="ap-northeast-2a" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<div className="space-y-4">
											<div className="text-sm font-medium">
												AWS 자격 증명 파일
											</div>
											<div className="space-y-2">
												<div className="flex items-center space-x-2">
													<Input
														type="file"
														accept=".csv"
														onChange={handleFileChange}
														required
													/>
												</div>
												<div className="text-sm text-muted-foreground">
													AWS 인증 정보 파일(.csv)을 업로드해주세요
												</div>
											</div>
										</div>
										<Button type="submit" className="w-full">
											추가
										</Button>
									</form>
								</Form>
							</TabsContent>
							<TabsContent value="GCP">
								<Form {...gcpForm}>
									<form
										onSubmit={gcpForm.handleSubmit(onSubmit)}
										className="space-y-6"
									>
										<FormField
											control={gcpForm.control}
											name="name"
											render={({ field }) => (
												<FormItem>
													<FormLabel>이름</FormLabel>
													<FormControl>
														<Input placeholder="GCP 연결 이름" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<FormField
											control={gcpForm.control}
											name="region"
											render={({ field }) => (
												<FormItem>
													<FormLabel>리전(Region)</FormLabel>
													<FormControl>
														<Input placeholder="us-central1" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<FormField
											control={gcpForm.control}
											name="zone"
											render={({ field }) => (
												<FormItem>
													<FormLabel>영역(Zone)</FormLabel>
													<FormControl>
														<Input placeholder="us-central1-a" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
										<div className="mt-2 text-sm">
											<div>
												<span className="font-medium">프로젝트 ID:</span>{" "}
												{extractedProjectId || "파일에서 추출됩니다"}
											</div>
										</div>
										<div className="space-y-4">
											<div className="text-sm font-medium">
												서비스 계정 키 파일
											</div>
											<div className="space-y-2">
												<div className="flex items-center space-x-2">
													<Input
														type="file"
														accept=".json"
														onChange={handleFileChange}
														required
													/>
												</div>
												<div className="text-sm text-muted-foreground">
													GCP 인증 정보 파일(.json)을 업로드해주세요
												</div>
											</div>
										</div>
										<Button type="submit" className="w-full">
											추가
										</Button>
									</form>
								</Form>
							</TabsContent>
						</Tabs>
					</DialogContent>
				</Dialog>
			</div>

			{isLoading ? (
				<div className="flex justify-center items-center py-12">
					<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
				</div>
			) : (
				<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
					{/* 클라우드 연결 목록 */}
					<Card className="md:col-span-2">
						<CardHeader>
							<CardTitle>클라우드 연결 목록</CardTitle>
							<CardDescription>
								등록된 클라우드 연결과 상태를 확인하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<ScrollArea className="h-[calc(100vh-320px)]">
								<Table>
									<TableHeader>
										<TableRow>
											<TableHead>이름</TableHead>
											<TableHead>제공자</TableHead>
											<TableHead>리전</TableHead>
											<TableHead>상태</TableHead>
											<TableHead>작업</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{clouds && clouds.length > 0 ? (
											clouds.map((cloud) => (
												<TableRow
													key={cloud.id}
													className={`cursor-pointer hover:bg-muted/50 ${
														selectedCloud?.id === cloud.id ? "bg-muted" : ""
													}`}
													onClick={() => setSelectedCloud(cloud)}
												>
													<TableCell className="font-medium">{cloud.name}</TableCell>
													<TableCell>{cloud.provider}</TableCell>
													<TableCell>{cloud.region}</TableCell>
													<TableCell>{getStatusBadge(cloud.status)}</TableCell>
													<TableCell>
														<div className="flex space-x-2">
															<Button
																variant="ghost"
																size="icon"
																onClick={(e) => {
																	e.stopPropagation();
																	handleDeleteCloud(cloud.id);
																}}
															>
																<Trash2 className="h-4 w-4" />
															</Button>
														</div>
													</TableCell>
												</TableRow>
											))
										) : (
											<TableRow>
												<TableCell colSpan={5} className="text-center py-12 text-muted-foreground">
													<p>등록된 클라우드 연결이 없습니다.</p>
													<p className="text-sm mt-2">
														위의 &apos;인증 정보 추가&apos; 버튼을 클릭하여 새로운 연결을 추가하세요.
													</p>
												</TableCell>
											</TableRow>
										)}
									</TableBody>
								</Table>
							</ScrollArea>
						</CardContent>
					</Card>

					{/* 클라우드 연결 상세 정보 */}
					<Card>
						<CardHeader>
							<CardTitle>연결 상세 정보</CardTitle>
							<CardDescription>
								선택한 클라우드 연결의 상세 정보를 확인하세요.
							</CardDescription>
						</CardHeader>
						<CardContent>
							{selectedCloud ? (
								<div className="space-y-4">
									<div>
										<h3 className="font-semibold text-lg">{selectedCloud.name}</h3>
										<p className="text-sm text-muted-foreground">{selectedCloud.provider} 클라우드</p>
									</div>
									
									<div className="space-y-3">
										<div>
											<span className="text-sm font-medium">제공자:</span>
											<p className="text-sm">{selectedCloud.provider}</p>
										</div>
										
										<div>
											<span className="text-sm font-medium">리전:</span>
											<p className="text-sm">{selectedCloud.region}</p>
										</div>
										
										<div>
											<span className="text-sm font-medium">영역:</span>
											<p className="text-sm">{selectedCloud.zone}</p>
										</div>
										
										<div>
											<span className="text-sm font-medium">상태:</span>
											<div className="mt-1">{getStatusBadge(selectedCloud.status)}</div>
										</div>
									</div>
								</div>
							) : (
								<div className="text-center py-8 text-muted-foreground">
									<p>클라우드 연결을 선택해주세요</p>
									<p className="text-sm mt-2">
										왼쪽 목록에서 클라우드 연결을 클릭하면 상세 정보를 확인할 수 있습니다.
									</p>
								</div>
							)}
						</CardContent>
					</Card>
				</div>
			)}
		</div>
	);
};

export default CloudsContent;
