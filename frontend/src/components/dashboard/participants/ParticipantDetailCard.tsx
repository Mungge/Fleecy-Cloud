import { Edit, CheckCircle, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Participant } from "@/types/participant";
import { getStatusBadge } from "./utils";

interface ParticipantDetailCardProps {
	selectedParticipant: Participant | null;
	onEditParticipant: (participant: Participant) => void;
	onViewVMs: (participant: Participant) => void;
	onHealthCheck: (participant: Participant) => void;
}

export function ParticipantDetailCard({
	selectedParticipant,
	onEditParticipant,
	onViewVMs,
	onHealthCheck,
}: ParticipantDetailCardProps) {
	return (
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
						{selectedParticipant.region && (
							<div>
								<span className="text-sm font-medium">리전:</span>
								<p className="text-sm">{selectedParticipant.region}</p>
							</div>
						)}
						<div>
							<span className="text-sm font-medium">생성일:</span>
							<p className="text-sm">
								{new Date(selectedParticipant.created_at).toLocaleString()}
							</p>
						</div>
						{selectedParticipant.metadata && (
							<div>
								<span className="text-sm font-medium">메타데이터:</span>
								<p className="text-sm">{selectedParticipant.metadata}</p>
							</div>
						)}
						<div>
							<span className="text-sm font-medium">Cluster Endpoint:</span>
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
									onClick={() => onEditParticipant(selectedParticipant)}
									className="justify-start"
								>
									<Edit className="h-4 w-4 mr-2" />
									편집
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={() => onViewVMs(selectedParticipant)}
									className="justify-start"
								>
									<Server className="h-4 w-4 mr-2" />
									가상머신 목록
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={() => onHealthCheck(selectedParticipant)}
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
							왼쪽 목록에서 클러스터를 클릭하면 상세 정보를 확인할 수 있습니다.
						</p>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
