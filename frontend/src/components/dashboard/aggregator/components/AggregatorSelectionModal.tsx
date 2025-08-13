// components/AggregatorSelectionModal.tsx
import { useState } from "react";
import { AggregatorSelectionModalProps, AggregatorOption } from "../aggregator.types";

export const AggregatorSelectionModal = ({ 
  results, 
  onSelect, 
  onCancel 
}: AggregatorSelectionModalProps) => {
  const [selectedOption, setSelectedOption] = useState<AggregatorOption | null>(null);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-6xl max-h-[80vh] overflow-y-auto">
        <h2 className="text-2xl font-bold mb-4">집계자 선택</h2>
        
        {/* 요약 정보 */}
        <div className="mb-6 p-4 bg-gray-100 rounded-lg">
          <h3 className="font-semibold mb-2">최적화 요약</h3>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>참여자 수: {results.summary.totalParticipants}명</div>
            <div>참여자 지역: {results.summary.participantRegions.join(', ')}</div>
            <div>후보 옵션: {results.summary.totalCandidateOptions}개</div>
            <div>조건 만족 옵션: {results.summary.feasibleOptions}개</div>
          </div>
        </div>

        {/* 옵션 리스트 */}
        <div className="space-y-3 mb-6">
          {results.optimizedOptions.map((option) => (
            <div
              key={`${option.region}-${option.instanceType}`}
              className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                selectedOption?.rank === option.rank
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
              }`}
              onClick={() => setSelectedOption(option)}
            >
              <div className="flex justify-between items-start mb-2">
                <div className="flex items-center space-x-2">
                  <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-sm font-medium">
                    #{option.rank}
                  </span>
                  <span className="font-semibold text-lg">
                    {option.cloudProvider} {option.region}
                  </span>
                  <span className="bg-green-100 text-green-800 px-2 py-1 rounded text-sm">
                    추천도: {option.recommendationScore}%
                  </span>
                </div>
                <div className="text-right">
                  <div className="text-2xl font-bold text-blue-600">
                    ₩{option.estimatedMonthlyCost.toLocaleString()}
                  </div>
                  <div className="text-sm text-gray-500">월 예상 비용</div>
                </div>
              </div>

              <div className="grid grid-cols-4 gap-4 mt-3">
                <div>
                  <div className="text-sm text-gray-600">인스턴스</div>
                  <div className="font-medium">{option.instanceType}</div>
                  <div className="text-xs text-gray-500">
                    {option.vcpu}vCPU, {option.memory}GB
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">평균 지연시간</div>
                  <div className="font-medium text-orange-600">{option.avgLatency}ms</div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">최대 지연시간</div>
                  <div className="font-medium text-red-600">{option.maxLatency}ms</div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">시간당 비용</div>
                  <div className="font-medium">${option.estimatedHourlyPrice}</div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* 버튼 */}
        <div className="flex justify-end space-x-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
          >
            취소
          </button>
          <button
            onClick={() => selectedOption && onSelect(selectedOption)}
            disabled={!selectedOption}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
          >
            선택한 집계자 생성하기
          </button>
        </div>
      </div>
    </div>
  );
};