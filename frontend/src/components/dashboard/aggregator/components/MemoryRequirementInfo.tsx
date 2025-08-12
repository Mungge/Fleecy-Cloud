// components/MemoryRequirementInfo.tsx
import { Alert, AlertDescription } from "../../../../components/ui/alert";
import { Info } from "lucide-react";
import { calculateMemoryRequirementDetails} from "../utils/modelMemoryCalculator";
import React from "react";

interface MemoryRequirementInfoProps {
  modelFileSize: number; // bytes
  participantCount: number;
  safetyFactor?: number;
}

export const MemoryRequirementInfo = ({ 
  modelFileSize, 
  participantCount,
  safetyFactor = 1.5
}: MemoryRequirementInfoProps) => {
  if (!modelFileSize || modelFileSize === 0) {
    return null;
  }

  const memoryDetails = calculateMemoryRequirementDetails(
    modelFileSize,
    participantCount,
    safetyFactor
  );

  return (
    <Alert className="mb-4">
      <Info className="h-4 w-4" />
      <AlertDescription>
        <div className="space-y-2">
          <div className="font-medium">최소 메모리 요구사항 계산</div>
          <div className="text-sm space-y-1">
            <div>• 모델 파일 크기: {memoryDetails.modelFileSizeFormatted}</div>
            <div>• 참여자 수: {memoryDetails.participantCount}명</div>
            <div>• 안전 계수: {memoryDetails.safetyFactor}배</div>
            <div className="pt-1 font-medium text-blue-600">
              → 최소 요구 RAM: {memoryDetails.recommendedMemoryGB}GB
            </div>
            <div className="text-xs text-gray-500">
              계산식: {memoryDetails.formula}
            </div>
          </div>
        </div>
      </AlertDescription>
    </Alert>
  );
};