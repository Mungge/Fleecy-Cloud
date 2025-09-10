// components/dashboard/federated-learning/constants.ts

// 집계 알고리즘 목록
export const AGGREGATION_ALGORITHMS = [
    { id: "fedavg", name: "FedAvg" },
    { id: "fedprox", name: "FedProx" },
    { id: "scaffold", name: "SCAFFOLD" },
    { id: "fedopt", name: "FedOpt" },
  ];
  
  // 지원하는 모델 유형
  export const MODEL_TYPES = [
    { id: "image_classification", name: "이미지 분류" },
    { id: "nlp", name: "자연어 처리" },
    { id: "tabular", name: "테이블 형식 데이터" },
  ];
  
 export const SELECTION_STRATEGIES = [
    { id: "accuracy", name: "정확도" },
    { id: "f1-score", name: "F1-score" },
    { id: "precision", name: "정밀도" },
    { id: "recall", name: "재현율" },
  ]; 
  // 파일 형식
  export const SUPPORTED_FILE_FORMATS = ".py";
  
  // 폼 기본값
  export const FORM_DEFAULTS = {
    ALGORITHM: "fedavg",
    ROUNDS: 10,
  } as const;