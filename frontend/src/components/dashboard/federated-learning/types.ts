// components/dashboard/federated-learning/types.ts
import * as z from "zod";

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

export interface FederatedLearningHookReturn {
  jobs: any[];
  selectedJob: any | null;
  isLoading: boolean;
  fetchJobs: () => Promise<void>;
  deleteJob: (id: string) => Promise<void>;
  selectJob: (job: any) => void;
}

export interface ParticipantsHookReturn {
  participants: any[];
  fetchParticipants: () => Promise<void>;
}

export interface CreateJobFormHookReturn {
  form: any;
  modelFile: File | null;
  setModelFile: (file: File | null) => void;
  isDialogOpen: boolean;
  openDialog: () => void;
  closeDialog: () => void;
  handleSubmit: (values: FormValues) => Promise<void>;
}