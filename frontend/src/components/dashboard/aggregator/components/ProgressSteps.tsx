// components/ProgressSteps.tsx
import { Check } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { ProgressStepsProps } from "../aggregator.types";

export const ProgressSteps = ({ creationStatus, isLoading }: ProgressStepsProps) => {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="w-full py-4">
          <div className="flex items-center justify-between max-w-2xl mx-auto">
            {/* Step 1: 정보 입력 (완료) */}
            <div className="flex flex-col items-center">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-green-500 text-white text-lg font-medium shadow-lg">
                <Check className="w-6 h-6" />
              </div>
              <span className="mt-3 text-base font-medium text-green-600">
                정보 입력
              </span>
              <span className="mt-1 text-sm text-gray-500">
                연합학습 정보 설정
              </span>
            </div>

            {/* Connector Line (완료) */}
            <div className="flex-1 h-1 bg-green-500 mx-6 rounded-full"></div>

            {/* Step 2: 집계자 생성 (현재/완료) */}
            <div className="flex flex-col items-center">
              <div
                className={`flex items-center justify-center w-12 h-12 rounded-full text-white text-lg font-medium shadow-lg ${
                  creationStatus?.step === "completed"
                    ? "bg-green-500"
                    : "bg-blue-500"
                }`}
              >
                {creationStatus?.step === "completed" ? (
                  <Check className="w-6 h-6" />
                ) : isLoading ? (
                  <div className="animate-spin rounded-full h-6 w-6 border-2 border-white border-t-transparent"></div>
                ) : (
                  "2"
                )}
              </div>
              <span
                className={`mt-3 text-base font-medium ${
                  creationStatus?.step === "completed"
                    ? "text-green-600"
                    : "text-blue-600"
                }`}
              >
                집계자 생성
              </span>
              <span className="mt-1 text-sm text-gray-500">집계자 설정</span>
            </div>

            {/* Connector Line */}
            <div
              className={`flex-1 h-1 mx-6 rounded-full ${
                creationStatus?.step === "completed"
                  ? "bg-green-500"
                  : "bg-gray-200"
              }`}
            ></div>

            {/* Step 3: 연합학습 생성 */}
            <div className="flex flex-col items-center">
              <div
                className={`flex items-center justify-center w-12 h-12 rounded-full text-lg font-medium ${
                  creationStatus?.step === "completed"
                    ? "bg-green-500 text-white shadow-lg"
                    : "bg-gray-200 text-gray-400"
                }`}
              >
                {creationStatus?.step === "completed" ? (
                  <Check className="w-6 h-6" />
                ) : (
                  "3"
                )}
              </div>
              <span
                className={`mt-3 text-base ${
                  creationStatus?.step === "completed"
                    ? "text-green-600 font-medium"
                    : "text-gray-400"
                }`}
              >
                연합학습 생성
              </span>
              <span className="mt-1 text-sm text-gray-400">
                최종 생성 완료
              </span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};