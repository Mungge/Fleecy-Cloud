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

interface CloudConnection {
	id: string;
	provider: "AWS" | "GCP";
	name: string;
	region: string;
	status: "connected" | "disconnected";
}

const awsSchema = z.object({
	name: z.string().min(1, "이름을 입력해주세요"),
	region: z.string().min(1, "리전을 입력해주세요"),
});

const gcpSchema = z.object({
	name: z.string().min(1, "이름을 입력해주세요"),
	projectId: z.string().min(1, "프로젝트 ID를 입력해주세요"),
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

	const awsForm = useForm<z.infer<typeof awsSchema>>({
		resolver: zodResolver(awsSchema),
		defaultValues: {
			name: "",
			region: "",
		},
	});

	const gcpForm = useForm<z.infer<typeof gcpSchema>>({
		resolver: zodResolver(gcpSchema),
		defaultValues: {
			name: "",
			projectId: "",
		},
	});

	useEffect(() => {
		fetchClouds();
	}, []);

	const fetchClouds = async () => {
		try {
			setIsLoading(true);
			const response = await fetch("http://localhost:8080/api/clouds");

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const result = await response.json();
			console.log("API 응답:", result); // 디버깅용 로그

			if (!result || typeof result !== "object") {
				console.error("API 응답이 객체가 아닙니다:", result);
				setClouds([]);
				return;
			}

			if (!result.data) {
				console.error("API 응답에 data 필드가 없습니다:", result);
				setClouds([]);
				return;
			}

			if (!Array.isArray(result.data)) {
				console.error("API 응답의 data 필드가 배열이 아닙니다:", result.data);
				setClouds([]);
				return;
			}

			setClouds(result.data);
		} catch (error) {
			console.error("Failed to fetch clouds:", error);
			setClouds([]);
		} finally {
			setIsLoading(false);
		}
	};

	const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const file = event.target.files?.[0];
		if (file) {
			setCredentialFile(file);
		}
	};

	const onSubmit = async (data: CloudFormData) => {
		try {
			if (!credentialFile) {
				throw new Error("자격 증명 파일이 필요합니다");
			}

			const formData = new FormData();
			formData.append("file", credentialFile);
			formData.append("provider", selectedProvider);
			formData.append("name", data.name);

			if (selectedProvider === "AWS") {
				formData.append("region", (data as z.infer<typeof awsSchema>).region);
			} else {
				formData.append(
					"projectId",
					(data as z.infer<typeof gcpSchema>).projectId
				);
			}

			const response = await fetch("http://localhost:8080/api/clouds/upload", {
				method: "POST",
				body: formData,
			});

			if (!response.ok) {
				throw new Error("자격 증명 파일 업로드 실패");
			}

			setIsDialogOpen(false);
			fetchClouds();
		} catch (error) {
			console.error("클라우드 추가 실패:", error);
		}
	};

	const handleDeleteCloud = async (id: string) => {
		try {
			const response = await fetch(`http://localhost:8080/api/clouds/${id}`, {
				method: "DELETE",
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
							클라우드 추가
						</Button>
					</DialogTrigger>
					<DialogContent className="sm:max-w-[425px]">
						<DialogHeader>
							<DialogTitle>클라우드 추가</DialogTitle>
							<DialogDescription>
								클라우드 제공자를 선택하고 자격 증명 파일을 업로드해주세요.
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
													<FormLabel>리전</FormLabel>
													<FormControl>
														<Input placeholder="ap-northeast-2" {...field} />
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
													AWS 자격 증명 파일(.csv)을 업로드해주세요
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
											name="projectId"
											render={({ field }) => (
												<FormItem>
													<FormLabel>프로젝트 ID</FormLabel>
													<FormControl>
														<Input placeholder="your-project-id" {...field} />
													</FormControl>
													<FormMessage />
												</FormItem>
											)}
										/>
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
													GCP 서비스 계정 키 파일(JSON)을 업로드해주세요
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
								<CardTitle className="flex justify-between items-center">
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
										<span className="font-semibold">제공자:</span>{" "}
										{cloud.provider}
									</p>
									<p>
										<span className="font-semibold">리전:</span> {cloud.region}
									</p>
									<p>
										<span className="font-semibold">상태:</span>{" "}
										<span
											className={`inline-block px-2 py-1 rounded-full text-xs ${
												cloud.status === "connected"
													? "bg-green-100 text-green-800"
													: "bg-red-100 text-red-800"
											}`}
										>
											{cloud.status === "connected" ? "연결됨" : "연결 끊김"}
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
