// components/dashboard/federated-learning/federated-learning-content.tsx
"use client";

import { useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Check } from "lucide-react";
import { useFederatedLearning } from "./hooks/useFederatedLearning";
import { useParticipants } from "./hooks/useParticipants";
import { useCreateJobForm } from "./hooks/useCreateJobForm";
import { FederatedLearningList } from "./components/FederatedLearningList";
import { FederatedLearningDetail } from "./components/FederatedLearningDetail";
import { CreateJobDialog } from "./components/CreateJobDialog";

const FederatedLearningContent = () => {
  const federatedLearning = useFederatedLearning();
  const participantsHook = useParticipants();
  const createFormHook = useCreateJobForm(participantsHook.participants);

  // 페이지 로드 시 데이터 가져오기
  useEffect(() => {
	let mounted = true;
  
	const loadData = async () => {
	  if (mounted) {
		await Promise.all([
		  federatedLearning.fetchJobs(),
		  participantsHook.fetchParticipants()
		]);
	  }
	};
  
	loadData();
  
	return () => {
	  mounted = false;
	};
	// eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // 컴포넌트 마운트 시에만 실행하고 싶다면

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">연합학습</h2>
          <p className="text-muted-foreground">
            연합학습 작업을 생성하고 모니터링하세요.
          </p>
        </div>

        <CreateJobDialog
          participants={participantsHook.participants}
          formHook={createFormHook}
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <FederatedLearningList
          jobs={federatedLearning.jobs}
          isLoading={federatedLearning.isLoading}
          onJobSelect={federatedLearning.selectJob}
          onJobDelete={federatedLearning.deleteJob}
        />

        <FederatedLearningDetail
          selectedJob={federatedLearning.selectedJob}
        />
      </div>

      {/* Progress Steps - Bottom */}
      <Card>
        <CardHeader>
          <CardTitle>연합학습 생성 단계</CardTitle>
          <CardDescription>
            연합학습을 생성하기 위한 단계별 진행 상황을 확인하세요.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="w-full py-6">
            <div className="flex items-center justify-between max-w-2xl mx-auto">
              {/* Step 1: 정보 입력 */}
              <div className="flex flex-col items-center">
                <div className="flex items-center justify-center w-12 h-12 rounded-full bg-blue-500 text-white text-lg font-medium shadow-lg">
                  <Check className="w-6 h-6" />
                </div>
                <span className="mt-3 text-base font-medium text-blue-600">
                  정보 입력
                </span>
                <span className="mt-1 text-sm text-gray-500">
                  연합학습 정보 설정
                </span>
              </div>

              {/* Connector Line */}
              <div className="flex-1 h-1 bg-gray-200 mx-6 rounded-full"></div>

              {/* Step 2: 집계자 생성 */}
              <div className="flex flex-col items-center">
                <div className="flex items-center justify-center w-12 h-12 rounded-full bg-gray-200 text-gray-400 text-lg font-medium">
                  2
                </div>
                <span className="mt-3 text-base text-gray-400">
                  집계자 생성
                </span>
                <span className="mt-1 text-sm text-gray-400">
                  Aggregator 설정
                </span>
              </div>

              {/* Connector Line */}
              <div className="flex-1 h-1 bg-gray-200 mx-6 rounded-full"></div>

              {/* Step 3: 연합학습 생성 */}
              <div className="flex flex-col items-center">
                <div className="flex items-center justify-center w-12 h-12 rounded-full bg-gray-200 text-gray-400 text-lg font-medium">
                  3
                </div>
                <span className="mt-3 text-base text-gray-400">
                  연합학습 생성
                </span>
                <span className="mt-1 text-sm text-gray-400">
                  최종 생성 완료
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default FederatedLearningContent;