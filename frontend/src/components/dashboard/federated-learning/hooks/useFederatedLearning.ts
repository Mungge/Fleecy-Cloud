// components/dashboard/federated-learning/hooks/useFederatedLearning.ts
import { useState, useCallback } from "react";
import { toast } from "sonner";
import { getFederatedLearnings, deleteFederatedLearning } from "@/api/federatedLearning";
import { FederatedLearningJob } from "@/types/federated-learning";

export const useFederatedLearning = () => {
  const [jobs, setJobs] = useState<FederatedLearningJob[]>([]);
  const [selectedJob, setSelectedJob] = useState<FederatedLearningJob | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchJobs = useCallback(async () => {
    try {
      setIsLoading(true);
      const jobList = await getFederatedLearnings();
      setJobs(jobList);
    } catch (error) {
      toast.error("연합학습 작업 조회에 실패했습니다: " + error);
      setJobs([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const deleteJob = useCallback(async (id: string) => {
    try {
      await deleteFederatedLearning(id);
      toast.success("연합학습 작업이 삭제되었습니다.");
      await fetchJobs();
      
      if (selectedJob?.id === id) {
        setSelectedJob(null);
      }
    } catch (error) {
      console.error("연합학습 작업 삭제에 실패했습니다: ", error);
      toast.error("연합학습 작업 삭제에 실패했습니다.");
    }
  }, [selectedJob, fetchJobs]);

  const selectJob = useCallback((job: FederatedLearningJob) => {
    setSelectedJob(job);
  }, []);

  return {
    jobs,
    selectedJob,
    isLoading,
    fetchJobs,
    deleteJob,
    selectJob,
  };
};