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

	useEffect(() => {
		// TODO: API 호출로 클라우드 연결 목록 가져오기
		fetchClouds();
	}, []);

	const fetchClouds = async () => {
		try {
			const response = await fetch("/api/clouds");
			const data = await response.json();
			setClouds(data);
		} catch (error) {
			console.error("Failed to fetch clouds:", error);
		}
	};

	const handleAddCloud = async () => {
		try {
			const response = await fetch("/api/clouds", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(newCloud),
			});

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
			}
		} catch (error) {
			console.error("Failed to add cloud:", error);
		}
	};

	const handleDeleteCloud = async (id: string) => {
		try {
			const response = await fetch(`/api/clouds/${id}`, {
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
		</div>
	);
}
