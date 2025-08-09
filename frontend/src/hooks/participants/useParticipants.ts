import { useState, useEffect } from "react";
import { toast } from "sonner";
import {
	createParticipant,
	getParticipants,
	updateParticipant,
	deleteParticipant,
	healthCheckVM,
} from "@/api/participants";
import { Participant } from "@/types/participant";

export function useParticipants() {
	const [participants, setParticipants] = useState<Participant[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [selectedParticipant, setSelectedParticipant] =
		useState<Participant | null>(null);

	// 클러스터 목록 로드
	const loadParticipants = async () => {
		try {
			setIsLoading(true);
			const data = await getParticipants();
			setParticipants(data);
		} catch (error) {
			console.error("클러스터 목록 로드 실패:", error);
			toast.error("클러스터 목록을 불러오는데 실패했습니다.");
		} finally {
			setIsLoading(false);
		}
	};

	// 클러스터 생성
	const handleCreateParticipant = async (formData: FormData) => {
		try {
			await createParticipant(formData);
			toast.success("클러스터가 성공적으로 추가되었습니다.");
			await loadParticipants();
			return true;
		} catch (error) {
			console.error("참여자 생성 실패:", error);
			toast.error("클러스터 추가에 실패했습니다.");
			return false;
		}
	};

	// 참여자 수정
	const handleUpdateParticipant = async (id: string, formData: FormData) => {
		try {
			await updateParticipant(id, formData);
			toast.success("클러스터 정보가 성공적으로 수정되었습니다.");
			await loadParticipants();
			return true;
		} catch (error) {
			console.error("참여자 수정 실패:", error);
			toast.error("클러스터 수정에 실패했습니다.");
			return false;
		}
	};

	// 참여자 삭제
	const handleDeleteParticipant = async (id: string) => {
		try {
			await deleteParticipant(id);
			toast.success("참여자가 성공적으로 삭제되었습니다.");
			// 삭제된 참여자가 선택되어 있었다면 선택 해제
			if (selectedParticipant?.id === id) {
				setSelectedParticipant(null);
			}
			await loadParticipants();
		} catch (error) {
			console.error("참여자 삭제 실패:", error);
			toast.error("참여자 삭제에 실패했습니다.");
		}
	};

	// VM 헬스체크
	const handleHealthCheck = async (participant: Participant) => {
		try {
			const healthResult = await healthCheckVM(participant.id);

			// 상세한 헬스체크 결과 표시
			if (healthResult.healthy) {
				toast.success(
					`✅ ${participant.name} 헬스체크 성공\n상태: ${healthResult.status}\n응답시간: ${healthResult.response_time_ms}ms\n${healthResult.message}`,
					{
						duration: 5000,
					}
				);
			} else {
				toast.error(
					`❌ ${participant.name} 헬스체크 실패\n상태: ${healthResult.status}\n응답시간: ${healthResult.response_time_ms}ms\n${healthResult.message}`,
					{
						duration: 8000,
					}
				);
			} // 헬스체크 완료 후 참여자 목록 새로고침으로 UI 상태 동기화
			await loadParticipants();
		} catch (error) {
			console.error("헬스체크 실패:", error);
			toast.error(
				`🚨 ${participant.name} 헬스체크 오류\n${
					error instanceof Error ? error.message : String(error)
				}`
			);

			// 오류 발생 시에도 참여자 목록 새로고침
			await loadParticipants();
		}
	};

	useEffect(() => {
		loadParticipants();
	}, []);

	return {
		participants,
		isLoading,
		selectedParticipant,
		setSelectedParticipant,
		loadParticipants,
		handleCreateParticipant,
		handleUpdateParticipant,
		handleDeleteParticipant,
		handleHealthCheck,
	};
}
