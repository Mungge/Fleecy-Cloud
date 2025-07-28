// components/dashboard/federated-learning/types.ts
import * as z from "zod";
import { Participant } from "@/types/participant";
import { UseFormReturn } from "react-hook-form";

export const formSchema = z.object({
    name: z.string().min(1, "이름을 입력해주세요"),
    description: z.string().optional(),
    modelType: z.string().min(1, "모델 유형을 선택해주세요"),
    algorithm: z.string().min(1, "집계 알고리즘을 선택해주세요"),
    rounds: z
      .number()
      .int()
      .min(1, "최소 1회 이상의 라운드가 필요합니다")
      .max(100, "최대 100회까지 설정 가능합니다"),
    participants: z
      .array(z.string())
      .min(1, "최소 1개 이상의 참여자가 필요합니다"),
  });
  
  export type FormValues = z.infer<typeof formSchema>;
  
  // Job 관련 타입 정의
  export interface Job {
    id: string;
    name: string;
    description?: string;
    modelType: string;
    algorithm: string;
    rounds: number;
    participants: Participant[];
    status: JobStatus;
    createdAt: Date;
    updatedAt: Date;
    progress?: number;
  }
  
  export type JobStatus = 'pending' | 'running' | 'completed' | 'failed' | 'paused';

  
  export interface ParticipantMetrics {
    accuracy: number;
    loss: number;
    round: number;
    trainingTime: number;
  }
  
  // Hook return 타입들
  export interface FederatedLearningHookReturn {
    jobs: Job[];
    selectedJob: Job | null;
    isLoading: boolean;
    fetchJobs: () => Promise<void>;
    deleteJob: (id: string) => Promise<void>;
    selectJob: (job: Job) => void;
  }
  
  export interface ParticipantsHookReturn {
    participants: Participant[];
    fetchParticipants: () => Promise<void>;
  }
  
  export interface CreateJobFormHookReturn {
    form: UseFormReturn<FormValues>;
    modelFile: File | null;
    setModelFile: (file: File | null) => void;
    isDialogOpen: boolean;
    openDialog: () => void;
    closeDialog: () => void;
    handleSubmit: (values: FormValues) => Promise<void>;
  }
  
  // 추가로 필요할 수 있는 타입들
  export interface FederatedLearningData {
    name: string;
    description: string;
    modelType: string;
    algorithm: string;
    rounds: number;
    participants: Participant[];
    modelFileName: string | null;
  }
  
  // 상수들
  export const FORM_DEFAULTS = {
    ALGORITHM: "fedavg",
    ROUNDS: 10,
  } as const;
