// components/MemoryRequirementInfo.tsx
import { Alert, AlertDescription } from "../../../../components/ui/alert";
import { Info } from "lucide-react";
import { calculateMemoryRequirementDetails} from "../utils/modelMemoryCalculator";
import { ModelAnalysis, formatModelSize } from "../utils/modelDefinitionParser";
import React, { useEffect, useState } from "react";

interface MemoryRequirementInfoProps {
  participantCount: number;
  safetyFactor?: number;
}

export const MemoryRequirementInfo = ({ 
  participantCount,
  safetyFactor = 1.5
}: MemoryRequirementInfoProps) => {
  const [modelAnalysis, setModelAnalysis] = useState<ModelAnalysis | null>(null);

  useEffect(() => {
    // sessionStorage에서 모델 분석 결과 가져오기
    const modelAnalysisData = sessionStorage.getItem("modelAnalysis");
    if (modelAnalysisData) {
      try {
        const analysis = JSON.parse(modelAnalysisData);
        setModelAnalysis(analysis);
        console.log("MemoryRequirementInfo: 모델 분석 데이터 로드됨", {
          totalParams: analysis.totalParams,
          modelSizeBytes: analysis.modelSizeBytes,
          framework: analysis.framework
        });
      } catch (error) {
        console.error("모델 분석 데이터 파싱 실패:", error);
      }
    } else {
      console.warn("MemoryRequirementInfo: 모델 분석 데이터 없음");
    }
  }, []);

  // 모델 분석 결과가 없으면 컴포넌트를 숨김
  if (!modelAnalysis) {
    return (
      <Alert className="mb-4">
        <Info className="h-4 w-4" />
        <AlertDescription>
          <div className="space-y-2">
            <div className="font-medium text-amber-600">모델 분석 정보 없음</div>
            <div className="text-sm">
              모델 정의 파일을 업로드하면 메모리 요구사항을 자동으로 계산합니다.
            </div>
          </div>
        </AlertDescription>
      </Alert>
    );
  }

  const memoryDetails = calculateMemoryRequirementDetails(
    modelAnalysis.modelSizeBytes,
    participantCount,
    safetyFactor
  );

  return (
    <Alert className="mb-4">
      <Info className="h-4 w-4" />
      <AlertDescription>
        <div className="space-y-3">
          <div className="font-medium">모델 기반 메모리 요구사항</div>
          
          <div className="text-sm space-y-2">
            {/* 모델 정보 */}
            <div className="bg-gray-50 p-2 rounded space-y-1">
              <div className="font-medium text-sm">📊 모델 분석 결과</div>
              <div>• 프레임워크: {modelAnalysis.framework}</div>
              <div>• 파라미터 수: {formatModelSize(modelAnalysis.totalParams)}</div>
              <div>• 모델 크기: {memoryDetails.modelSizeFormatted}</div>
            </div>

            {/* 계산 정보 */}
            <div>
              <div>• 참여자 수: {memoryDetails.participantCount}명</div>
              <div>• 안전 계수: {memoryDetails.safetyFactor}배</div>
              <div className="pt-1 font-medium text-blue-600">
                → 최소 요구 RAM: {memoryDetails.recommendedMemoryGB}GB
              </div>
            </div>

            {/* 계산식 */}
            <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
              <div>계산식: {memoryDetails.formula}</div>
              {memoryDetails.notes && (
                <div className="mt-1 text-amber-600">{memoryDetails.notes}</div>
              )}
            </div>

            {/* 레이어 정보 (있는 경우) */}
            {modelAnalysis.layerInfo.length > 0 && (
              <details className="text-xs">
                <summary className="cursor-pointer font-medium">주요 레이어 정보</summary>
                <div className="mt-1 space-y-1 bg-gray-50 p-2 rounded max-h-20 overflow-y-auto">
                  {modelAnalysis.layerInfo.slice(0, 3).map((layer, idx) => (
                    <div key={idx}>
                      • {layer.name}: {layer.params.toLocaleString()} params
                    </div>
                  ))}
                  {modelAnalysis.layerInfo.length > 3 && (
                    <div className="text-gray-500">
                      ... 및 {modelAnalysis.layerInfo.length - 3}개 레이어 더
                    </div>
                  )}
                </div>
              </details>
            )}
          </div>
        </div>
      </AlertDescription>
    </Alert>
  );
};