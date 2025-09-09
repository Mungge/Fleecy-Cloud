// @/components/dashboard/federated-learning/federated-learning-start-content.tsx
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAggregatorCreationStore } from "../aggregator/aggregator.types";
import {
  startFederatedLearning,
  getFirstActiveCloudConnection,
} from "@/api/federatedLearning";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Play,
  Server,
  HardDrive,
  Users,
  FileText,
  Layers,
  Clock,
  ArrowLeft,
  CheckCircle,
  Monitor,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { VMListDialog } from "@/components/dashboard/participants/dialogs/VMListDialog";
import { OpenStackVMInstance } from "@/types/virtual-machine";
import { Participant } from "@/types/participant";
import { useVirtualMachines } from "@/hooks/participants/useVirtualMachines";

// Mock data for FederatedLearningStartContent
const mockPayload = {
  selectedOption: {
    cloudProvider: "AWS",
    region: "ap-northeast-2",
    instanceType: "t3.medium",
    vcpu: 2,
    memory: 4096, // MB
    avgLatency: 15,
    estimatedMonthlyCost: 85000,
  },
  federatedLearningData: {
    name: "의료 이미지 분류 모델",
    description: "병원간 협업을 통한 X-ray 이미지 분류 연합학습",
    model_type: "CNN",
    algorithm: "FedAvg",
    rounds: 10,
    participants: [
      {
        id: "5fb4a51a-3974-4568-8080-8dbf792fc6c",
        name: "서울대학교병원",
        status: "active",
      },
      {
        id: "5fb4a51a-3974-4568-8080-8dbf792fc6c",
        name: "연세세브란스병원",
        status: "active",
      },
      {
        id: "5fb4a51a-3974-4568-8080-8dbf792fc6c",
        name: "삼성서울병원",
        status: "active",
      },
      {
        id: "5fb4a51a-3974-4568-8080-8dbf792fc6c",
        name: "아산의료원",
        status: "active",
      },
    ],
    modelFileName: "xray_classification_model.py",
  },
  aggregatorId: "agg-12345-abcde",
};

const FederatedLearningStartContent = () => {
  const router = useRouter();
  // Store에서 payload를 가져오되, 없으면 mock data 사용
  const storePayload = useAggregatorCreationStore((s) => s.payload);
  const payload = storePayload || mockPayload;
  const [isStarting, setIsStarting] = useState(false);

  // useVirtualMachines 훅 사용
  const {
    vmList,
    isVmListLoading,
    vmListDialogOpen,
    optimalVMInfo,
    isOptimalVMLoading,
    handleViewVMs,
    handleGetOptimalVM,
    handleViewVMMonitoring,
    closeVMListDialog,
    setVmListDialogOpen,
  } = useVirtualMachines();

  // Collapsible states
  const [isAggregatorExpanded, setIsAggregatorExpanded] = useState(true);
  const [isJobInfoExpanded, setIsJobInfoExpanded] = useState(true);
  const [isParticipantsExpanded, setIsParticipantsExpanded] = useState(true);

  // 현재 선택된 참여자
  const [selectedParticipant, setSelectedParticipant] =
    useState<Participant | null>(null);

  // payload가 없으면 이전 페이지로 리다이렉트
  // 원치 않는다면 주석처리 => mock data가 보일 것임
  useEffect(() => {
    // if (!storePayload) {
    //   router.replace("/dashboard/federated-learning");
    // }
  }, [storePayload, router]);

  const selectedOption = payload?.selectedOption;
  const federatedLearningData = payload?.federatedLearningData;

  const handleStartFederatedLearning = async (): Promise<void> => {
    setIsStarting(true);

    try {
      // payload 데이터를 사용하여 연합학습 시작 API 호출
      if (!payload || !selectedOption || !federatedLearningData) {
        throw new Error("필요한 데이터가 없습니다. 다시 시도해주세요.");
      }

      // 사용자의 첫 번째 활성 클라우드 연결 가져오기
      const cloudConnectionId = await getFirstActiveCloudConnection();

      // aggregatorId 확인
      if (!payload.aggregatorId) {
        throw new Error(
          "Aggregator ID가 없습니다. 먼저 집계자를 배포해주세요."
        );
      }

      const result = await startFederatedLearning({
        aggregatorId: payload.aggregatorId,
        cloudConnectionId,
        federatedLearningData: {
          name: federatedLearningData.name,
          description: federatedLearningData.description || "",
          modelType: federatedLearningData.model_type || "CNN",
          algorithm: federatedLearningData.algorithm,
          rounds: federatedLearningData.rounds,
          participants: federatedLearningData.participants || [],
          modelFileName: federatedLearningData.modelFileName || undefined,
        },
      });

      console.log("연합학습 시작 성공:", result);

      toast.success("연합학습이 성공적으로 시작되었습니다!", {
        description: `연합학습 ID: ${result.federatedLearningId}`,
        duration: 5000,
      });

      // 성공 후 대시보드 또는 모니터링 페이지로 이동
      router.push("/dashboard/federated-learning");
    } catch (error) {
      console.error("연합학습 시작 실패:", error);

      const errorMessage =
        error instanceof Error
          ? error.message
          : "연합학습 시작에 실패했습니다.";

      toast.error("연합학습 시작 실패", {
        description: errorMessage,
        duration: 5000,
      });
    } finally {
      setIsStarting(false);
    }
  };

  const handleGoBack = () => {
    router.back();
  };

  const handleParticipantVMClick = async (participant: any) => {
    setSelectedParticipant(participant);

    // VM 목록 조회
    await handleViewVMs(participant);

    // 최적 VM 조회
    await handleGetOptimalVM(participant);
  };

  const handleVMClick = (vm: OpenStackVMInstance) => {
    console.log("Selected VM:", vm);
    toast.success(`VM ${vm.name}이 선택되었습니다.`);
    // 필요하다면 VM 모니터링 다이얼로그 열기
    // handleViewVMMonitoring(vm);
  };

  const handleVMRefresh = async () => {
    if (selectedParticipant) {
      // VM 목록 새로고침
      await handleViewVMs(selectedParticipant);
      // 최적 VM 재계산
      await handleGetOptimalVM(selectedParticipant);
    }
  };

  //mock 데이터로 인해 항상 payload가 존재하므로 로딩 화면이 표시되지 않음
  if (!payload || !selectedOption || !federatedLearningData) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
        <span className="ml-3">데이터를 불러오는 중...</span>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 헤더 */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <div className="flex items-center gap-3">
            <Play className="h-8 w-8 text-green-600" />
            <h1 className="text-3xl font-bold text-green-600">
              연합학습 시작 준비
            </h1>
          </div>
          <p className="text-muted-foreground">
            모든 설정이 완료되었습니다. 연합학습을 시작하시겠습니까?
          </p>
          {!storePayload && (
            <div className="text-xs text-orange-600 bg-orange-50 px-2 py-1 rounded">
              💡 Mock 데이터로 미리보기 중입니다
            </div>
          )}
        </div>
        <Button variant="outline" onClick={handleGoBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          이전으로
        </Button>
      </div>

      {/* 배포된 집계자 정보 */}
      <Card className="border-2 border-green-200 bg-green-50/30">
        <Collapsible
          open={isAggregatorExpanded}
          onOpenChange={setIsAggregatorExpanded}
        >
          <CollapsibleTrigger asChild>
            <CardHeader className="cursor-pointer hover:bg-green-100/50 transition-colors">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500" />
                  배포된 집계자 정보
                </CardTitle>
                {isAggregatorExpanded ? (
                  <ChevronUp className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
              <CardDescription>
                성공적으로 배포된 집계자 인스턴스 정보
              </CardDescription>
            </CardHeader>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <CardContent className="grid gap-4 md:grid-cols-2 pt-0">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">클라우드 제공자</span>
                  <Badge className="bg-orange-500">
                    {selectedOption.cloudProvider}
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">리전</span>
                  <Badge variant="outline">{selectedOption.region}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">인스턴스 타입</span>
                  <div className="flex items-center gap-1">
                    <Server className="h-3 w-3" />
                    <span className="text-sm font-mono">
                      {selectedOption.instanceType}
                    </span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">상태</span>
                  <Badge className="bg-green-500">활성</Badge>
                </div>
              </div>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">vCPU</span>
                  <span className="text-sm">{selectedOption.vcpu} 코어</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">메모리</span>
                  <div className="flex items-center gap-1">
                    <HardDrive className="h-3 w-3" />
                    <span className="text-sm">
                      {((selectedOption.memory || 0) / 1024).toFixed(1)}GB
                    </span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">지연시간</span>
                  <span className="text-sm text-green-600">
                    {selectedOption.avgLatency}ms
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">월 예상 비용</span>
                  <span className="text-sm font-semibold">
                    ₩{selectedOption.estimatedMonthlyCost.toLocaleString()}
                  </span>
                </div>
              </div>
            </CardContent>
          </CollapsibleContent>
        </Collapsible>
      </Card>

      {/* 연합학습 작업 정보 */}
      <Card>
        <Collapsible
          open={isJobInfoExpanded}
          onOpenChange={setIsJobInfoExpanded}
        >
          <CollapsibleTrigger asChild>
            <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Layers className="h-5 w-5 text-blue-500" />
                  연합학습 작업 정보
                </CardTitle>
                {isJobInfoExpanded ? (
                  <ChevronUp className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
              <CardDescription>
                실행할 연합학습 작업의 세부 정보
              </CardDescription>
            </CardHeader>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <CardContent className="space-y-4 pt-0">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">작업 이름</span>
                    <span className="text-sm font-semibold">
                      {federatedLearningData.name}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">알고리즘</span>
                    <Badge variant="outline">
                      {federatedLearningData.algorithm}
                    </Badge>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">라운드 수</span>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      <span className="text-sm">
                        {federatedLearningData.rounds}회
                      </span>
                    </div>
                  </div>
                </div>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">참여자 수</span>
                    <div className="flex items-center gap-1">
                      <Users className="h-3 w-3" />
                      <span className="text-sm">
                        {federatedLearningData?.participants?.length ?? 0}명
                      </span>
                    </div>
                  </div>
                  {federatedLearningData.modelFileName && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">모델 파일</span>
                      <div className="flex items-center gap-1">
                        <FileText className="h-3 w-3" />
                        <span className="text-sm font-mono">
                          {federatedLearningData.modelFileName}
                        </span>
                      </div>
                    </div>
                  )}
                  {federatedLearningData.description && (
                    <div className="space-y-1">
                      <span className="text-sm font-medium">설명</span>
                      <p className="text-sm text-muted-foreground">
                        {federatedLearningData.description}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </CollapsibleContent>
        </Collapsible>
      </Card>

      {/* 참여자 목록 */}
      <Card>
        <Collapsible
          open={isParticipantsExpanded}
          onOpenChange={setIsParticipantsExpanded}
        >
          <CollapsibleTrigger asChild>
            <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5 text-purple-500" />
                  참여자 목록
                </CardTitle>
                {isParticipantsExpanded ? (
                  <ChevronUp className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
              <CardDescription>
                연합학습에 참여할{" "}
                {federatedLearningData?.participants?.length ?? 0}
                명의 참여자 및 VM 정보
              </CardDescription>
            </CardHeader>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <CardContent className="pt-0">
              <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                {federatedLearningData?.participants?.map(
                  (participant, index) => (
                    <div
                      key={index}
                      className="flex items-center justify-between p-3 border rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                          <span className="text-xs font-medium text-purple-600">
                            {index + 1}
                          </span>
                        </div>
                        <div>
                          <p className="text-sm font-medium">
                            {participant.name || `참여자 ${index + 1}`}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {participant.id}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="text-xs">
                          준비됨
                        </Badge>
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-8 px-3"
                          onClick={() => handleParticipantVMClick(participant)}
                        >
                          <Monitor className="h-3 w-3 mr-1" />
                          VM
                        </Button>
                      </div>
                    </div>
                  )
                )}
              </div>
            </CardContent>
          </CollapsibleContent>
        </Collapsible>
      </Card>

      {/* 시작 버튼 */}
      <div className="flex justify-center gap-4">
        <Button
          onClick={handleStartFederatedLearning}
          disabled={isStarting}
          size="lg"
          className="min-w-[200px] bg-green-600 hover:bg-green-700"
        >
          {isStarting ? (
            <>
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
              연합학습 저장 중...
            </>
          ) : (
            <>
              <Play className="h-4 w-4 mr-2" />
              연합학습 시작 및 저장
            </>
          )}
        </Button>
      </div>

      {/* VM List Dialog */}
      <VMListDialog
        open={vmListDialogOpen}
        onOpenChange={setVmListDialogOpen}
        selectedParticipant={selectedParticipant}
        vmList={vmList}
        isLoading={isVmListLoading}
        onVMClick={handleVMClick}
        onRefresh={handleVMRefresh}
        optimalVMInfo={optimalVMInfo}
        isOptimalVMLoading={isOptimalVMLoading}
      />
    </div>
  );
};

export default FederatedLearningStartContent;
