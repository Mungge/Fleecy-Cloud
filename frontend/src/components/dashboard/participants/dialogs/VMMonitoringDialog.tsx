import {
	Dialog,
	DialogContent,
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
			<DialogContent className="!max-w-[95vw] w-[95vw] min-h-[800px] max-h-[95vh] overflow-auto">
				<DialogHeader>
					<DialogTitle>VM: {selectedVM?.name} 실시간 모니터링</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					{selectedVM && selectedParticipant && (
						<div className="w-full space-y-4">
							<div className="grid grid-cols-4 gap-4">
								<div className="flex flex-col gap-2">
									<div className="flex justify-center">
										<iframe
											src={`${
												selectedParticipant.openstack_endpoint
											}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754794423153&to=1754880823153&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
												selectedVM.name
											}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=14&__feature.dashboardSceneSolo=true`}
											width="70%"
											height="100"
											title="CPU Cores"
											style={{ borderRadius: "8px", border: "none" }}
										/>
									</div>
									<div className="flex justify-center">
										<iframe
											src={`${
												selectedParticipant.openstack_endpoint
											}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754794543348&to=1754880943348&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
												selectedVM.name
											}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=75&__feature.dashboardSceneSolo=true`}
											width="70%"
											height="100"
											title="RAM Total"
											style={{ borderRadius: "8px", border: "none" }}
										/>
									</div>
								</div>

								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754798369210&to=1754884769210&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=20&__feature.dashboardSceneSolo=true`}
										width="90%"
										height="200"
										title="CPU Busy"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>

								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754798546254&to=1754884946254&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=16&__feature.dashboardSceneSolo=true`}
										width="90%"
										height="200"
										title="RAM Used"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754799948036&to=1754886348036&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=154&__feature.dashboardSceneSolo=true`}
										width="90%"
										height="200"
										title="FS Used"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
							</div>

							<div className="grid grid-cols-2 gap-4">
								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754794603353&to=1754881003353&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=77&__feature.dashboardSceneSolo=true`}
										width="100%"
										height="250"
										title="VM Monitoring Panel 77"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754794603353&to=1754881003353&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=78&__feature.dashboardSceneSolo=true`}
										width="100%"
										height="250"
										title="VM Monitoring Panel 78"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
							</div>

							<div className="grid grid-cols-2 gap-4">
								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754799055654&to=1754885455654&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=152&__feature.dashboardSceneSolo=true`}
										width="100%"
										height="250"
										title="VM Monitoring Panel 152"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
								<div className="flex-1">
									<iframe
										src={`${
											selectedParticipant.openstack_endpoint
										}:3000/d-solo/rYdddlPWk/node-exporter-full?orgId=1&from=1754799115711&to=1754885515711&timezone=browser&var-DS_PROMETHEUS=feudbfom5j5kwc&var-job=openstack-vm&var-nodename=${
											selectedVM.name
										}&var-node=${getVMIPAddress()}:9100&var-diskdevices=%5Ba-z%5D%2B%7Cnvme%5B0-9%5D%2Bn%5B0-9%5D%2B%7Cmmcblk%5B0-9%5D%2B&refresh=1m&theme=light&panelId=74&__feature.dashboardSceneSolo=true`}
										width="100%"
										height="250"
										title="VM Monitoring Panel 74"
										style={{ borderRadius: "8px", border: "none" }}
									/>
								</div>
							</div>
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
