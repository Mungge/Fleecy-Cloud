"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { PlusCircle, Trash2, AlertCircle } from "lucide-react";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface CloudConnection {
	id: string;
	provider: "AWS" | "GCP";
	name: string;
	region: string;
	status: "connected" | "disconnected";
}

export default function CloudsPage() {
	const [clouds, setClouds] = useState<CloudConnection[]>([]);
	const [isAddingCloud, setIsAddingCloud] = useState(false);
	const [newCloud, setNewCloud] = useState({
		provider: "",
		name: "",
		region: "",
		accessKey: "",
		secretKey: "",
	});
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		fetchClouds();
	}, []);

	const fetchClouds = async () => {
		setIsLoading(true);
		setError(null);

		try {
			// 백엔드 API URL 설정
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			console.log(`API URL: ${apiUrl}/api/clouds`);

			const response = await fetch(`${apiUrl}/api/clouds`, {
				method: "GET",
				headers: {
					Accept: "application/json",
				},
				credentials: "include",
			});

			console.log("API 응답 상태:", response.status);

			// 응답 텍스트 먼저 확인
			const responseText = await response.text();
			console.log("API 응답 내용 미리보기:", responseText.substring(0, 100));

			// 빈 응답 처리
			if (!responseText.trim()) {
				console.warn("빈 응답이 반환되었습니다");
				setClouds([]);
				setIsLoading(false);
				return;
			}

			// JSON 파싱 시도
			try {
				const data = JSON.parse(responseText);
				setClouds(Array.isArray(data) ? data : []);
			} catch (parseError) {
				console.error("JSON 파싱 오류:", parseError);
				setError(
					"서버 응답을 파싱할 수 없습니다. 백엔드 서버가 올바른 JSON을 반환하는지 확인하세요."
				);
				setClouds([]);
			}
		} catch (error) {
			console.error("클라우드 데이터 가져오기 실패:", error);
			setError(
				"클라우드 연결 정보를 가져오는 데 실패했습니다. 백엔드 서버가 실행 중인지 확인하세요."
			);
			setClouds([]);
		} finally {
			setIsLoading(false);
		}
	};

	const handleAddCloud = async () => {
		try {
			// 기본 유효성 검사
			if (!newCloud.name || !newCloud.provider || !newCloud.region) {
				setError("모든 필수 필드를 입력해주세요.");
				return;
			}

			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			console.log(`클라우드 추가 요청: ${apiUrl}/api/clouds`);

			const response = await fetch(`${apiUrl}/api/clouds`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					Accept: "application/json",
				},
				credentials: "include",
				body: JSON.stringify(newCloud),
			});

			console.log("API 응답 상태:", response.status);

			// 응답이 JSON이 아닐 수 있으므로 텍스트로 먼저 읽음
			const responseText = await response.text();

			if (response.ok) {
				setIsAddingCloud(false);
				fetchClouds();
				setNewCloud({
					provider: "",
					name: "",
					region: "",
					accessKey: "",
					secretKey: "",
				});
				setError(null);
			} else {
				let errorMessage = "클라우드 추가에 실패했습니다.";
				try {
					const errorData = JSON.parse(responseText);
					errorMessage = errorData.error || errorMessage;
				} catch {
					// 파싱 실패 시 기본 메시지 사용
				}
				setError(errorMessage);
			}
		} catch (error) {
			console.error("클라우드 추가 실패:", error);
			setError("클라우드 추가 중 오류가 발생했습니다.");
		}
	};

	const handleDeleteCloud = async (id: string) => {
		try {
			const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
			console.log(`클라우드 삭제 요청: ${apiUrl}/api/clouds/${id}`);

			const response = await fetch(`${apiUrl}/api/clouds/${id}`, {
				method: "DELETE",
				headers: {
					Accept: "application/json",
				},
				credentials: "include",
			});

			console.log("API 응답 상태:", response.status);

			if (response.ok) {
				setError(null);
				fetchClouds();
			} else {
				const responseText = await response.text();
				let errorMessage = "클라우드 삭제에 실패했습니다.";
				try {
					const errorData = JSON.parse(responseText);
					errorMessage = errorData.error || errorMessage;
				} catch {
					// 파싱 실패 시 기본 메시지 사용
				}
				setError(errorMessage);
			}
		} catch (error) {
			console.error("클라우드 삭제 실패:", error);
			setError("클라우드 삭제 중 오류가 발생했습니다.");
		}
	};

	return (
		<div className="container mx-auto p-6">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-3xl font-bold">클라우드 연결 관리</h1>
				<Dialog open={isAddingCloud} onOpenChange={setIsAddingCloud}>
					<DialogTrigger asChild>
						<Button>
							<PlusCircle className="mr-2 h-4 w-4" />
							클라우드 추가
						</Button>
					</DialogTrigger>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>새 클라우드 연결 추가</DialogTitle>
						</DialogHeader>
						<div className="grid gap-4 py-4">
							<div className="grid gap-2">
								<Label htmlFor="provider">클라우드 제공자</Label>
								<Select
									value={newCloud.provider}
									onValueChange={(value) =>
										setNewCloud({ ...newCloud, provider: value })
									}
								>
									<SelectTrigger>
										<SelectValue placeholder="선택하세요" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="AWS">AWS</SelectItem>
										<SelectItem value="GCP">GCP</SelectItem>
									</SelectContent>
								</Select>
							</div>
							<div className="grid gap-2">
								<Label htmlFor="name">연결 이름</Label>
								<Input
									id="name"
									value={newCloud.name}
									onChange={(e) =>
										setNewCloud({ ...newCloud, name: e.target.value })
									}
								/>
							</div>
							<div className="grid gap-2">
								<Label htmlFor="region">리전</Label>
								<Input
									id="region"
									value={newCloud.region}
									onChange={(e) =>
										setNewCloud({ ...newCloud, region: e.target.value })
									}
								/>
							</div>
							<div className="grid gap-2">
								<Label htmlFor="accessKey">액세스 키</Label>
								<Input
									id="accessKey"
									type="password"
									value={newCloud.accessKey}
									onChange={(e) =>
										setNewCloud({ ...newCloud, accessKey: e.target.value })
									}
								/>
							</div>
							<div className="grid gap-2">
								<Label htmlFor="secretKey">시크릿 키</Label>
								<Input
									id="secretKey"
									type="password"
									value={newCloud.secretKey}
									onChange={(e) =>
										setNewCloud({ ...newCloud, secretKey: e.target.value })
									}
								/>
							</div>
						</div>
						<div className="flex justify-end">
							<Button onClick={handleAddCloud}>연결 추가</Button>
						</div>
					</DialogContent>
				</Dialog>
			</div>

			{error && (
				<Alert variant="destructive" className="mb-6">
					<AlertCircle className="h-4 w-4" />
					<AlertTitle>오류</AlertTitle>
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			)}

			{isLoading ? (
				<div className="flex justify-center items-center h-40">
					<p className="text-gray-500">로딩 중...</p>
				</div>
			) : clouds.length === 0 ? (
				<div className="bg-gray-50 rounded-lg p-10 text-center">
					<h3 className="text-xl font-medium mb-2">
						등록된 클라우드가 없습니다
					</h3>
					<p className="text-gray-500 mb-6">
						&apos;클라우드 추가&apos; 버튼을 클릭하여 첫 번째 클라우드 연결을
						추가하세요.
					</p>
					<Button onClick={() => setIsAddingCloud(true)}>
						<PlusCircle className="mr-2 h-4 w-4" />
						클라우드 추가
					</Button>
				</div>
			) : (
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
			)}
		</div>
	);
}
