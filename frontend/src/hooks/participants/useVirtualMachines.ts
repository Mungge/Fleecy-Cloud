import { useState } from "react";
import { toast } from "sonner";
import { getOpenStackVMs, getOptimalVM } from "@/api/participants";
import { OpenStackVMInstance } from "@/types/virtual-machine";
import { Participant } from "@/types/participant";

interface SelectedVM {
  id: string;
  participant_id: string;
  name: string;
  instance_id: string;
  status: string;
  flavor_id: string;
  flavor_name: string;
  vcpus: number;
  ram: number;
  disk: number;
  ip_addresses: string;
  created_at: string;
  updated_at: string;
}

interface OptimalVMResponse {
  data: {
    selected_vm: SelectedVM;
    selection_reason: string;
    candidate_count: number;
  };
  message: string;
  success: boolean;
}

export function useVirtualMachines() {
  const [vmList, setVmList] = useState<OpenStackVMInstance[]>([]);
  const [isVmListLoading, setIsVmListLoading] = useState(false);
  const [vmListDialogOpen, setVmListDialogOpen] = useState(false);
  const [monitoringDialogOpen, setMonitoringDialogOpen] = useState(false);
  const [selectedVM, setSelectedVM] = useState<OpenStackVMInstance | null>(
    null
  );
  const [isOptimalVMLoading, setIsOptimalVMLoading] = useState(false);
  const [optimalVMInfo, setOptimalVMInfo] = useState<
    OptimalVMResponse["data"] | null
  >(null);

  // VM 목록 조회
  const handleViewVMs = async (participant: Participant) => {
    setIsVmListLoading(true);
    setVmListDialogOpen(true);

    try {
      const vms = await getOpenStackVMs(participant.id);
      setVmList(vms);
    } catch (error) {
      console.error("VM 목록 조회 실패:", error);
      toast.error("VM 목록을 불러오는데 실패했습니다.");
      setVmList([]);
    } finally {
      setIsVmListLoading(false);
    }
  };

  // 최적의 VM 조회
  const handleGetOptimalVM = async (participant: Participant) => {
    setIsOptimalVMLoading(true);

    try {
      const response = (await getOptimalVM(
        participant.id
      )) as OptimalVMResponse;
      setOptimalVMInfo(response.data);
      toast.success(response.message);
      return response.data.selected_vm;
    } catch (error) {
      console.error("최적 VM 조회 실패:", error);
      toast.error("최적의 VM을 찾는데 실패했습니다.");
      setOptimalVMInfo(null);
      return null;
    } finally {
      setIsOptimalVMLoading(false);
    }
  };

  // VM 모니터링 다이얼로그 열기
  const handleViewVMMonitoring = (vm: OpenStackVMInstance) => {
    setSelectedVM(vm);
    setMonitoringDialogOpen(true);
  };

  const closeVMListDialog = () => {
    setVmListDialogOpen(false);
  };

  const closeMonitoringDialog = () => {
    setMonitoringDialogOpen(false);
    setSelectedVM(null);
  };

  return {
    vmList,
    isVmListLoading,
    vmListDialogOpen,
    monitoringDialogOpen,
    selectedVM,
    isOptimalVMLoading,
    optimalVMInfo,
    handleViewVMs,
    handleGetOptimalVM,
    handleViewVMMonitoring,
    closeVMListDialog,
    closeMonitoringDialog,
    setVmListDialogOpen,
    setMonitoringDialogOpen,
  };
}
