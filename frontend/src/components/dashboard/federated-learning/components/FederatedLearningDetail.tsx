// components/dashboard/federated-learning/components/FederatedLearningDetail.tsx
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FileText, Settings } from "lucide-react";
import { StatusBadge } from "./StatusBadge";
import { FederatedLearningLogsDialog } from "./FederatedLearningLogsDialog";
import { GlobalModelManagementDialog } from "./GlobalModelManagementDialog";
import { FederatedLearningJob } from "@/types/federated-learning";
import { AGGREGATION_ALGORITHMS, MODEL_TYPES } from "../constants";


interface FederatedLearningDetailProps {
	selectedJob: FederatedLearningJob | null;
}

export const FederatedLearningDetail = ({
	selectedJob,
}: FederatedLearningDetailProps) => {
	return (
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
								<StatusBadge status={selectedJob.status} />
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">참여 클러스터:</div>
							<div className="text-sm col-span-2">
								{selectedJob.participant_count !== undefined &&
								selectedJob.participant_count !== null
									? selectedJob.participant_count
									: selectedJob.participants?.length ?? 0}
								개
							</div>
						</div>
						{selectedJob.participants &&
							Array.isArray(selectedJob.participants) &&
							selectedJob.participants.length > 0 && (
								<div className="space-y-2">
									<div className="text-sm font-medium">참여 클러스터 목록:</div>
									<div className="space-y-1">
										{selectedJob.participants.map((participant) => (
											<div
												key={participant.id}
												className="flex items-center justify-between p-2 bg-gray-50 rounded text-sm"
											>
												<span>{participant.name}</span>
												<span
													className={`px-2 py-1 rounded text-xs ${
														participant.status === "active"
															? "bg-green-100 text-green-800"
															: "bg-gray-100 text-gray-600"
													}`}
												>
													{participant.status === "active" ? "활성" : "비활성"}
												</span>
											</div>
										))}
									</div>
								</div>
							)}
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">라운드 수:</div>
							<div className="text-sm col-span-2">
								{selectedJob.rounds || "-"}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">집계 알고리즘:</div>
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
								{MODEL_TYPES.find((type) => type.id === selectedJob.model_type)
									?.name ||
									selectedJob.model_type ||
									"-"}
							</div>
						</div>
						<div className="grid grid-cols-3 gap-2">
							<div className="text-sm font-medium">생성일:</div>
							<div className="text-sm col-span-2">{selectedJob.created_at}</div>
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
								<div className="text-sm col-span-2">{selectedJob.accuracy}</div>
							</div>
						)}
					</div>
				) : (
					<div className="flex justify-center items-center h-40 text-muted-foreground">
						좌측에서 연합학습 작업을 선택하세요.
					</div>
				)}

				{selectedJob && (
					<div className="mt-6 pt-4 border-t">
						<FederatedLearningLogsDialog job={selectedJob}>
							<Button className="w-full" variant="outline">
								<FileText className="h-4 w-4 mr-2" />
								로그 확인
							</Button>
						</FederatedLearningLogsDialog>

						<GlobalModelManagementDialog job={selectedJob}>
							<Button className="w-full" variant="outline">
								<Settings className="h-4 w-4 mr-2" />
								글로벌 모델 관리
							</Button>
						</GlobalModelManagementDialog>
					</div>
				)}
			</CardContent>
		</Card>
	);
};
