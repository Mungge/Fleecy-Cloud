"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
			} catch (fetchError) {
				console.error("Fetch 에러:", fetchError);
				throw fetchError;
			}
		} catch (error) {
			console.error("클라우드 추가 실패:", error);
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
			}
		} catch (error) {
			console.error("Failed to delete cloud:", error);
		}
	};

	return (
		<div className="container mx-auto p-6">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-3xl font-bold">클라우드 인증 정보 관리</h1>
				<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
					<DialogTrigger asChild>
						<Button>
							<Plus className="mr-2 h-4 w-4" />
							인증 정보 추가
						</Button>
					</DialogTrigger>
					<DialogContent className="sm:max-w-[425px]">
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
										className="space-y-4"
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
										className="space-y-4"
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
			) : clouds && clouds.length > 0 ? (
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{clouds.map((cloud) => (
						<Card key={cloud.id}>
							<CardHeader>
								<CardTitle className="font-bold text-xl flex justify-between items-center">
									<span>{cloud.name}</span>
									<Button
										variant="ghost"
										size="icon"
										onClick={() => handleDeleteCloud(cloud.id)}
									>
										<Trash2 className="h-4 w-4" />
									</Button>
								</CardTitle>
							</CardHeader>
							<CardContent>
								<div className="space-y-2">
									<p>
										<span className="font-semibold">CSP:</span> {cloud.provider}
									</p>
									<p>
										<span className="font-semibold">Region:</span>{" "}
										{cloud.region}
									</p>
									<p>
										<span className="font-semibold">Zone:</span> {cloud.zone}
									</p>
									<p>
										<span className="font-semibold">Status:</span>{" "}
										<span
											className={`inline-block px-2 py-1 rounded-full text-xs ${
												cloud.status === "active"
													? "bg-green-100 text-green-800"
													: "bg-red-100 text-red-800"
											}`}
										>
											{cloud.status === "active" ? "active" : "inactive"}
										</span>
									</p>
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			) : (
				<div className="text-center py-12 text-muted-foreground">
					<p>등록된 클라우드 연결이 없습니다.</p>
					<p className="text-sm mt-2">
						위의 &apos;클라우드 추가&apos; 버튼을 클릭하여 새로운 연결을
						추가하세요.
					</p>
				</div>
			)}
		</div>
	);
};

export default CloudsContent;
