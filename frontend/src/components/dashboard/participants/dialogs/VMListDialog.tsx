import { Server, CheckCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { OpenStackVMInstance } from "@/types/virtual-machine";
import { Participant } from "@/types/participant";
import { getVMStatusBadge, formatBytes, getLastAddress } from "../utils";

interface OptimalVMData {
  selected_vm: {
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
  };
  selection_reason: string;
  candidate_count: number;
}

interface VMListDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedParticipant: Participant | null;
  vmList: OpenStackVMInstance[];
  isLoading: boolean;
  onVMClick: (vm: OpenStackVMInstance) => void;
  onRefresh: () => void;
  optimalVMInfo?: OptimalVMData | null; // 최적 VM 정보
  isOptimalVMLoading?: boolean; // 최적 VM 로딩 상태
}

export function VMListDialog({
  open,
  onOpenChange,
  selectedParticipant,
  vmList,
  isLoading,
  onVMClick,
  onRefresh,
  optimalVMInfo,
  isOptimalVMLoading = false,
}: VMListDialogProps) {
  // 최적 VM의 instance_id 가져오기
  const selectedVMId = optimalVMInfo?.selected_vm.instance_id;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="!max-w-[80vw] w-full min-h-[700px] max-h-[95vh] overflow-auto flex flex-col">
        <DialogHeader>
          <DialogTitle>가상머신 목록</DialogTitle>
          {/* 👇 이 부분이 수정되었습니다. */}
          <DialogDescription>
            {selectedParticipant?.name} 클러스터의 가상머신 목록
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
            <span className="ml-2">VM 목록을 가져오는 중...</span>
          </div>
        ) : (
          <div className="flex-1 overflow-auto space-y-4">
            {vmList.length > 0 ? (
              <>
                <div className="rounded-md border">
                  <Table>
                    <TableHeader className="sticky top-0 bg-white z-10">
                      <TableRow>
                        <TableHead className="w-[180px]">이름</TableHead>
                        <TableHead className="w-[120px]">상태</TableHead>
                        <TableHead className="w-[180px]">
                          스펙 (CPU/RAM/Disk)
                        </TableHead>
                        <TableHead className="w-[220px]">IP 주소</TableHead>
                        <TableHead className="w-[140px]">최적 VM</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {vmList.map((vm) => {
                        const isOptimalVM = selectedVMId === vm.id;
                        return (
                          <TableRow
                            key={vm.id}
                            className={`hover:bg-muted/50 cursor-pointer ${
                              isOptimalVM
                                ? "bg-green-50 border-l-4 border-l-green-500"
                                : ""
                            }`}
                            onClick={() => onVMClick(vm)}
                          >
                            <TableCell className="align-top">
                              <div className="flex items-center gap-2">
                                {isOptimalVM && (
                                  <CheckCircle className="h-4 w-4 text-green-600 flex-shrink-0" />
                                )}
                                <div className="font-medium break-words">
                                  {vm.name}
                                </div>
                              </div>
                            </TableCell>
                            <TableCell className="align-top">
                              <div className="space-y-1">
                                {getVMStatusBadge(vm.status)}
                                <div className="text-xs text-gray-500">
                                  {vm["OS-EXT-STS:power_state"] === 1
                                    ? "Running"
                                    : "Stopped"}
                                </div>
                              </div>
                            </TableCell>
                            <TableCell className="align-top">
                              <div className="space-y-1">
                                <div className="font-medium text-sm">
                                  {vm.flavor.name || vm.flavor.id}
                                </div>
                                <div className="text-xs text-gray-600 space-y-0.5">
                                  <div className="flex items-center gap-1">
                                    <span className="font-mono">CPU:</span>
                                    <span>{vm.flavor.vcpus || 0} vCPU</span>
                                  </div>
                                  <div className="flex items-center gap-1">
                                    <span className="font-mono">RAM:</span>
                                    <span>
                                      {vm.flavor.ram
                                        ? formatBytes(vm.flavor.ram)
                                        : "0 GB"}
                                    </span>
                                  </div>
                                  <div className="flex items-center gap-1">
                                    <span className="font-mono">Disk:</span>
                                    <span>{vm.flavor.disk || 0} GB</span>
                                  </div>
                                </div>
                              </div>
                            </TableCell>
                            <TableCell className="align-top">
                              <div className="space-y-1 max-w-[220px]">
                                {Object.keys(vm.addresses || {}).length > 0 ? (
                                  (() => {
                                    const lastAddress = getLastAddress(
                                      vm.addresses
                                    );
                                    return lastAddress ? (
                                      <div className="space-y-1">
                                        <div className="flex items-center gap-2 flex-wrap">
                                          <span className="font-mono text-sm break-all">
                                            {lastAddress.addr}
                                          </span>
                                          <Badge
                                            variant="outline"
                                            className="text-xs flex-shrink-0"
                                          >
                                            {lastAddress.type}
                                          </Badge>
                                        </div>
                                        <div className="text-xs text-gray-500">
                                          {lastAddress.networkName}
                                        </div>
                                      </div>
                                    ) : null;
                                  })()
                                ) : (
                                  <span className="text-sm text-gray-500">
                                    없음
                                  </span>
                                )}
                              </div>
                            </TableCell>
                            <TableCell className="align-top">
                              {isOptimalVM ? (
                                <Badge
                                  variant="secondary"
                                  className="bg-green-100 text-green-800 border-green-300"
                                >
                                  <CheckCircle className="h-3 w-3 mr-1" />
                                  최적 VM
                                </Badge>
                              ) : (
                                <span className="text-sm text-gray-400">-</span>
                              )}
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </div>

                <div className="flex items-center justify-between text-sm text-gray-500 px-2 pb-2">
                  <span>
                    총 {vmList.length}개의 가상머신이 있습니다.
                    {selectedVMId && " (최적 VM 1개 선택됨)"}
                  </span>
                  <span>
                    마지막 업데이트: {new Date().toLocaleTimeString()}
                  </span>
                </div>
              </>
            ) : (
              <div className="text-center py-12">
                <div className="mx-auto w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                  <Server className="h-12 w-12 text-gray-400" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">
                  가상머신이 없습니다
                </h3>
                <p className="text-gray-500">
                  이 클러스터에는 아직 가상머신이 없습니다.
                </p>
              </div>
            )}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            닫기
          </Button>
          <Button
            onClick={onRefresh}
            disabled={isLoading || isOptimalVMLoading}
          >
            {isLoading || isOptimalVMLoading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-current mr-2" />
                {isOptimalVMLoading ? "최적 VM 계산 중..." : "새로고침 중..."}
              </>
            ) : (
              "새로고침"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
