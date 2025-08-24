// components/AggregatorSettings.tsx (업데이트된 부분)
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../../components/ui/card";
import { Button } from "../../../../components/ui/button";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Slider } from "@/components/ui/slider";
import { AggregatorOptimizeConfig, CreationStatus } from "../aggregator.types";
import { MemoryRequirementInfo } from "./MemoryRequirementInfo";

interface AggregatorSettingsProps {
  config: AggregatorOptimizeConfig;
  onConfigChange: (config: AggregatorOptimizeConfig) => void;
  onOptimize: () => void;
  isLoading: boolean;
  creationStatus: CreationStatus | null;
  participantCount: number;
}

export const AggregatorSettings = ({
  config,
  onConfigChange,
  onOptimize,
  isLoading,
  creationStatus,
  participantCount,
}: AggregatorSettingsProps) => {
  const canOptimize = !isLoading && (!creationStatus || creationStatus.step === "error");

  return (
    <Card>
      <CardHeader>
        <CardTitle>집계자 설정</CardTitle>
        <CardDescription>
          집계자 배치 최적화를 위한 제약 조건을 설정하세요.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 메모리 요구사항 정보 */}
        <MemoryRequirementInfo 
          participantCount={participantCount}
          safetyFactor={1.5}
        />

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
            max={300000}
            min={10000}
            step={5000}
            className="w-full"
          />
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>1만원</span>
            <span>30만원</span>
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

        <Button
          onClick={onOptimize}
          disabled={!canOptimize}
          className="w-full"
        >
          {isLoading ? "최적화 중..." : "집계자 배치 최적화"}
        </Button>

        {creationStatus?.step === "selecting" && (
          <Button
            onClick={onOptimize}
            variant="outline"
            className="w-full"
          >
            다시 최적화하기
          </Button>
        )}
      </CardContent>
    </Card>
  );
};