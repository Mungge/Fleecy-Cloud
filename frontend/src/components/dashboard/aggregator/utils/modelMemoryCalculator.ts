// utils/modelMemoryCalculator.ts

/**
 * 모델 파일 크기와 참여자 수를 기반으로 최소 메모리 요구사항을 계산합니다.
 * 
 * @param modelFileSizeInBytes - 모델 파일 크기 (bytes)
 * @param participantCount - 참여자 수
 * @param safetyFactor - 안전 계수 (기본값: 1.5)
 * @returns 최소 메모리 요구사항 (GB)
 */
export const calculateMinimumMemoryRequirement = (
    modelFileSizeInBytes: number,
    participantCount: number,
    safetyFactor: number = 1.5
  ): number => {
    // bytes to GB 변환 (1GB = 1024^3 bytes)
    const modelFileSizeInGB = modelFileSizeInBytes / (1024 * 1024 * 1024);
    
    // 최소 메모리 요구사항 계산
    // 공식: 모델 크기 * 참여자 수 * 안전 계수
    const minMemoryGB = modelFileSizeInGB * participantCount * safetyFactor;
    
    // 최소 2GB는 보장 (시스템 오버헤드용)
    const minSystemMemory = 2;
    
    // 소수점 둘째자리까지 반올림
    return Math.max(Math.round(minMemoryGB * 100) / 100, minSystemMemory);
  };
  
  /**
   * 파일 크기를 읽기 쉬운 형식으로 변환합니다.
   * 
   * @param bytes - 바이트 단위 파일 크기
   * @returns 포맷된 파일 크기 문자열
   */
  export const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };
  
  /**
   * 메모리 요구사항 계산 결과를 포함한 상세 정보를 반환합니다.
   */
  export interface MemoryRequirementDetails {
    modelFileSizeBytes: number;
    modelFileSizeFormatted: string;
    participantCount: number;
    safetyFactor: number;
    calculatedMemoryGB: number;
    recommendedMemoryGB: number; // 가장 가까운 일반적인 메모리 크기로 반올림
    formula: string;
  }
  
  /**
   * 메모리 요구사항에 대한 상세 정보를 계산합니다.
   */
  export const calculateMemoryRequirementDetails = (
    modelFileSizeInBytes: number,
    participantCount: number,
    safetyFactor: number = 1.5
  ): MemoryRequirementDetails => {
    const calculatedMemoryGB = calculateMinimumMemoryRequirement(
      modelFileSizeInBytes,
      participantCount,
      safetyFactor
    );
    
    // 일반적인 메모리 크기로 반올림 (2, 4, 8, 16, 32, 64, 128, 256 GB)
    const commonMemorySizes = [2, 4, 8, 16, 32, 64, 128, 256];
    const recommendedMemoryGB = commonMemorySizes.find(size => size >= calculatedMemoryGB) || calculatedMemoryGB;
    
    return {
      modelFileSizeBytes: modelFileSizeInBytes,
      modelFileSizeFormatted: formatFileSize(modelFileSizeInBytes),
      participantCount,
      safetyFactor,
      calculatedMemoryGB,
      recommendedMemoryGB,
      formula: `(${formatFileSize(modelFileSizeInBytes)} × ${participantCount}명 × ${safetyFactor}) = ${calculatedMemoryGB.toFixed(2)}GB → ${recommendedMemoryGB}GB 권장`
    };
  };