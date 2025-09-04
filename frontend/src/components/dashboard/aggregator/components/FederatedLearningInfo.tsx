// components/FederatedLearningInfo.tsx
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FederatedLearningData } from "../aggregator.types";

interface FederatedLearningInfoProps {
  data: FederatedLearningData;
}

export const FederatedLearningInfo = ({ data }: FederatedLearningInfoProps) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>연합학습 정보 요약</CardTitle>
        <CardDescription>
          이전 단계에서 설정한 연합학습 정보를 확인하세요.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">이름:</div>
          <div className="text-sm col-span-2">{data.name}</div>
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">설명:</div>
          <div className="text-sm col-span-2">{data.description || "-"}</div>
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">모델 유형:</div>
          <div className="text-sm col-span-2">{data.model_type}</div>
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">알고리즘:</div>
          <div className="text-sm col-span-2">{data.algorithm}</div>
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">라운드 수:</div>
          <div className="text-sm col-span-2">{data.rounds}</div>
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="text-sm font-medium">참여자:</div>
          <div className="text-sm col-span-2">
            {data.participants?.length}명
          </div>
        </div>
        {data.modelFileName && (
          <div className="grid grid-cols-3 gap-2">
            <div className="text-sm font-medium">모델 파일:</div>
            <div className="text-sm col-span-2">{data.modelFileName}</div>
          </div>
        )}

        {/* 참여자 목록 */}
        <div className="space-y-2">
          <div className="text-sm font-medium">참여자 목록:</div>
          <div className="space-y-1">
            {data.participants?.map((participant) => (
              <div
                key={participant.id}
                className="flex items-center justify-between p-2 bg-gray-50 rounded"
              >
                <span className="text-sm">{participant.name}</span>
                <Badge
                  variant={
                    participant.status === "active" ? "default" : "secondary"
                  }
                >
                  {participant.status === "active" ? "활성" : "비활성"}
                </Badge>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
