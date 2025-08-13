// components/dashboard/federated-learning/hooks/useCreateJobForm.ts
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { formSchema, FormValues } from "../types";
import { FORM_DEFAULTS } from "../constants";
import { Participant } from "@/types/participant";

export const useCreateJobForm = (participants: Participant[]) => {
    const router = useRouter();
    const [modelFile, setModelFile] = useState<File | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
  
    const form = useForm<FormValues>({
      resolver: zodResolver(formSchema),
      defaultValues: {
        name: "",
        description: "",
        modelType: "",
        algorithm: FORM_DEFAULTS.ALGORITHM,
        rounds: FORM_DEFAULTS.ROUNDS,
        participants: [],
      },
    });
  
    const openDialog = () => setIsDialogOpen(true);
    
    const closeDialog = () => {
      setIsDialogOpen(false);
      form.reset();
      setModelFile(null);
    };
  
    const handleSubmit = async (values: FormValues) => {
      try {
        // 선택된 참여자 정보 가져오기
        const selectedParticipants = participants.filter((p) =>
          values.participants.includes(p.id)
        );
  
        // 연합학습 정보를 sessionStorage에 저장
        const federatedLearningData = {
          name: values.name,
          description: values.description || "",
          modelType: values.modelType,
          algorithm: values.algorithm,
          rounds: values.rounds,
          participants: selectedParticipants,
          modelFileName: modelFile?.name || null,
          modelFileSize: modelFile ? modelFile.size : 0,
        };
  
        sessionStorage.setItem(
          "federatedLearningData",
          JSON.stringify(federatedLearningData)
        );
  
        // 모델 파일이 있는 경우 별도 저장
        if (modelFile) {
          sessionStorage.setItem("modelFileName", modelFile.name);
          sessionStorage.setItem("modelFileSize", modelFile.size.toString());
        }
  
        // 폼 초기화 및 다이얼로그 닫기
        closeDialog();
  
        // Aggregator 생성 페이지로 이동
        router.push("/dashboard/aggregator/create");
      } catch (error) {
        toast.error("페이지 이동에 실패했습니다: " + error);
      }
    };
  
    return {
      form,
      modelFile,
      setModelFile,
      isDialogOpen,
      openDialog,
      closeDialog,
      handleSubmit,
    };
  };