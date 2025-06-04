"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { PlusCircle, Trash2 } from "lucide-react";
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
import { toast } from "sonner";

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
	const [isLoading, setIsLoading] = useState(false);
	const [newCloud, setNewCloud] = useState({
		provider: "",
		name: "",
		region: "",
		accessKey: "",
		secretKey: "",
	});

	useEffect(() => {
		fetchClouds();
	}, []);

	const fetchClouds = async () => {
		try {
			setIsLoading(true);
			const response = await fetch("/api/clouds");
			if (!response.ok) {
				throw new Error("클라우드 목록을 불러오는데 실패했습니다");
			}
			const data = await response.json();
			setClouds(data);
		} catch (error) {
			console.error("Failed to fetch clouds:", error);
			toast.error("클라우드 목록을 불러오는데 실패했습니다");
		} finally {
			setIsLoading(false);
		}
	};

	const validateForm = () => {
		if (!newCloud.provider) {
			toast.error("클라우드 제공자를 선택해주세요");
			return false;
		}
		if (!newCloud.name.trim()) {
			toast.error("연결 이름을 입력해주세요");
			return false;
		}
		if (!newCloud.region.trim()) {
			toast.error("리전을 입력해주세요");
			return false;
		}
		if (!newCloud.accessKey.trim()) {
			toast.error("액세스 키를 입력해주세요");
			return false;
		}
		if (!newCloud.secretKey.trim()) {
			toast.error("시크릿 키를 입력해주세요");
			return false;
		}
		return true;
	};

	const handleAddCloud = async () => {
		if (!validateForm()) return;

		try {
			setIsLoading(true);
			const response = await fetch("/api/clouds", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(newCloud),
			});

			if (response.ok) {
				toast.success(`${newCloud.name} 클라우드 연결이 성공적으로 추가되었습니다`);
				setIsAddingCloud(false);
				fetchClouds();
				setNewCloud({
					provider: "",
					name: "",
					region: "",
					accessKey: "",
					secretKey: "",
				});
			} else {
				const errorData = await response.json();
				throw new Error(errorData.message || "클라우드 연결 추가에 실패했습니다");
			}
		} catch (error) {
			console.error("Failed to add cloud:", error);
			toast.error(error instanceof Error ? error.message : "클라우드 연결 추가에 실패했습니다");
		} finally {
			setIsLoading(false);
		}
	};

	const handleDeleteCloud = async (id: string, cloudName: string) => {
		// 삭제 확인
		const confirmed = window.confirm(`정말로 "${cloudName}" 클라우드 연결을 삭제하시겠습니까?`);
		if (!confirmed) return;

		try {
			setIsLoading(true);
			const response = await fetch(`/api/clouds/${id}`, {
				method: "DELETE",
			});

			if (response.ok) {
				toast.success(`${cloudName} 클라우드 연결이 성공적으로 삭제되었습니다`);
				fetchClouds();
			} else {
				const errorData = await response.json();
				throw new Error(errorData.message || "클라우드 연결 삭제에 실패했습니다");
			}
		} catch (error) {
			console.error("Failed to delete cloud:", error);
			toast.error(error instanceof Error ? error.message : "클라우드 연결 삭제에 실패했습니다");
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<div className="container mx-auto p-6">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-3xl font-bold">클라우드 연결</h1>
				<Dialog open={isAddingCloud} onOpenChange={setIsAddingCloud}>
					<DialogTrigger asChild>
						<Button disabled={isLoading}>
							<PlusCircle className="mr-2 h-4 w-4" />
							새 클라우드 추가
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
						<div className="flex justify-end space-x-2">
							<Button
								variant="outline"
								onClick={() => setIsAddingCloud(false)}
								disabled={isLoading}
							>
								취소
							</Button>
							<Button onClick={handleAddCloud} disabled={isLoading}>
								{isLoading ? "추가 중..." : "추가"}
							</Button>
						</div>
					</DialogContent>
				</Dialog>
			</div>

			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{isLoading ? (
					<div className="col-span-full text-center py-8 text-gray-500">
						클라우드 목록을 불러오는 중...
					</div>
				) : clouds.length === 0 ? (
					<div className="col-span-full text-center py-8 text-gray-500">
						연결된 클라우드가 없습니다. 새 클라우드를 추가해보세요.
					</div>
				) : (
					clouds.map((cloud) => (
						<Card key={cloud.id}>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-lg font-semibold">
									{cloud.name}
								</CardTitle>
								<Button
									variant="ghost"
									size="sm"
									onClick={() => handleDeleteCloud(cloud.id, cloud.name)}
									disabled={isLoading}
									className="text-red-500 hover:text-red-700"
								>
									<Trash2 className="h-4 w-4" />
								</Button>
							</CardHeader>
							<CardContent>
								<div className="space-y-2">
									<div className="flex items-center justify-between">
										<span className="text-sm text-gray-500">제공자:</span>
										<span className="text-sm font-medium">{cloud.provider}</span>
									</div>
									<div className="flex items-center justify-between">
										<span className="text-sm text-gray-500">리전:</span>
										<span className="text-sm font-medium">{cloud.region}</span>
									</div>
									<div className="flex items-center justify-between">
										<span className="text-sm text-gray-500">상태:</span>
										<span
											className={`text-sm font-medium ${
												cloud.status === "connected"
													? "text-green-600"
													: "text-red-600"
											}`}
										>
											{cloud.status === "connected" ? "연결됨" : "연결 끊김"}
										</span>
									</div>
								</div>
							</CardContent>
						</Card>
					))
				)}
			</div>
		</div>
	);
}
