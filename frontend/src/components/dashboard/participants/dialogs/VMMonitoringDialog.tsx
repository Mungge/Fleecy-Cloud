import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { OpenStackVMInstance } from "@/types/virtual-machine";
import { Participant } from "@/types/participant";
import { getLastAddress } from "../utils";

interface VMMonitoringDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	selectedVM: OpenStackVMInstance | null;
	selectedParticipant: Participant | null;
}

export function VMMonitoringDialog({
	open,
	onOpenChange,
	selectedVM,
	selectedParticipant,
}: VMMonitoringDialogProps) {
	const getVMIPAddress = () => {
		if (!selectedVM?.addresses) return "172.24.4.108";
		const lastAddress = getLastAddress(selectedVM.addresses);
		return lastAddress ? lastAddress.addr : "172.24.4.108";
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="!max-w-[70vw] w-[90vw] min-h-[600px] max-h-[90vh] overflow-auto">
				<DialogHeader>
					<DialogTitle>VM 모니터링 정보</DialogTitle>
					<DialogDescription>
						{selectedVM?.name}의 실시간 모니터링 데이터
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-4">
					{selectedVM && selectedParticipant && (
						<div className="w-full space-y-4">
							<iframe
								src={`${
									selectedParticipant.openstack_endpoint
								}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754642861508&to=1754643013333&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
									selectedVM.name
								}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=20&__feature.dashboardSceneSolo=true`}
								width="100%"
								height="500"
								frameBorder="0"
								title="VM Monitoring Panel 20"
								style={{ borderRadius: "8px" }}
							/>
							<iframe
								src={`${
									selectedParticipant.openstack_endpoint
								}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754712466874&to=1754742634472&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
									selectedVM.name
								}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&panelId=16&__feature.dashboardSceneSolo=true`}
								width="100%"
								height="500"
								frameBorder="0"
								title="VM Monitoring Panel 16"
								style={{ borderRadius: "8px" }}
							/>
						</div>
					)}
				</div>

				<DialogFooter>
					<Button variant="outline" onClick={() => onOpenChange(false)}>
						닫기
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
