import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Participant } from "@/types/participant";

// 폼 스키마 정의
const participantSchema = z.object({
	name: z.string().min(1, "이름은 필수입니다"),
	region: z.string().min(1, "리전은 필수입니다"),
	metadata: z.string().optional(),
});

export type ParticipantFormData = z.infer<typeof participantSchema>;

export function useParticipantForm() {
	const [createDialogOpen, setCreateDialogOpen] = useState(false);
	const [editDialogOpen, setEditDialogOpen] = useState(false);
	const [configFile, setConfigFile] = useState<File | null>(null);

	const form = useForm<ParticipantFormData>({
		resolver: zodResolver(participantSchema),
		defaultValues: {
			name: "",
			region: "",
			metadata: "",
		},
	});

	// YAML 파일 업로드 처리
	const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];

			// YAML 파일 확장자 검증
			if (
				!file.name.toLowerCase().endsWith(".yaml") &&
				!file.name.toLowerCase().endsWith(".yml")
			) {
				toast.error("YAML 파일만 업로드 가능합니다.");
				return;
			}

			setConfigFile(file);
		}
	};

	// 생성 다이얼로그 열기
	const openCreateDialog = () => {
		form.reset({
			name: "",
			metadata: "",
		});
		setConfigFile(null);
		setCreateDialogOpen(true);
	};

	// 편집 다이얼로그 열기
	const openEditDialog = (participant: Participant) => {
		form.reset({
			name: participant.name,
			region: participant.region || "",
			metadata: participant.metadata || "",
		});
		setConfigFile(null);
		setEditDialogOpen(true);
	};

	// 다이얼로그 닫기
	const closeCreateDialog = () => {
		setCreateDialogOpen(false);
		form.reset({
			name: "",
			region: "",
			metadata: "",
		});
		setConfigFile(null);
	};

	const closeEditDialog = () => {
		setEditDialogOpen(false);
		setConfigFile(null);
		form.reset();
	};

	// 폼 데이터를 FormData로 변환
	const createFormData = (data: ParticipantFormData) => {
		const formData = new FormData();
		formData.append("name", data.name);
		formData.append("region", data.region);
		if (data.metadata) {
			formData.append("metadata", data.metadata);
		}
		if (configFile) {
			formData.append("configFile", configFile);
		}
		return formData;
	};

	return {
		form,
		createDialogOpen,
		editDialogOpen,
		configFile,
		handleFileChange,
		openCreateDialog,
		openEditDialog,
		closeCreateDialog,
		closeEditDialog,
		createFormData,
		setCreateDialogOpen,
		setEditDialogOpen,
	};
}
