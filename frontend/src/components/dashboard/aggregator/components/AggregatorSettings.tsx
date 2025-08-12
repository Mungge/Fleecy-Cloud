// components/AggregatorSettings.tsx
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Check } from "lucide-react";
import { AggregatorOptimizeConfig, CreationStatus } from "../aggregator.types";
import { MemoryRequirementInfo } from "../components/MemoryRequirementInfo";

interface AggregatorSettingsProps {
  config: AggregatorOptimizeConfig;
  onConfigChange: (config: AggregatorOptimizeConfig) => void;
  onOptimize: () => void;
  isLoading: boolean;
  creationStatus: CreationStatus | null;
  modelFileSize?: number; // 추가
  participantCount?: number; // 추가
}

export const AggregatorSettings = ({
  config,
  onConfigChange,
  onOptimize,
  isLoading,
  creationStatus,
  modelFileSize,
  participantCount
}: AggregatorSettingsProps) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>연합학습 집계자 설정</CardTitle>
        <CardDescription>
          연합학습을 위한 집계자의 리소스를 설정하세요.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 메모리 요구사항 정보 표시 */}
        {modelFileSize && participantCount && (
          <MemoryRequirementInfo 
            modelFileSize={modelFileSize}
            participantCount={participantCount}
            safetyFactor={1.5}
          />
        )}

        {/* 제약조건 설정 */}
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <Label htmlFor="budget">최대 월 예산 제약조건</Label>
            <span className="text-sm font-medium text-green-600">
              {config.maxBudget.toLocaleString()}원
            </span>
          </div>
          <Slider
            id="budget"
            value={[config.maxBudget]}
            onValueChange={([value]) => 
              onConfigChange({ ...config, maxBudget: value })
            }
            max={1000000}
            min={50000}
            step={10000}
            className="w-full"
          />
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>10만원</span>
            <span>200만원</span>
          </div>
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <Label htmlFor="latency">최대 허용 지연시간 제약조건</Label>
            <span className="text-sm font-medium text-blue-600">
              {config.maxLatency}ms
            </span>
          </div>
          <Slider
            id="latency"
            value={[config.maxLatency]}
            onValueChange={([value]) => 
              onConfigChange({ ...config, maxLatency: value })
            }
            max={500}
            min={20}
            step={5}
            className="w-full"
          />
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>20ms (매우 빠름)</span>
            <span>500ms (여유)</span>
          </div>
        </div>

        {/* 현재 설정 요약 */}
        <div className="mt-4 p-3 bg-gray-50 rounded-md">
          <div className="text-sm text-muted-foreground mb-1">제약조건:</div>
          <div className="text-sm">
            월 최대 <span className="font-medium text-green-600">{config.maxBudget.toLocaleString()}원</span> 예산으로{" "}
            <span className="font-medium text-blue-600">{config.maxLatency}ms</span> 이하의 응답속도 보장
          </div>
        </div>

        <div className="pt-4">
          <Button
            onClick={onOptimize}
            disabled={isLoading || creationStatus?.step === "completed"}
            className="w-full"
            variant={
              creationStatus?.step === "completed" ? "secondary" : "default"
            }
          >
            {isLoading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-white mr-2"></div>
                {creationStatus?.message || "생성 중..."}
              </>
            ) : creationStatus?.step === "completed" ? (
              <>
                <Check className="mr-2 h-4 w-4" />
                생성 완료
              </>
            ) : creationStatus?.step === "error" ? (
              "다시 시도"
            ) : (
              "집계자 배치 최적화 실행"
            )}
          </Button>

          {creationStatus?.step === "completed" && (
            <p className="text-sm text-green-600 text-center mt-2">
              잠시 후 연합학습 페이지로 이동합니다...
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  );
};