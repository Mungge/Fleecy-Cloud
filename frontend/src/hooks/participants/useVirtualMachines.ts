import { useState } from "react";
import { toast } from "sonner";
import { getOpenStackVMs } from "@/api/participants";
import { OpenStackVMInstance } from "@/types/virtual-machine";
import { Participant } from "@/types/participant";

export function useVirtualMachines() {
	const [vmList, setVmList] = useState<OpenStackVMInstance[]>([]);
	const [isVmListLoading, setIsVmListLoading] = useState(false);
	const [vmListDialogOpen, setVmListDialogOpen] = useState(false);
	const [monitoringDialogOpen, setMonitoringDialogOpen] = useState(false);
	const [selectedVM, setSelectedVM] = useState<OpenStackVMInstance | null>(
		null
	);

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

	// VM 모니터링 다이얼로그 열기
	const handleViewVMMonitoring = (vm: OpenStackVMInstance) => {
		setSelectedVM(vm);
		setMonitoringDialogOpen(true);
	};

	const closeVMListDialog = () => {
		setVmListDialogOpen(false);
		setVmList([]);
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
		handleViewVMs,
		handleViewVMMonitoring,
		closeVMListDialog,
		closeMonitoringDialog,
		setVmListDialogOpen,
		setMonitoringDialogOpen,
	};
}
