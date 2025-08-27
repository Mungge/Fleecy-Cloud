"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useParticipants } from "@/hooks/participants/useParticipants";
import { useVirtualMachines } from "@/hooks/participants/useVirtualMachines";
import {
  useParticipantForm,
  ParticipantFormData,
} from "@/hooks/participants/useParticipantForm";
import { Participant } from "@/types/participant";
import { CreateParticipantDialog } from "./dialogs/CreateParticipantDialog";
import { EditParticipantDialog } from "./dialogs/EditParticipantDialog";
import { VMListDialog } from "./dialogs/VMListDialog";
import { VMMonitoringDialog } from "./dialogs/VMMonitoringDialog";
import { ParticipantsTable } from "./tables/ParticipantsTable";
import { ParticipantDetailCard } from "./ParticipantDetailCard";

export default function ParticipantsContent() {
  // Hooks
  const {
    participants,
    isLoading,
    selectedParticipant,
    setSelectedParticipant,
    handleCreateParticipant,
    handleUpdateParticipant,
    handleDeleteParticipant,
    handleHealthCheck,
  } = useParticipants();

  const {
    vmList,
    isVmListLoading,
    vmListDialogOpen,
    monitoringDialogOpen,
    selectedVM,
    isOptimalVMLoading,
    optimalVMInfo,
    handleViewVMs,
    handleViewVMMonitoring,
    handleGetOptimalVM,
    setVmListDialogOpen,
    setMonitoringDialogOpen,
  } = useVirtualMachines();

  const {
    form,
    createDialogOpen,
    editDialogOpen,
    configFile,
    handleFileChange,
    openEditDialog,
    closeCreateDialog,
    closeEditDialog,
    createFormData,
    setCreateDialogOpen,
    setEditDialogOpen,
  } = useParticipantForm();

  // Event handlers
  const handleCreateSubmit = async (data: ParticipantFormData) => {
    const formData = createFormData(data);
    const success = await handleCreateParticipant(formData);
    if (success) {
      closeCreateDialog();
    }
  };

  const handleEditSubmit = async (data: ParticipantFormData) => {
    if (!selectedParticipant) return;
    const formData = createFormData(data);
    const success = await handleUpdateParticipant(
      selectedParticipant.id,
      formData
    );
    if (success) {
      closeEditDialog();
      setSelectedParticipant(null);
    }
  };

  const handleEditParticipant = (participant: Participant) => {
    setSelectedParticipant(participant);
    openEditDialog(participant);
  };

  const handleViewVMsWrapper = async (participant: Participant) => {
    setSelectedParticipant(participant);
    // VM 목록을 가져오는 동시에 최적 VM 계산도 실행
    await Promise.all([
      handleViewVMs(participant),
      handleGetOptimalVM(participant),
    ]);
  };

  const handleRefreshVMs = async () => {
    if (selectedParticipant) {
      // 새로고침 시에도 VM 목록과 최적 VM을 동시에 갱신
      await Promise.all([
        handleViewVMs(selectedParticipant),
        handleGetOptimalVM(selectedParticipant),
      ]);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">
            연합학습 클러스터 관리
          </h2>
          <p className="text-muted-foreground">
            연합학습에 클러스터를 관리하세요.
          </p>
        </div>

        <CreateParticipantDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          form={form}
          configFile={configFile}
          onFileChange={handleFileChange}
          onSubmit={handleCreateSubmit}
          onClose={closeCreateDialog}
        />
      </div>

      <EditParticipantDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        form={form}
        configFile={configFile}
        onFileChange={handleFileChange}
        onSubmit={handleEditSubmit}
      />

      {isLoading ? (
        <div className="flex justify-center items-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* 클러스터 목록 */}
          <Card className="md:col-span-2">
            <CardHeader>
              <CardTitle>클러스터 목록</CardTitle>
              <CardDescription>
                등록된 클러스터들을 관리하고 상태를 모니터링하세요. 행을
                클릭하면 상세 정보를 확인할 수 있습니다.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ParticipantsTable
                participants={participants}
                selectedParticipant={selectedParticipant}
                onSelectParticipant={setSelectedParticipant}
                onEditParticipant={handleEditParticipant}
                onViewVMs={handleViewVMsWrapper}
                onHealthCheck={handleHealthCheck}
                onDeleteParticipant={handleDeleteParticipant}
              />
            </CardContent>
          </Card>

          {/* 상세 정보 카드 */}
          <ParticipantDetailCard
            selectedParticipant={selectedParticipant}
            onEditParticipant={handleEditParticipant}
            onViewVMs={handleViewVMsWrapper}
            onHealthCheck={handleHealthCheck}
            optimalVMInfo={optimalVMInfo}
          />
        </div>
      )}

      {/* VM 목록 다이얼로그 */}
      <VMListDialog
        open={vmListDialogOpen}
        onOpenChange={setVmListDialogOpen}
        selectedParticipant={selectedParticipant}
        vmList={vmList}
        isLoading={isVmListLoading}
        onVMClick={handleViewVMMonitoring}
        onRefresh={handleRefreshVMs}
        optimalVMInfo={optimalVMInfo}
        isOptimalVMLoading={isOptimalVMLoading}
      />

      {/* VM 모니터링 다이얼로그 */}
      <VMMonitoringDialog
        open={monitoringDialogOpen}
        onOpenChange={setMonitoringDialogOpen}
        selectedVM={selectedVM}
        selectedParticipant={selectedParticipant}
      />
    </div>
  );
}
