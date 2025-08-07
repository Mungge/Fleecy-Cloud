// components/CreationStatus.tsx
import { Card, CardContent } from "@/components/ui/card";
import { CreationStatusDisplayProps } from "../aggregator.types";

export const CreationStatusDisplay = ({ status }: CreationStatusDisplayProps) => {
  if (!status) return null;

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium">배포 진행 상황</h3>
            <span className="text-sm text-gray-500">
              {status.progress}%
            </span>
          </div>

          {/* Progress Bar */}
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all duration-500 ${
                status.step === "error"
                  ? "bg-red-500"
                  : "bg-blue-500"
              }`}
              style={{ width: `${status.progress || 0}%` }}
            ></div>
          </div>

          <p
            className={`text-sm ${
              status.step === "error"
                ? "text-red-600"
                : "text-gray-600"
            }`}
          >
            {status.message}
          </p>
        </div>
      </CardContent>
    </Card>
  );
};