// utils/modelMemoryCalculator.ts

/**
 * 모델 정의 측정 결과(파라미터 수 기반)로부터
 * 최소 메모리 요구사항을 계산하는 유틸.
 *
 * 기존: 파일 크기(bytes) 기반 → 폐지(호환 래퍼 유지)
 * 신규: ModelMeta.fp32_bytes / fp16_bytes / int8_bytes 중 선택
 */

export type Precision = "fp32" | "fp16" | "int8";

export interface ModelMeta {
  params: number;
  buffers: number;
  total_elements: number;
  fp32_bytes: number;
  fp16_bytes: number;
  int8_bytes: number;
}

/** Bytes → GB */
export const bytesToGB = (bytes: number): number => bytes / (1024 ** 3);
/** Bytes → 사람이 읽기 쉬운 단위 */
export const formatFileSize = (bytes: number): string => {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const val = bytes / Math.pow(k, i);
  return `${Math.round(val * 100) / 100} ${sizes[i]}`;
};

/** precision에 따른 모델 상태 크기(bytes) 선택 */
const pickBytesByPrecision = (meta: ModelMeta, precision: Precision): number => {
  switch (precision) {
    case "fp16": return meta.fp16_bytes;
    case "int8": return meta.int8_bytes;
    case "fp32":
    default:     return meta.fp32_bytes;
  }
};

/**
 * 신규 API
 * 모델 정의 측정 메타 + 참여자 수 기반 최소 RAM 요구량(GB) 계산
 * 공식: ceil( (model_state_bytes * participantCount * safetyFactor) / 1GB )
 * - 기본 최소값 2GB 보장(시스템/오버헤드)
 */
export const calculateMinimumMemoryRequirementFromMeta = (
  meta: ModelMeta,
  participantCount: number,
  safetyFactor: number = 1.5,
  precision: Precision = "fp32"
): number => {
  const perModelBytes = pickBytesByPrecision(meta, precision);
  const totalBytes = perModelBytes * Math.max(1, participantCount) * safetyFactor;
  const gb = bytesToGB(totalBytes);
  const rounded = Math.round(gb * 100) / 100; // 소수점 둘째자리
  const MIN_SYSTEM_GB = 2;
  return Math.max(rounded, MIN_SYSTEM_GB);
};

/** 결과 상세 정보 타입 */
export interface MemoryRequirementDetails {
  precision: Precision;
  participantCount: number;
  safetyFactor: number;
  modelSizeBytes: number;
  modelSizeFormatted: string;
  calculatedMemoryGB: number;
  recommendedMemoryGB: number; // 일반 메모리 단계로 올림
  formula: string;
  notes?: string;
}

/** 일반 메모리 단계(필요시 조정) */
const COMMON_MEMORY_SIZES = [2, 4, 8, 16, 32, 64, 128, 256];

/**
 * 신규 API (상세)
 * - 선택 precision 기준 모델 상태 크기와 계산식을 함께 반환
 */
export const calculateMemoryRequirementDetailsFromMeta = (
  meta: ModelMeta,
  participantCount: number,
  safetyFactor: number = 1.5,
  precision: Precision = "fp32"
): MemoryRequirementDetails => {
  const modelBytes = pickBytesByPrecision(meta, precision);
  const calculatedMemoryGB = calculateMinimumMemoryRequirementFromMeta(
    meta, participantCount, safetyFactor, precision
  );
  const recommendedMemoryGB =
    COMMON_MEMORY_SIZES.find(sz => sz >= calculatedMemoryGB) ?? Math.ceil(calculatedMemoryGB);

  return {
    precision,
    participantCount,
    safetyFactor,
    modelSizeBytes: modelBytes,
    modelSizeFormatted: formatFileSize(modelBytes),
    calculatedMemoryGB,
    recommendedMemoryGB,
    formula: `ceil( ${formatFileSize(modelBytes)} × ${participantCount} × ${safetyFactor} )`,
    notes: precision === "fp32"
      ? "기본값은 FP32 기준입니다. 통신/저장 최적화를 위해 FP16/INT8도 참고하세요."
      : "혼합정밀/양자화 환경 가정. 프레임워크/런타임 설정에 따라 실제 값이 달라질 수 있습니다.",
  };
};

/* ------------------------------ 호환 래퍼 ------------------------------ */
/**
 * (구) 파일 크기 기반 API 유지 (내부적으로 FP32 메타로 변환하여 재사용)
 * 기존 코드가 크게 안 바뀌도록 남겨둡니다.
 */
export const calculateMinimumMemoryRequirement = (
  modelFileSizeInBytes: number,
  participantCount: number,
  safetyFactor: number = 1.5
): number => {
  const pseudoMeta: ModelMeta = {
    params: 0, buffers: 0, total_elements: 0,
    fp32_bytes: modelFileSizeInBytes,
    fp16_bytes: Math.floor(modelFileSizeInBytes / 2),
    int8_bytes: Math.floor(modelFileSizeInBytes / 4),
  };
  return calculateMinimumMemoryRequirementFromMeta(
    pseudoMeta, participantCount, safetyFactor, "fp32"
  );
};

/**
 * (구) 상세 API 유지
 */
export const calculateMemoryRequirementDetails = (
  modelFileSizeInBytes: number,
  participantCount: number,
  safetyFactor: number = 1.5
): MemoryRequirementDetails => {
  const pseudoMeta: ModelMeta = {
    params: 0, buffers: 0, total_elements: 0,
    fp32_bytes: modelFileSizeInBytes,
    fp16_bytes: Math.floor(modelFileSizeInBytes / 2),
    int8_bytes: Math.floor(modelFileSizeInBytes / 4),
  };
  return calculateMemoryRequirementDetailsFromMeta(
    pseudoMeta, participantCount, safetyFactor, "fp32"
  );
};
